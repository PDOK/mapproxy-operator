package mapproxy

import (
	"fmt"

	pdoknlv2 "github.com/pdok/mapproxy-operator/api/v2"
	"github.com/pdok/mapproxy-operator/internal/controller/constants"
	"github.com/pdok/mapproxy-operator/internal/controller/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func GetMapproxyContainer(obj *pdoknlv2.WMTS, images *types.Images) (*corev1.Container, error) {
	resources := getResources(obj)
	env := getEnv(obj)

	var result = corev1.Container{
		Name:      constants.MapproxyName,
		Image:     images.MapproxyImage,
		Resources: resources,
		VolumeMounts: []corev1.VolumeMount{{
			Name:      "data",
			MountPath: "/var/www",
		}, {
			Name:      "mapproxy",
			MountPath: "/srv/mapproxy/config/mapproxy.yaml",
			SubPath:   "mapproxy.yaml",
		}, {
			Name:      "lighttpd",
			MountPath: "/srv/mapproxy/config/include.conf",
			SubPath:   "include.conf",
		}, {
			Name:      "lighttpd",
			MountPath: "/srv/mapproxy/config/response.lua",
			SubPath:   "response.lua",
		}},
		Env: env,
		Lifecycle: &corev1.Lifecycle{
			PreStop: &corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{Command: []string{"sleep", "15"}},
			},
		},
		LivenessProbe:   getLivenessProbe(),
		ReadinessProbe:  getReadinessProbe(obj),
		StartupProbe:    getStartupProbe(obj),
		ImagePullPolicy: corev1.PullIfNotPresent,
	}
	return &result, nil
}

func getResources(obj *pdoknlv2.WMTS) corev1.ResourceRequirements {
	wmtsResources := obj.Spec.PodSpecPatch.Containers[0].Resources

	memoryLimit := resource.MustParse("4G")
	_, exist := wmtsResources.Limits[corev1.ResourceMemory]
	if exist {
		memoryLimit = wmtsResources.Limits[corev1.ResourceMemory]
	}

	cpuRequest := resource.MustParse("1")
	_, exist = wmtsResources.Requests[corev1.ResourceCPU]
	if exist {
		cpuRequest = wmtsResources.Requests[corev1.ResourceCPU]
	}

	limits := corev1.ResourceList{
		corev1.ResourceMemory: memoryLimit,
	}

	_, exists := wmtsResources.Limits[corev1.ResourceEphemeralStorage]
	if exists {
		limits[corev1.ResourceEphemeralStorage] = wmtsResources.Limits[corev1.ResourceEphemeralStorage]
	}

	requests := corev1.ResourceList{
		corev1.ResourceCPU: cpuRequest,
	}

	return corev1.ResourceRequirements{
		Limits:   limits,
		Requests: requests,
	}
}

func getEnv(obj *pdoknlv2.WMTS) []corev1.EnvVar {
	result := []corev1.EnvVar{{
		Name:  "MAX_PROCS",
		Value: "16",
	}, {
		Name:  "MIN_PROCS",
		Value: "16",
	}}

	var mapproxyContainer *corev1.Container
	for _, container := range obj.Spec.PodSpecPatch.Containers {
		if container.Name == constants.MapproxyName {
			mapproxyContainer = &container
			break
		}
	}

	if mapproxyContainer != nil {
		for _, env := range mapproxyContainer.Env {
			if env.Name == "AZURE_STORAGE_CONNECTION_STRING" {
				result = append(result, env)
				break
			}
		}
	}

	_, isDebug := obj.Annotations["pdok.nl/debug"]
	debugValue := "0"
	if isDebug {
		debugValue = "1"
	}
	result = append(result, corev1.EnvVar{
		Name:  "DEBUG",
		Value: debugValue,
	})

	return result
}

func getLivenessProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/mapproxy",
				Port: intstr.IntOrString{Type: intstr.Int, IntVal: 80},
				HTTPHeaders: []corev1.HTTPHeader{{
					Name:  "Content-Type",
					Value: "text/html",
				}},
			},
		},
		InitialDelaySeconds: 20,
		TimeoutSeconds:      30,
		PeriodSeconds:       10,
		FailureThreshold:    1,
	}
}

func getReadinessProbe(obj *pdoknlv2.WMTS) *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: obj.Spec.Service.BaseURL.Path + "/" + obj.Spec.HealthCheck.Querystring,
				Port: intstr.IntOrString{Type: intstr.Int, IntVal: 80},
				HTTPHeaders: []corev1.HTTPHeader{{
					Name:  "Content-Type",
					Value: obj.Spec.HealthCheck.Mimetype,
				}},
			},
		},
		InitialDelaySeconds: 20,
		TimeoutSeconds:      30,
		PeriodSeconds:       10,
		FailureThreshold:    3,
	}
}

func getStartupProbe(obj *pdoknlv2.WMTS) *corev1.Probe {
	commandArray := []string{"/bin/sh", "-c"}
	wmsUrls := obj.GetWmsUrls()
	for _, url := range wmsUrls {
		commandString := fmt.Sprintf("wget -SO- -T 60 -t 5 '%s&request=getCapabilities&service=WMS&version=1.3.0' 2>&1 | egrep -aiA10 'HTTP/1.1 200' | egrep -i 'Content-Type: text/xml|Content-Type: application/vnd.ogc.wms_xml'", url)
		commandArray = append(commandArray, commandString)
	}

	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{Command: commandArray},
		},
		TimeoutSeconds:   500,
		FailureThreshold: 3,
	}
}
