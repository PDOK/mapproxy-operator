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
	"sigs.k8s.io/yaml"
)

var storageClassName string //nolint:unused

func SetStorageClassName(name string) {
	storageClassName = name
}

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
	b, _ := yaml.Marshal(initContainers)
	println("Init containers:")
	println(string(b))

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
	b, _ = yaml.Marshal(volumes)
	println("Volumes:")
	println(string(b))

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
	//nolint:gocritic
	//baseVolume := corev1.Volume{Name: constants.BaseVolumeName}
	//if use, size := mapperutils.UseEphemeralVolume(obj); use {
	//	baseVolume.Ephemeral = &corev1.EphemeralVolumeSource{
	//		VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{
	//			Spec: corev1.PersistentVolumeClaimSpec{
	//				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
	//				Resources: corev1.VolumeResourceRequirements{Requests: corev1.ResourceList{
	//					corev1.ResourceStorage: *size,
	//				}},
	//			},
	//		},
	//	}
	//	if storageClassName != "" {
	//		baseVolume.Ephemeral.VolumeClaimTemplate.Spec.StorageClassName = &storageClassName
	//	}
	//} else {
	//	baseVolume.EmptyDir = &corev1.EmptyDirVolumeSource{}
	//}
	//
	//volumes := []corev1.Volume{
	//	baseVolume,
	//	{Name: constants.DataVolumeName, VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
	//	getConfigMapVolume(constants.MapserverName, configMapNames.Mapserver),
	//}
	//
	//if mapfile := obj.Mapfile(); mapfile != nil {
	//	volumes = append(volumes, getConfigMapVolume(constants.ConfigMapCustomMapfileVolumeName, mapfile.ConfigMapKeyRef.Name))
	//}
	//
	//if obj.Type() == pdoknlv3.ServiceTypeWMS && obj.Options().UseWebserviceProxy() {
	//	volumes = append(volumes, getConfigMapVolume(constants.ConfigMapOgcWebserviceProxyVolumeName, configMapNames.OgcWebserviceProxy))
	//}
	//
	//if obj.Options().PrefetchData {
	//	vol := getConfigMapVolume(constants.InitScriptsName, configMapNames.InitScripts)
	//	vol.ConfigMap.DefaultMode = smoothoperatorutils.Pointer(int32(0777))
	//	volumes = append(volumes, vol)
	//}
	//
	//volumes = append(volumes, getConfigMapVolume(constants.ConfigMapCapabilitiesGeneratorVolumeName, configMapNames.CapabilitiesGenerator))
	//
	//if obj.Mapfile() == nil {
	//	volumes = append(volumes, getConfigMapVolume(constants.ConfigMapMapfileGeneratorVolumeName, configMapNames.MapfileGenerator))
	//}
	//
	//if obj.Type() == pdoknlv3.ServiceTypeWMS {
	//	if obj.Mapfile() == nil {
	//		wms, _ := any(obj).(*pdoknlv3.WMS)
	//		volumeProjections := []corev1.VolumeProjection{}
	//		for _, cm := range wms.Spec.Service.StylingAssets.ConfigMapRefs {
	//			volumeProjections = append(volumeProjections, corev1.VolumeProjection{
	//				ConfigMap: &corev1.ConfigMapProjection{LocalObjectReference: corev1.LocalObjectReference{Name: cm.Name}},
	//			})
	//		}
	//
	//		volumes = append(volumes, corev1.Volume{
	//			Name:         constants.ConfigMapStylingFilesVolumeName,
	//			VolumeSource: corev1.VolumeSource{Projected: &corev1.ProjectedVolumeSource{Sources: volumeProjections}},
	//		})
	//	} else {
	//		// If there is a custom mapfile, we still want a styling-files volume, even if it is empty
	//		volumes = append(volumes, corev1.Volume{
	//			Name:         constants.ConfigMapStylingFilesVolumeName,
	//			VolumeSource: corev1.VolumeSource{Projected: &corev1.ProjectedVolumeSource{Sources: []corev1.VolumeProjection{}}},
	//		})
	//	}
	//
	//	volumes = append(
	//		volumes,
	//		getConfigMapVolume(constants.ConfigMapFeatureinfoGeneratorVolumeName, configMapNames.FeatureInfoGenerator),
	//		getConfigMapVolume(constants.ConfigMapLegendGeneratorVolumeName, configMapNames.LegendGenerator),
	//	)
	//}
	//
	//return volumes
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
