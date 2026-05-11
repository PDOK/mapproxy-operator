package blobdownload

import (
	"testing"

	"github.com/pdok/mapproxy-operator/internal/controller/utils"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestDeployment(t *testing.T) {
	actual, err := GetBlobDownloadInitContainer(&utils.Cache, utils.TestImages)
	assert.NoError(t, err)

	assert.Equal(t, expected, *actual)
}

var expected = v1.Container{
	Name:       "blob-download",
	Image:      "test.test/image:test5",
	Command:    []string{"/bin/sh", "-c"},
	Args:       []string{"set -e;\nrclone config create --non-interactive --obscure blobs azureblob endpoint $BLOBS_ENDPOINT account $BLOBS_ACCOUNT key $BLOBS_KEY use_emulator true;\nmkdir -p /var/www/images;\nrclone copyto blobs:/resources/images/owner/dataset/mylegendpicture.png /var/www/images/mylegendpicture.png;\n"},
	WorkingDir: "",
	Ports:      nil,
	EnvFrom: []v1.EnvFromSource{{
		Prefix: "",
		ConfigMapRef: &v1.ConfigMapEnvSource{
			LocalObjectReference: v1.LocalObjectReference{
				Name: "mysecretblobreference",
			},
			Optional: nil,
		},
		SecretRef: nil,
	}, {
		Prefix:       "",
		ConfigMapRef: nil,
		SecretRef: &v1.SecretEnvSource{
			LocalObjectReference: v1.LocalObjectReference{
				Name: "mysecretblobreference",
			},
			Optional: nil,
		},
	}},
	Resources: v1.ResourceRequirements{
		Limits: v1.ResourceList{
			"cpu": resource.MustParse("0.2"),
		},
		Requests: v1.ResourceList{
			"cpu": resource.MustParse("0.15"),
		},
	},
	VolumeMounts: []v1.VolumeMount{{
		Name:              "data",
		ReadOnly:          false,
		RecursiveReadOnly: nil,
		MountPath:         "/var/www",
		SubPath:           "",
		MountPropagation:  nil,
		SubPathExpr:       "",
	}},
	ImagePullPolicy: "IfNotPresent",
}
