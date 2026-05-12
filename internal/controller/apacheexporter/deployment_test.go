package apacheexporter

import (
	"testing"

	"github.com/pdok/mapproxy-operator/internal/controller/utils"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestDeployment(t *testing.T) {
	actual, err := GetApacheContainer(&utils.TestImages)
	assert.NoError(t, err)

	assert.Equal(t, expected, *actual)

}

var expected = v1.Container{
	Name:       "apache-exporter",
	Image:      "test.test/image:test1",
	Command:    nil,
	Args:       []string{"-scrape_uri=http://localhost/server-status?auto"},
	WorkingDir: "",
	Ports: []v1.ContainerPort{{
		ContainerPort: 9117,
		Protocol:      "TCP",
	}},
	Resources: v1.ResourceRequirements{
		Limits: v1.ResourceList{
			"memory": resource.MustParse("24M"),
		},
		Requests: v1.ResourceList{
			"cpu": resource.MustParse("0.02"),
		},
	},
	ImagePullPolicy: "IfNotPresent",
}
