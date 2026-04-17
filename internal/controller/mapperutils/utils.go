package mapperutils

import (
	pdoknlv2 "github.com/pdok/mapproxy-operator/api/v2"
	"github.com/pdok/mapproxy-operator/internal/controller/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Use ephemeral volume when ephemeral storage is greater then 10Gi
func UseEphemeralVolume(obj *pdoknlv2.WMTS) (bool, *resource.Quantity) {
	value := EphemeralStorageLimit(obj)
	threshold := resource.MustParse("10Gi")

	if value != nil {
		return value.Value() > threshold.Value(), value
	}

	return false, nil
}

func EphemeralStorageLimit(obj *pdoknlv2.WMTS) *resource.Quantity {
	return GetContainerResourceLimit(obj, constants.MapproxyName, corev1.ResourceEphemeralStorage)
}
func GetContainerResourceLimit(obj *pdoknlv2.WMTS, containerName string, resource corev1.ResourceName) *resource.Quantity {
	for _, container := range obj.Spec.PodSpecPatch.Containers {
		if container.Name == containerName {
			q := container.Resources.Limits[resource]
			if !q.IsZero() {
				return &q
			}
		}
	}

	return nil
}

func GetContainerResourceRequest(obj *pdoknlv2.WMTS, containerName string, resource corev1.ResourceName) *resource.Quantity {
	for _, container := range obj.Spec.PodSpecPatch.Containers {
		if container.Name == containerName {
			q := container.Resources.Requests[resource]
			if !q.IsZero() {
				return &q
			}
		}
	}

	return nil
}
