package kvptorestful

import (
	"github.com/pdok/mapproxy-operator/internal/controller/constants"
	"github.com/pdok/mapproxy-operator/internal/controller/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func GetKvpToRestfulContainer(images *types.Images) (*corev1.Container, error) {
	probe := getProbe()

	result := corev1.Container{
		Name:    constants.KvpToRestfulName,
		Image:   images.KvpToRestfulImage,
		Command: []string{"wmts-kvp-to-restful"},
		Args:    []string{"-host=http://127.0.0.1:80", "-t=/var/www/WMTSCapabilities.xml", "-d=15"},
		Ports: []corev1.ContainerPort{{
			ContainerPort: 9001,
			Protocol:      corev1.ProtocolTCP,
		}},
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("128M"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("0.1"),
			},
		},
		VolumeMounts: []corev1.VolumeMount{{
			Name:      "data",
			MountPath: "/var/www",
		}},
		LivenessProbe:   probe,
		ReadinessProbe:  probe,
		ImagePullPolicy: corev1.PullIfNotPresent,
	}

	return &result, nil
}

func getProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/health",
				Port: intstr.IntOrString{Type: intstr.Int, IntVal: 9001},
			},
		},
		InitialDelaySeconds: 5,
		TimeoutSeconds:      20,
		PeriodSeconds:       10,
	}
}
