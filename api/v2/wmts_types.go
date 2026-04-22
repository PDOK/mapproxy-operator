package v2

import (
	"strconv"
	"strings"

	smoothoperatormodel "github.com/pdok/smooth-operator/model"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	SchemeBuilder.Register(&WMTS{}, &WMTSList{})
}

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
	// Functional settings
	Spec WMTSSpec `json:"spec"`
	// Status set by the cluster
	Status smoothoperatormodel.OperatorStatus `json:"status,omitempty"`
}

func (w *WMTS) OperatorStatus() *smoothoperatormodel.OperatorStatus {
	return &w.Status
}

func (w *WMTS) TypedName() string {
	name := w.GetName()
	typeSuffix := strings.ToLower("WMTS")
	if strings.HasSuffix(name, typeSuffix) {
		return name
	}

	return name + "-" + typeSuffix
}

// +kubebuilder:object:root=true
type WMTSList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WMTS `json:"items"`
}

type WMTSSpec struct {
	// Boolean options
	Options *WMTSOptions `json:"options,omitempty"`
	// Optional lifecycle settings
	Lifecycle *smoothoperatormodel.Lifecycle `json:"lifecycle,omitempty"`
	// Scaling behavior
	HorizontalPodAutoscalerPatch *HorizontalPodAutoscalerPatch `json:"horizontalPodAutoscalerPatch,omitempty"`
	// +kubebuilder:validation:Type=object
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	// Strategic merge patch for the pod in the deployment. E.g. to patch the resources or add extra env vars.
	PodSpecPatch corev1.PodSpec `json:"podSpecPatch"`
	// Custom healthcheck options
	HealthCheck *WMTSHealthCheck `json:"healthCheck,omitempty"`
	// Alternative ingress urls
	IngressRouteURLs smoothoperatormodel.IngressRouteURLs `json:"ingressRouteUrls,omitempty"`
	// service configuration
	Service WMTSService `json:"service"`
}

type WMTSOptions struct {
	// Cached enables the pregenerated tiles from a permanent storage
	// +kubebuilder:default:=true
	// +kubebuilder:validation:Optional
	Cached bool `json:"cached"`
	// IncludeIngress dictates whether to deploy an Ingress or ensure none exists.
	// +kubebuilder:default:=true
	// +kubebuilder:validation:Optional
	IncludeIngress bool `json:"includeIngress"`
	// GetFeatureInfo adds a GetFeatureInfo endpoint
	// +kubebuilder:default:=false
	// +kubebuilder:validation:Optional
	GetFeatureInfo bool `json:"getFeatureInfo"`
}

type WMTSService struct {
	// Base url. Distinguished from an actual URL as the path can also be used as a base path for other URLs
	BaseURL smoothoperatormodel.URL `json:"baseUrl"`
	// Service title
	// +kubebuilder:validation:MinLength:=1
	Title string `json:"title"`
	// Service abstract
	// +kubebuilder:validation:MinLength:=1
	Abstract string `json:"abstract"`
	// AccessConstraints URL
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="https://creativecommons.org/publicdomain/zero/1.0/deed.nl"
	AccessConstraints *smoothoperatormodel.URL `json:"accessConstraints,omitempty"`
	// Predefined tile matrices
	TileMatrixSets []TileMatrixSet `json:"tileMatrixSets"`
	// Queryable layers
	Layers []WMTSLayer `json:"layers"`
	// Backing cache layer configuration
	Cache WMTSCache `json:"cache"`
}

// HorizontalPodAutoscalerPatch - copy of autoscalingv2.HorizontalPodAutoscalerSpec without ScaleTargetRef
// This way we don't have to specify the scaleTargetRef field in the CRD.
type HorizontalPodAutoscalerPatch struct {
	MinReplicas *int32                                         `json:"minReplicas,omitempty"`
	MaxReplicas *int32                                         `json:"maxReplicas,omitempty"`
	Metrics     []autoscalingv2.MetricSpec                     `json:"metrics,omitempty"`
	Behavior    *autoscalingv2.HorizontalPodAutoscalerBehavior `json:"behavior,omitempty"`
}

