package bases

import (
	_ "embed"

	"github.com/pdok/smooth-operator/pkg/validation"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

//go:embed pdok.nl_wmts.yaml
var wmtsCRD []byte

func init() {
	wmts, err := GetWmtsCRD()
	if err != nil {
		panic(err)
	}

	err = validation.AddValidator(wmts)
	if err != nil {
		panic(err)
	}
}

func GetWmtsCRD() (v1.CustomResourceDefinition, error) {
	crd := v1.CustomResourceDefinition{}
	err := yaml.Unmarshal(wmtsCRD, &crd)
	return crd, err
}
