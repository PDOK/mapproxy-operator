package apacheexporter

import (
	"github.com/pdok/mapproxy-operator/internal/controller/constants"
	"github.com/pdok/mapproxy-operator/internal/controller/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func GetApacheContainer(images *types.Images) (*corev1.Container, error) {
	result := corev1.Container{
		Name:  constants.ApacheExporterName,
		Image: images.ApacheExporterImage,
		Args:  []string{"-scrape_uri=http://localhost/server-status?auto"},
		Ports: []corev1.ContainerPort{{
			ContainerPort: 9117,
			Protocol:      corev1.ProtocolTCP,
		}},
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("24M"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("0.02"),
			},
		},
		ImagePullPolicy: corev1.PullIfNotPresent,
	}

	return &result, nil
}
