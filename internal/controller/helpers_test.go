package controller

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo bdd
	. "github.com/onsi/gomega"    //nolint:revive // ginkgo bdd
	pdoknlv2 "github.com/pdok/mapproxy-operator/api/v2"
	"github.com/pdok/mapproxy-operator/internal/controller/constants"
	"github.com/pdok/mapproxy-operator/internal/controller/types"
	"github.com/pdok/smooth-operator/pkg/validation"
	traefikiov1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	yaml2 "gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	v2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func testMutates(reconcilerFn func() *WMTSReconciler, resource *pdoknlv2.WMTS, name string, ignoreFiles ...string) {
	inputPath := testPath(name) + "input/"
	outputPath := testPath(name) + "expected/"

	var fileName = "wmts.yaml"

	It("Should parse the input files correctly", func() {
		data, err := readTestFile(inputPath + fileName)
		Expect(err).NotTo(HaveOccurred())
		err = yaml.UnmarshalStrict(data, resource)
		Expect(err).NotTo(HaveOccurred())
		Expect(resource.GetName()).Should(Equal(name))

		var validationError error
		_, validationError = resource.ValidateCreate(k8sClient)
		Expect(validationError).NotTo(HaveOccurred())
	})

	configMapNames := types.HashedConfigMapNames{}

	//It("Should generate a correct Configmap", func() {
	//	cm := getBareConfigMap(resource, constants.MapserverName)
	//	testMutateConfigMap(cm, outputPath+"configmap-mapserver.yaml", func(cm *corev1.ConfigMap) error {
	//		return mutateConfigMap(reconciler, resource, cm)
	//	}, true)
	//	configMapNames.Mapserver = cm.Name
	//})
	//
	//It("Should generate a correct BlobDownload Configmap", func() {
	//	if path, include := shouldIncludeFile("configmap-init-scripts.yaml"); include {
	//		cm := getBareConfigMap(resource, constants.InitScriptsName)
	//		testMutateConfigMap(cm, path, func(cm *corev1.ConfigMap) error {
	//			return mutateConfigMapBlobDownload(reconciler, resource, cm)
	//		}, true)
	//		configMapNames.InitScripts = cm.Name
	//	}
	//})
	//
	//It("Should generate a correct MapfileGenerator Configmap", func() {
	//	if path, include := shouldIncludeFile("configmap-mapfile-generator.yaml"); include {
	//		cm := getBareConfigMap(resource, constants.MapfileGeneratorName)
	//		testMutateConfigMap(cm, path, func(cm *corev1.ConfigMap) error {
	//			return mutateConfigMapMapfileGenerator(reconciler, resource, cm)
	//		}, true)
	//		configMapNames.MapfileGenerator = cm.Name
	//	}
	//})

	It("Should generate a correct CapabilitiesGenerator Configmap", func() {
		cm := getBareConfigMap(resource, constants.CapabilitiesGeneratorName)
		testMutateConfigMap(cm, outputPath+"configmap-capabilities-generator.yaml", func(cm *corev1.ConfigMap) error {
			return mutateConfigMapCapabilitiesGenerator(reconcilerFn(), resource, cm)
		}, true)
		configMapNames.CapabilitiesGenerator = cm.Name
	})

	It("Should generate a Deployment correctly", func() {
		testMutate("Deployment", getBareDeployment(resource), outputPath+"deployment.yaml", func(d *appsv1.Deployment) error {
			return mutateDeployment(reconcilerFn(), resource, d, configMapNames)
		})
	})

	It("Should generate a correct Service", func() {
		testMutate("Service", getBareService(resource), outputPath+"service.yaml", func(s *corev1.Service) error {
			return mutateService(reconcilerFn(), resource, s)
		})
	})

	It("Should generate a correct Headers Middleware", func() {
		testMutate("Headers Middleware", getBareCorsHeadersMiddleware(resource), outputPath+"middleware-headers.yaml", func(m *traefikiov1alpha1.Middleware) error {
			return mutateCorsHeadersMiddleware(reconcilerFn(), resource, m)
		})
	})

	It("Should generate a correct IngressRoute", func() {
		testMutate("IngressRoute", getBareIngressRoute(resource, ""), outputPath+"ingressroute.yaml", func(i *traefikiov1alpha1.IngressRoute) error {
			return mutateDirectIngressRoute(reconcilerFn(), resource, i)
		})
	})

	It("Should generate a correct IngressRoute", func() {
		testMutate("IngressRoute", getBareIngressRoute(resource, "-restful"), outputPath+"ingressroute-restful.yaml", func(i *traefikiov1alpha1.IngressRoute) error {
			return mutateRestfulIngressRoute(reconcilerFn(), resource, i)
		})
	})

	It("Should generate a correct PodDisruptionBudget", func() {
		testMutate("PodDisruptionBudget", getBarePodDisruptionBudget(resource), outputPath+"poddisruptionbudget.yaml", func(p *policyv1.PodDisruptionBudget) error {
			return mutatePodDisruptionBudget(reconcilerFn(), resource, p)
		})
	})

	It("Should generate a correct HorizontalPodAutoscaler", func() {
		testMutate("PodDisruptionBudget", getBareHorizontalPodAutoScaler(resource), outputPath+"horizontalpodautoscaler.yaml", func(h *v2.HorizontalPodAutoscaler) error {
			return mutateHorizontalPodAutoscaler(reconcilerFn(), resource, h)
		})
	})
}

