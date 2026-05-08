package utils

import (
	v2 "github.com/pdok/mapproxy-operator/api/v2"
	smoothoperatormodel "github.com/pdok/smooth-operator/model"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// Sample WMTS'ses useful for testing

// todo: finish the cache WMTS
var cacheURL = mustParseURL("https://test.example.com/owner/cache/wmts/v1_0")
var cache = v2.WMTS{
	TypeMeta: v2.TypeMetaWMTS,
	ObjectMeta: metav1.ObjectMeta{
		Name: "owner-cache",
		Labels: map[string]string{
			"pdok.nl/dataset-id":      "cache",
			"pdok.nl/owner-id":        "owner",
			"pdok.nl/service-type":    "wmts",
			"pdok.nl/service-version": "v1_0",
		},
	},
	Spec: v2.WMTSSpec{
		Options: &v2.WMTSOptions{
			Cached:         true,
			IncludeIngress: true,
			GetFeatureInfo: false,
		},
		HorizontalPodAutoscalerPatch: &v2.HorizontalPodAutoscalerPatch{
			MinReplicas: ptr.To(int32(2)),
			MaxReplicas: ptr.To(int32(8)),
		},
		PodSpecPatch: corev1.PodSpec{
			InitContainers: []corev1.Container{},
			Containers: []corev1.Container{{
				Name: "mapproxy",
				Env: []corev1.EnvVar{{
					Name: "AZURE_STORAGE_CONNECTION_STRING",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "mysecretblobname",
							},
							Key: "AZURE_STORAGE_CONNECTION_STRING",
						},
					},
				}},
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						"ephemeral-storage": resource.MustParse("21Mi"),
					},
				},
			}},
		},
		HealthCheck: nil,
		IngressRouteURLs: smoothoperatormodel.IngressRouteURLs{{
			URL: smoothoperatormodel.URL{URL: cacheURL},
		}},
		Service: v2.WMTSService{
			BaseURL:           smoothoperatormodel.URL{},
			Title:             "",
			Abstract:          "",
			AccessConstraints: nil,
			TileMatrixSets:    nil,
			Layers:            nil,
			Cache:             nil,
		},
	},
}

var noCache = v2.WMTS{
	TypeMeta: v2.TypeMetaWMTS,
	ObjectMeta: metav1.ObjectMeta{
		Name: "owner-nocache",
		Labels: map[string]string{
			"pdok.nl/dataset-id":      "nocache",
			"pdok.nl/owner-id":        "owner",
			"pdok.nl/service-type":    "wmts",
			"pdok.nl/service-version": "v1_0",
		},
	},
	Spec: v2.WMTSSpec{},
}

var featureInfo = v2.WMTS{
	TypeMeta: v2.TypeMetaWMTS,
	ObjectMeta: metav1.ObjectMeta{
		Name: "owner-featureinfo",
		Labels: map[string]string{
			"pdok.nl/dataset-id":      "featureinfo",
			"pdok.nl/owner-id":        "owner",
			"pdok.nl/service-type":    "wmts",
			"pdok.nl/service-version": "v1_0",
		},
	},
	Spec: v2.WMTSSpec{},
}
