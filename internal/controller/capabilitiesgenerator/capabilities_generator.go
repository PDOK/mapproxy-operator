package capabilitiesgenerator

import (
	pdoknlv2 "github.com/pdok/mapproxy-operator/api/v2"
	"github.com/pdok/mapproxy-operator/internal/controller/constants"
	"github.com/pdok/mapproxy-operator/internal/controller/types"
	"github.com/pdok/mapproxy-operator/internal/controller/utils"

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

func GetInput(webservice *pdoknlv2.WMTS) (input string, err error) {
	return createInputForWMTS(webservice)
}

func createInputForWMTS(wmts *pdoknlv2.WMTS) (config string, err error) {
	return "", nil
}
