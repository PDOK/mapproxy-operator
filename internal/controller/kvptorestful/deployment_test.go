package kvptorestful

import (
	"testing"

	"github.com/pdok/mapproxy-operator/internal/controller/utils"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestDeployment(t *testing.T) {
	actual, err := GetKvpToRestfulContainer(&utils.TestImages)
	assert.NoError(t, err)

	assert.Equal(t, expected, *actual)
}

var expected = v1.Container{
	Name:    "wmts-kvp-to-restful",
	Image:   "test.test/image:test3",
	Command: []string{"wmts-kvp-to-restful"},
	Args:    []string{"-host=http://127.0.0.1:80", "-t=/var/www/WMTSCapabilities.xml", "-d=15"},
	Ports: []v1.ContainerPort{{
		ContainerPort: 9001,
		Protocol:      "TCP",
	}},
	Resources: v1.ResourceRequirements{
		Limits: v1.ResourceList{
			"memory": resource.MustParse("128M"),
		},
		Requests: v1.ResourceList{
			"cpu": resource.MustParse("0.1"),
		},
	},
	VolumeMounts: []v1.VolumeMount{{
		Name:      "data",
		ReadOnly:  false,
		MountPath: "/var/www",
	}},
	VolumeDevices: nil,
	LivenessProbe: &v1.Probe{
		ProbeHandler: v1.ProbeHandler{
			Exec: nil,
			HTTPGet: &v1.HTTPGetAction{
				Path: "/health",
				Port: intstr.IntOrString{
					Type:   0,
					IntVal: 9001,
				},
			},
		},
		InitialDelaySeconds: 5,
		TimeoutSeconds:      20,
		PeriodSeconds:       10,
	},
	ReadinessProbe: &v1.Probe{
		ProbeHandler: v1.ProbeHandler{
			Exec: nil,
			HTTPGet: &v1.HTTPGetAction{
				Path: "/health",
				Port: intstr.IntOrString{
					Type:   0,
					IntVal: 9001,
				},
			},
		},
		InitialDelaySeconds: 5,
		TimeoutSeconds:      20,
		PeriodSeconds:       10,
	},
	StartupProbe:             nil,
	Lifecycle:                nil,
	TerminationMessagePath:   "/dev/termination-log",
	TerminationMessagePolicy: "File",
	ImagePullPolicy:          "IfNotPresent",
	SecurityContext:          nil,
	Stdin:                    false,
	StdinOnce:                false,
	TTY:                      false,
}
