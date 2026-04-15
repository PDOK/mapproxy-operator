// +kubebuilder:object:generate=true
// +groupName=pdok.nl
package v2

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	GroupVersion  = schema.GroupVersion{Group: "pdok.nl", Version: "v2"}
	SchemaBuilder = &scheme.Builder{GroupVersion: GroupVersion}
	GroupKind     = schema.GroupKind{Group: GroupVersion.Group, Kind: "WMTS"}
	AddToScheme   = SchemaBuilder.AddToScheme
	TypeMetaWMTS  = runtime.TypeMeta{
		Kind:       "WMTS",
		APIVersion: GroupVersion.String(),
	}
)
