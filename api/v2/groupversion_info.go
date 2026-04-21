// +kubebuilder:object:generate=true
// +groupName=pdok.nl
package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	GroupVersion  = schema.GroupVersion{Group: "pdok.nl", Version: "v2"}
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}
	GroupKind     = schema.GroupKind{Group: GroupVersion.Group, Kind: "WMTS"}
	AddToScheme   = SchemeBuilder.AddToScheme
	TypeMetaWMTS  = metav1.TypeMeta{
		Kind:       "WMTS",
		APIVersion: GroupVersion.String(),
	}
)
