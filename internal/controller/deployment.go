package controller

import (
	pdoknlv2 "github.com/pdok/mapproxy-operator/api/v2"
	"github.com/pdok/mapproxy-operator/internal/controller/apacheexporter"
	"github.com/pdok/mapproxy-operator/internal/controller/blobdownload"
	"github.com/pdok/mapproxy-operator/internal/controller/capabilitiesgenerator"
	"github.com/pdok/mapproxy-operator/internal/controller/constants"
	"github.com/pdok/mapproxy-operator/internal/controller/kvptorestful"
	"github.com/pdok/mapproxy-operator/internal/controller/mapperutils"
	"github.com/pdok/mapproxy-operator/internal/controller/mapproxy"
	"github.com/pdok/mapproxy-operator/internal/controller/types"
	smoothoperatorutils "github.com/pdok/smooth-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

func mutateDeployment(r *WMTSReconciler, obj *pdoknlv2.WMTS, deployment *appsv1.Deployment, configMapNames types.HashedConfigMapNames) error {
	reconcilerClient := r.Client
	labels := smoothoperatorutils.CloneOrEmptyMap(obj.GetLabels())
	if err := smoothoperatorutils.SetImmutableLabels(reconcilerClient, deployment, labels); err != nil {
		return err
	}

	deployment.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}

	deployment.Spec.RevisionHistoryLimit = smoothoperatorutils.Pointer(int32(1))
	deployment.Spec.Strategy = appsv1.DeploymentStrategy{
		Type: appsv1.RollingUpdateDeploymentStrategyType,
		RollingUpdate: &appsv1.RollingUpdateDeployment{
			MaxUnavailable: &intstr.IntOrString{IntVal: 1},
			MaxSurge:       &intstr.IntOrString{IntVal: 1},
		},
	}

	initContainers, err := getInitContainersForDeployment(r, obj)

	if err != nil {
		return err
	}
	setTerminationMessage(initContainers)

	images := r.Images
	containers, err := getContainers(obj, &images)
	if err != nil {
		return err
	}
	setTerminationMessage(containers)

	volumes := getVolumes(obj, configMapNames)

	podTemplateSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: getPodAnnotations(deployment),
			Labels:      labels,
		},
		Spec: corev1.PodSpec{
			RestartPolicy:                 corev1.RestartPolicyAlways,
			DNSPolicy:                     corev1.DNSClusterFirst,
			TerminationGracePeriodSeconds: smoothoperatorutils.Pointer(int64(60)),
			InitContainers:                initContainers,
			Containers:                    containers,
			Volumes:                       volumes,
		},
	}

	podPatch := obj.Spec.PodSpecPatch
	patchedSpec, err := smoothoperatorutils.StrategicMergePatch(&podTemplateSpec.Spec, &podPatch)
	if err != nil {
		return err
	}
	podTemplateSpec.Spec = *patchedSpec

	if use, _ := mapperutils.UseEphemeralVolume(obj); !use {
		ephStorage := podTemplateSpec.Spec.Containers[0].Resources.Limits[corev1.ResourceEphemeralStorage]
		threshold := resource.MustParse("200M")

		if ephStorage.Value() < threshold.Value() {
			podTemplateSpec.Spec.Containers[0].Resources.Limits[corev1.ResourceEphemeralStorage] = threshold
		}
	} else {
		delete(podTemplateSpec.Spec.Containers[0].Resources.Limits, corev1.ResourceEphemeralStorage)
		delete(podTemplateSpec.Spec.Containers[0].Resources.Requests, corev1.ResourceEphemeralStorage)
	}

	deployment.Spec.Template = podTemplateSpec

	if err = smoothoperatorutils.EnsureSetGVK(reconcilerClient, deployment, deployment); err != nil {
		return err
	}
	return ctrl.SetControllerReference(obj, deployment, r.Scheme)
}

func getInitContainersForDeployment(r *WMTSReconciler, obj *pdoknlv2.WMTS) ([]corev1.Container, error) { //nolint:revive
	result := []corev1.Container{}
	images := r.Images
	blobDownloadInitContainer, err := blobdownload.GetBlobDownloadInitContainer(obj, images)
	if err != nil {
		return nil, err
	}
	result = append(result, *blobDownloadInitContainer)
	capabilitiesGeneratorInitContainer, err := capabilitiesgenerator.GetCapabilitiesGeneratorInitContainer(obj, images)
	if err != nil {
		return nil, err
	}
	result = append(result, *capabilitiesGeneratorInitContainer)

	return result, nil
}

func setTerminationMessage(c []corev1.Container) {
	for i := range c {
		c[i].TerminationMessagePolicy = "File"
		c[i].TerminationMessagePath = "/dev/termination-log"
	}
}

func getContainers(obj *pdoknlv2.WMTS, images *types.Images) ([]corev1.Container, error) { //nolint:revive
	containers := []corev1.Container{}

	kvpToRestfulContainer, err := kvptorestful.GetKvpToRestfulContainer(images)
	if err != nil {
		return nil, err
	}
	containers = append(containers, *kvpToRestfulContainer)

	mapproxyContainer, err := mapproxy.GetMapproxyContainer(obj, images)
	if err != nil {
		return nil, err
	}
	containers = append(containers, *mapproxyContainer)

	apacheContainer, err := apacheexporter.GetApacheContainer(images)
	if err != nil {
		return nil, err
	}
	containers = append(containers, *apacheContainer)

	return containers, nil
}

func getVolumes(_ *pdoknlv2.WMTS, configMapNames types.HashedConfigMapNames) []corev1.Volume { //nolint:revive
	return []corev1.Volume{
		{
			Name:         "data",
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
		}, {
			Name:         constants.MapproxyVolumeName,
			VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: configMapNames.Mapproxy}}},
		}, {
			Name:         constants.LighttpdVolumeName,
			VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: configMapNames.Mapproxy}}},
		}, {
			Name:         constants.ConfigMapCapabilitiesGeneratorVolumeName,
			VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: configMapNames.CapabilitiesGenerator}}},
		},
	}
}

func getPodAnnotations(deployment *appsv1.Deployment) map[string]string {
	annotations := smoothoperatorutils.CloneOrEmptyMap(deployment.Spec.Template.GetAnnotations())
	annotations["cluster-autoscaler.kubernetes.io/safe-to-evict"] = "true"
	annotations["priority.version-checker.io/mapproxy"] = "4"
	annotations["priority.version-checker.io/wmts-kvp-to-restful"] = "4"
	annotations["prometheus.io/port"] = "9117"
	annotations["prometheus.io/scrape"] = "true"
	return annotations
}
