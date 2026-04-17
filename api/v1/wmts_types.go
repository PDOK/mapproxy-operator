package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// NOTE: This type is only used for manual conversion and can not be used for kubernetes functionality

type WMTS struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              WMTSSpec `json:"spec"`
}

type WMTSSpec struct {
	General    GeneralLabels      `json:"general"`
	Kubernetes KubernetesSettings `json:"kubernetes"`
	Options    Options            `json:"options"`
	Service    Service            `json:"service"`
}

type GeneralLabels struct {
	DatasetOwner   string `json:"datasetOwner"`
	Dataset        string `json:"dataset"`
	ServiceVersion string `json:"serviceVersion"`
	DataVersion    string `json:"dataVersion"`
}

type KubernetesSettings struct {
	Autoscaling *Autoscaling `json:"autoscaling"`
	HealthCheck HealthCheck  `json:"healthCheck"`
}

type Autoscaling struct {
	MinReplicas int `json:"minReplicas"`
	MaxReplicas int `json:"maxReplicas"`
}

type HealthCheck struct {
	QueryString string `json:"queryString"`
	Mimetype    string `json:"mimetype"`
}

type Options struct {
	Cached         *bool   `json:"cached"`
	IncludeIngress *bool   `json:"includeIngress"`
	MetaSize       *string `json:"metaSize"`
}

type Service struct {
	Title             string         `json:"title"`
	Abstract          string         `json:"abstract"`
	WmtsPath          string         `json:"wmtsPath"`
	BlobPath          string         `json:"blobPath"`
	AccessConstraints string         `json:"accessConstraints"`
	GetFeatureInfo    bool           `json:"getFeatureInfo"`
	SupportedSrs      []SupportedSrs `json:"supportedSrs"`
	Layers            []Layer        `json:"layers"`
}

type SupportedSrs struct {
	Srs        string   `json:"srs"`
	ZoomLevels []string `json:"zoomLevels"`
}

type Layer struct {
	Name       string    `json:"name"`
	Title      string    `json:"title"`
	Abstract   string    `json:"abstract"`
	LegendFile string    `json:"legendFile"`
	WmsSource  WmsSource `json:"wmsSource"`
}

type WmsSource struct {
	URL         string   `json:"url"`
	Layers      []string `json:"layers,omitempty"`
	Transparent *bool    `json:"transparent"`
	Styles      []string `json:"styles,omitempty"`
}
