package controller

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo bdd
	. "github.com/onsi/gomega"    //nolint:revive // ginkgo bdd
	pdoknlv2 "github.com/pdok/mapproxy-operator/api/v2"
	"github.com/pdok/mapproxy-operator/internal/controller/types"
	"github.com/pdok/smooth-operator/model"
	smoothoperatorutils "github.com/pdok/smooth-operator/pkg/util"
	"github.com/pdok/smooth-operator/pkg/validation"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	namespace      = "default"
	testImageName1 = "test.test/image:test1"
	testImageName2 = "test.test/image:test2"
	testImageName3 = "test.test/image:test3"
	testImageName4 = "test.test/image:test4"
	testImageName5 = "test.test/image:test5"
)

var _ = Describe("Testing WMTS Controller", func() {

	Context("Testing Mutate functions for Cache WMTS", func() {
		testMutates(getWMTSReconcilerPtr, &pdoknlv2.WMTS{}, "cache")
	})

	Context("Testing Mutate functions for WMTS with featureinfo", func() {
		testMutates(getWMTSReconcilerPtr, &pdoknlv2.WMTS{}, "featureinfo")
	})

	Context("Testing Mutate functions for WMTS without cache", func() {
		testMutates(getWMTSReconcilerPtr, &pdoknlv2.WMTS{}, "nocache")
	})

	Context("When reconciling a resource", func() {

		ctx := context.Background()

		inputPath := testPath("cache") + "input/"

		testWMTS := pdoknlv2.WMTS{}
		clusterWMTS := &pdoknlv2.WMTS{}

		objectKeyWMTS := k8stypes.NamespacedName{}

		var expectedResources []struct {
			obj client.Object
			key k8stypes.NamespacedName
		}

		It("Should create a WMTS resource on the cluster", func() {

			By("Creating a new resource for the Kind WMTS")
			data, err := readTestFile(inputPath + "wmts.yaml")
			Expect(err).NotTo(HaveOccurred())
			err = yaml.UnmarshalStrict(data, &testWMTS)
			Expect(err).NotTo(HaveOccurred())
			Expect(testWMTS.Name).Should(Equal("cache"))

			objectKeyWMTS = k8stypes.NamespacedName{
				Namespace: testWMTS.GetNamespace(),
				Name:      testWMTS.GetName(),
			}

			err = k8sClient.Get(ctx, objectKeyWMTS, clusterWMTS)
			Expect(client.IgnoreNotFound(err)).To(Not(HaveOccurred()))
			if err != nil && apierrors.IsNotFound(err) {
				resource := testWMTS.DeepCopy()
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
				Expect(k8sClient.Get(ctx, objectKeyWMTS, clusterWMTS)).To(Succeed())
			}

		})

		It("Should reconcile successfully", func() {
			controllerReconciler := getWMTSReconciler()

			By("Reconciling the WMS")
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: objectKeyWMTS})
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should create all expected resources", func() {
			expectedResources, err := getExpectedObjects(ctx, clusterWMTS)
			Expect(err).NotTo(HaveOccurred())

			for _, expectedResource := range expectedResources {
				Eventually(func() bool {
					err := k8sClient.Get(ctx, k8stypes.NamespacedName{Namespace: expectedResource.GetNamespace(), Name: expectedResource.GetName()}, expectedResource)
					return Expect(err).NotTo(HaveOccurred())
				}, "10s", "1s").Should(BeTrue())
			}
		})

		It("Should successfully reconcile after a change in an owned resource", func() {
			controllerReconciler := getWMTSReconciler()

			By("Getting the original Deployment")
			deployment := getBareDeployment(clusterWMTS)
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
				return Expect(err).NotTo(HaveOccurred())
			}, "10s", "1s").Should(BeTrue())
			originalRevisionHistoryLimit := *deployment.Spec.RevisionHistoryLimit
			expectedRevisionHistoryLimit := 99
			Expect(originalRevisionHistoryLimit).Should(Not(Equal(expectedRevisionHistoryLimit)))

			By("Altering the Deployment")
			err := k8sClient.Patch(ctx, deployment, client.RawPatch(k8stypes.MergePatchType, []byte(
				fmt.Sprintf(`{"spec": {"revisionHistoryLimit": %d}}`, expectedRevisionHistoryLimit))))
			Expect(err).NotTo(HaveOccurred())

			By("Verifying that the Deployment was altered")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
				return Expect(err).NotTo(HaveOccurred()) &&
					Expect(*deployment.Spec.RevisionHistoryLimit).To(BeEquivalentTo(expectedRevisionHistoryLimit))
			}, "10s", "1s").Should(BeTrue())

			By("Reconciling the WMS again")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: objectKeyWMTS})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying that the Deployment was restored")
			Eventually(func() bool {
				err = k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
				return Expect(err).NotTo(HaveOccurred()) &&
					Expect(*deployment.Spec.RevisionHistoryLimit).To(BeEquivalentTo(originalRevisionHistoryLimit))
			}, "10s", "1s").Should(BeTrue())
		})

		It("Respects the TTL of the WMS", func() {
			By("Creating a new resource for the Kind WMS")

			ttlName := testWMTS.GetName() + "-ttl"
			ttlWms := testWMTS.DeepCopy()
			ttlWms.Name = ttlName
			ttlWms.Spec.Lifecycle = &model.Lifecycle{TTLInDays: smoothoperatorutils.Pointer(int32(0))}
			objectKeyTTLWMS := client.ObjectKeyFromObject(ttlWms)

			err := k8sClient.Get(ctx, objectKeyTTLWMS, ttlWms)
			Expect(client.IgnoreNotFound(err)).To(Not(HaveOccurred()))
			if err != nil && apierrors.IsNotFound(err) {
				Expect(k8sClient.Create(ctx, ttlWms)).To(Succeed())
			}

			// Reconcile
			reconciler := getWMTSReconciler()
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: objectKeyTTLWMS})
			Expect(err).To(Not(HaveOccurred()))

			// Check the WMS cannot be found anymore
			Eventually(func() bool {
				err = k8sClient.Get(ctx, objectKeyTTLWMS, ttlWms)
				return apierrors.IsNotFound(err)
			}, "10s", "1s").Should(BeTrue())

			// Not checking owned resources because the test env does not do garbage collection
		})

		It("Should cleanup the cluster", func() {
			err := k8sClient.Get(ctx, objectKeyWMTS, clusterWMTS)
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance WMS")
			Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, clusterWMTS))).To(Succeed())

			// the testEnv does not do garbage collection (https://book.kubebuilder.io/reference/envtest#testing-considerations)
			By("Cleaning Owned Resources")
			for _, d := range expectedResources {
				err := k8sClient.Get(ctx, d.key, d.obj)
				Expect(err).NotTo(HaveOccurred())
				Expect(k8sClient.Delete(ctx, d.obj)).To(Succeed())
			}
		})
	})

	Context("When manually validating an incoming CRD", func() {
		It("Should not error", func() {
			err := validation.LoadSchemasForCRD(cfg, "default", "wmts.pdok.nl")
			Expect(err).NotTo(HaveOccurred())

			filepath := "input/wmts.yaml"
			testCases := []string{
				testPath("cache") + filepath,
				testPath("featureinfo") + filepath,
				testPath("nocache") + filepath,
			}

			for _, test := range testCases {
				yamlInput, err := readTestFile(test)
				Expect(err).NotTo(HaveOccurred())

				err = validation.ValidateSchema(string(yamlInput))
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})
})

func getWMTSReconciler() WMTSReconciler {
	return WMTSReconciler{
		Client: k8sClient,
		Scheme: k8sClient.Scheme(),
		Images: types.Images{
			ApacheExporterImage:        testImageName1,
			CapabilitiesGeneratorImage: testImageName2,
			KvpToRestfulImage:          testImageName3,
			MapproxyImage:              testImageName4,
			MultiToolImage:             testImageName5,
		},
	}
}

// function to defer construction of the reconciler to inside the test to ensure test injections
func getWMTSReconcilerPtr() *WMTSReconciler {
	result := getWMTSReconciler()
	return &result
}
