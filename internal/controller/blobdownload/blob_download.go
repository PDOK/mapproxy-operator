package blobdownload

import (
	"fmt"
	"strings"

	pdoknlv2 "github.com/pdok/mapproxy-operator/api/v2"
	"github.com/pdok/mapproxy-operator/internal/controller/constants"
	"github.com/pdok/mapproxy-operator/internal/controller/mapperutils"
	"github.com/pdok/mapproxy-operator/internal/controller/types"
	"github.com/pdok/mapproxy-operator/internal/controller/utils"
	"github.com/pkg/errors"

	"k8s.io/utils/strings/slices"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	legendPath = "/var/www/images"
)

func GetBlobDownloadInitContainer(obj *pdoknlv2.WMTS, images types.Images) (*corev1.Container, error) {
	blobkeys := []string{}
	for _, layer := range obj.Spec.Service.Layers {
		for _, style := range layer.Styles {
			blobKey := style.Legend.BlobKey
			if !slices.Contains(blobkeys, blobKey) {
				blobkeys = append(blobkeys, blobKey)
			}
		}
	}

	initContainer := corev1.Container{
		Name:            constants.BlobDownloadName,
		Image:           images.MultiToolImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("0.15"),
			},
		},
		Command: []string{"/bin/sh", "-c"},
		VolumeMounts: []corev1.VolumeMount{
			utils.GetDataVolumeMount(),
		},
	}

	// Additional blob-download configuration
	args, err := GetArgs(blobkeys)
	if err != nil {
		return nil, err
	}
	initContainer.Args = []string{args}

	resourceCPU := resource.MustParse("0.2")
	if use, _ := mapperutils.UseEphemeralVolume(obj); use {
		resourceCPU = resource.MustParse("1")
	}
	initContainer.Resources.Limits = corev1.ResourceList{
		corev1.ResourceCPU: resourceCPU,
	}

	var envFromSource []corev1.EnvFromSource
	for _, container := range obj.Spec.PodSpecPatch.InitContainers {
		if container.Name == "blob-download" {
			envFromSource = container.EnvFrom
		}
	}

	if envFromSource == nil {
		return nil, errors.New("could not find the envFrom for the init container 'blob-download'")
	}
	initContainer.EnvFrom = envFromSource

	return &initContainer, nil
}

func GetArgs(blobkeys []string) (args string, err error) {
	var sb strings.Builder
	createConfig(&sb)
	writeLine(&sb, "mkdir /var/www/images")
	for _, blobKey := range blobkeys {
		fileName, err := getFilenameFromBlobKey(blobKey)
		if err != nil {
			return "", err
		}

		writeLine(&sb, "rclone copyto blobs:/%s %s/%s", blobKey, legendPath, fileName)
	}

	return sb.String(), nil
}

func createConfig(sb *strings.Builder) {
	writeLine(sb, "set -e;")
	writeLine(sb, "rclone config create --non-interactive --obscure blobs azureblob endpoint $BLOBS_ENDPOINT account $BLOBS_ACCOUNT key $BLOBS_KEY use_emulator true;")
}

func getFilenameFromBlobKey(blobKey string) (string, error) {
	index := strings.LastIndex(blobKey, "/")
	if index == -1 {
		return "", fmt.Errorf("could not determine filename from blobkey %s", blobKey)
	}
	return blobKey[index+1:], nil
}

func writeLine(sb *strings.Builder, format string, a ...any) { //nolint:goprintffuncname
	sb.WriteString(fmt.Sprintf(format, a...) + "\n")
}