// WMTSHealthCheck is the struct with all fields to configure custom healthchecks
type WMTSHealthCheck struct {
	Querystring string `json:"querystring"`
	// +kubebuilder:validation:Pattern=(image/png|text/xml|text/html)
	Mimetype string `json:"mimetype"`
}

// TileMatrixSet specifies the predefined tile matrices per CRS
type TileMatrixSet struct {
	// The specified CRS
	// +kubebuilder:validation:Pattern:="^EPSG:(28992|25831|25832|3034|3035|3857|4258|4326)|WGS84$"
	CRS string `json:"crs"`
	// +kubebuilder:validation:items:Pattern:="^[0-9]{1,2}(-[0-9]{1,2})?$"
	ZoomLevels []string `json:"zoomLevels,omitempty"`
}

// Used for generation of capabilities
func (t *TileMatrixSet) GetMaxZoomLevel() *int {
	if len(t.ZoomLevels) == 0 {
		return nil
	}

	result := 0
	for _, zoomLevel := range t.ZoomLevels {
		split := strings.Split(zoomLevel, "-")
		if len(split) == 2 {
			maxVal, _ := strconv.Atoi(split[1])
			result = max(result, maxVal)
		} else {
			maxVal, _ := strconv.Atoi(split[0])
			result = max(result, maxVal)
		}
	}
	return &result
}

// WMTSLayer describes the layer provided to the service consumer
type WMTSLayer struct {
	// The unique reference of the layer
	// +kubebuilder:validation:MinLength:=1
	Identifier string `json:"identifier"`
	// The title of the layer
	// +kubebuilder:validation:MinLength:=1
	Title string `json:"title"`
	// The abstract of the layer
	// +kubebuilder:validation:MinLength:=1
	Abstract string `json:"abstract"`
	// Applied styles to the layer
	Styles []WMTSLayerStyle `json:"styles"`
	// The backing data source of the layer
	Source WMTSLayerSource `json:"source"`
}

// WMTSLayerStyle
type WMTSLayerStyle struct {
	// The style identifier unique for the layer
	// +kubebuilder:validation:MinLength:=1
	Identifier string `json:"identifier"`
	// Legend information
	Legend StyleLegend `json:"legend"`
}

// StyleLegend Legend information, now a small wrapper around a blob key
type StyleLegend struct {
	// Blob key location of the style
	// +kubebuilder:validation:MinLength:=1
	BlobKey string `json:"blobKey"`
}

type WMTSLayerSource struct {
	// A WMS as a data source
	Wms SourceWMS `json:"wms"`
}

type SourceWMS struct {
	// The WMS url used for retrieving maps
	URL smoothoperatormodel.URL `json:"url"`
	// The generated images have a transparent background
	// +kubebuilder:default:=true
	// +kubebuilder:validation:Optional
	Transparent *bool `json:"transparent,omitempty"`
	// References to layer names
	Layers []string `json:"layers"`
	// References to style names
	Styles []string `json:"styles"`
}

// WMTSCache Information used to retrieve cached data
type WMTSCache struct {
	// Cache retrieval dimensions
	// +kubebuilder:default="[9,9]"
	// +kubebuilder:validation:Pattern="^\\[[0-9],[0-9]\\]$"
	MetaSize string `json:"metaSize"`
	// The azure block. At the moment it is the only cache backing option
	Azure AzureCache `json:"azure"`
}

// AzureCache Cache information based on the Azure Blob Store
type AzureCache struct {
	// The blob storage container
	// +kubebuilder:validation:MinLength:=1
	Container string `json:"container"`
	// The blob store prefix on the container
	// +kubebuilder:validation:MinLength:=1
	BlobPrefix string `json:"blobPrefix"`
}