func readTestFile(fileName string) ([]byte, error) {
	dat, err := os.ReadFile(fileName)
	if err != nil {
		return []byte{}, err
	}

	// Apply defaults
	un := unstructured.Unstructured{}
	err = yaml.Unmarshal(dat, &un)
	defaulted, err := validation.ApplySchemaDefaults(un.Object)
	if err != nil {
		return []byte{}, err
	}

	return yaml2.Marshal(defaulted)
}

func testPath(test string) string {
	return fmt.Sprintf("test_data/%s/", test)
}

func testMutate[T any](kind string, result *T, expectedFile string, mutate func(*T) error) {
	By("Testing mutating the " + kind)
	err := mutate(result)
	Expect(err).NotTo(HaveOccurred())

	var expected T
	data, err := os.ReadFile(expectedFile)
	Expect(err).NotTo(HaveOccurred())
	err = yaml.UnmarshalStrict(data, &expected)
	Expect(err).NotTo(HaveOccurred())

	diff := cmp.Diff(expected, *result)
	if diff != "" {
		Fail(diff)
	}

	By(fmt.Sprintf("Testing mutating the %s twice has the same result", kind))
	generated := *result
	err = mutate(result)
	Expect(err).NotTo(HaveOccurred())
	diff = cmp.Diff(generated, *result)
	if diff != "" {
		Fail(diff)
	}
}

//nolint:unparam
func testMutateConfigMap(m *corev1.ConfigMap, expectedFile string, mutate func(*corev1.ConfigMap) error, ignoreValues bool) {
	clearConfigMapValues := func(cm *corev1.ConfigMap) {
		newMap := map[string]string{}
		for k := range cm.Data {
			newMap[k] = "IGNORED"
		}
		cm.Data = newMap
	}

	if !ignoreValues {
		testMutate("ConfigMap", m, expectedFile, mutate)
	} else {
		By("Testing mutating the ConfigMap")
		err := mutate(m)
		Expect(err).NotTo(HaveOccurred())

		expected := &corev1.ConfigMap{}
		data, err := os.ReadFile(expectedFile)
		Expect(err).NotTo(HaveOccurred())
		err = yaml.UnmarshalStrict(data, expected)
		Expect(err).NotTo(HaveOccurred())

		c := m.DeepCopy()
		clearConfigMapValues(c)
		clearConfigMapValues(expected)

		diff := cmp.Diff(*expected, *c)
		if diff != "" {
			Fail(diff)
		}
	}
}

func getExpectedObjects(ctx context.Context, obj *pdoknlv2.WMTS, includeBlobDownload bool, includeMapfileGeneratorConfigMap bool) ([]client.Object, error) {
	objects := []client.Object{
		getBareDeployment(obj),
		getBareHorizontalPodAutoScaler(obj),
		getBareService(obj),
		getBareIngressRoute(obj, ""),
		getBareIngressRoute(obj, "-restful"),
		getBareCorsHeadersMiddleware(obj),
		getBarePodDisruptionBudget(obj),
	}

	//// Add all ConfigMaps with hashed names
	//cm := getBareConfigMap(obj, constants.MapserverName)
	//hashedName, err := getHashedConfigMapNameFromClient(ctx, obj, constants.MapserverName)
	//if err != nil {
	//	return objects, err
	//}
	//cm.Name = hashedName
	//objects = append(objects, cm)
	//
	//
	//cm = getBareConfigMap(obj, constants.CapabilitiesGeneratorName)
	//hashedName, err = getHashedConfigMapNameFromClient(ctx, obj, constants.ConfigMapCapabilitiesGeneratorVolumeName)
	//if err != nil {
	//	return objects, err
	//}
	//cm.Name = hashedName
	//objects = append(objects, cm)

	return objects, nil
}
