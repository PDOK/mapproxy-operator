package capabilitiesgenerator

import (
	"fmt"

	pdoknlv2 "github.com/pdok/mapproxy-operator/api/v2"
	"github.com/pdok/mapproxy-operator/internal/controller/constants"
	"github.com/pdok/mapproxy-operator/internal/controller/types"
	"github.com/pdok/mapproxy-operator/internal/controller/utils"
	"sigs.k8s.io/yaml"

	corev1 "k8s.io/api/core/v1"
)

func GetCapabilitiesGeneratorInitContainer(_ *pdoknlv2.WMTS, images types.Images) (*corev1.Container, error) {
	initContainer := corev1.Container{
		Name:            constants.CapabilitiesGeneratorName,
		Image:           images.CapabilitiesGeneratorImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Env: []corev1.EnvVar{
			{
				Name:  "SERVICECONFIG",
				Value: "/input/input.yaml",
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			utils.GetDataVolumeMount(),
			utils.GetConfigVolumeMount(constants.ConfigMapCapabilitiesGeneratorVolumeName),
		},
	}
	return &initContainer, nil
}

func GetInput(webservice *pdoknlv2.WMTS) (string, error) {
	return createInputForWMTS(webservice)
}

func createInputForWMTS(wmts *pdoknlv2.WMTS) (string, error) {
	config, err := MapWMTSToCapabilitiesGeneratorInput(wmts)
	if err != nil {
		return "", err
	}
	yamlInput, err := yaml.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal the capabilities generator input to yaml: %w", err)
	}
	return string(yamlInput), nil
}
