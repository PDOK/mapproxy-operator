package v2

import (
	smoothoperatormodel "github.com/pdok/smooth-operator/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:conversion:hub
// +kubebuilder:subresource:status
// versionName=v2
// +kubebuilder:resource:path=wmts
// +kubebuilder:resource:categories=pdok
// +kubebuilder:printcolumn:name="ReadyPods",type=integer,JSONPath=`.status.podSummary[0].ready`
// +kubebuilder:printcolumn:name="DesiredPods",type=integer,JSONPath=`.status.podSummary[0].total`
// +kubebuilder:printcolumn:name="ReconcileStatus",type=string,JSONPath=`.status.conditions[?(@.type == "Reconciled")].reason`

// WMTS is the Schema for the wmts API
type WMTS struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              WMTSSpec                           `json:"spec"`
	Status            smoothoperatormodel.OperatorStatus `json:"status,omitempty"`
}

type WMTSSpec struct {
}
