package utils //nolint:revive

import (
	"net/url"

	v2 "github.com/pdok/mapproxy-operator/api/v2"
	"github.com/pdok/mapproxy-operator/internal/controller/types"
	smoothoperatormodel "github.com/pdok/smooth-operator/model"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// Sample WMTS'ses useful for testing

const (
	testImageName1 = "test.test/image:test1"
	testImageName2 = "test.test/image:test2"
	testImageName3 = "test.test/image:test3"
	testImageName4 = "test.test/image:test4"
	testImageName5 = "test.test/image:test5"
)

var TestImages = types.Images{
	ApacheExporterImage:        testImageName1,
	CapabilitiesGeneratorImage: testImageName2,
	KvpToRestfulImage:          testImageName3,
	MapproxyImage:              testImageName4,
	MultiToolImage:             testImageName5,
}

var cacheURL = mustParseURL("https://test.example.com/owner/cache/wmts/v1_0")
var Cache = v2.WMTS{ //nolint:dupl
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
			InitContainers: []corev1.Container{
				{
					Name: "blob-download",
					EnvFrom: []corev1.EnvFromSource{{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "mysecretblobreference"},
						},
					}, {
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "mysecretblobreference"},
						},
					}},
					Resources: corev1.ResourceRequirements{},
				},
			},
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
		HealthCheck: &v2.WMTSHealthCheck{
			Querystring: "mylayer/EPSG:28992/1/1/1.png",
			Mimetype:    "image/png",
		},
		IngressRouteURLs: smoothoperatormodel.IngressRouteURLs{{
			URL: smoothoperatormodel.URL{URL: cacheURL},
		}},
		Service: v2.WMTSService{
			BaseURL: smoothoperatormodel.URL{
				URL: cacheURL,
			},
			Title:             "My service title",
			Abstract:          "My service abstract",
			AccessConstraints: nil,
			TileMatrixSets: []v2.TileMatrixSet{{
				CRS:        "EPSG:28992",
				ZoomLevels: []string{"0-12"},
			}, {
				CRS:        "EPSG:28992",
				ZoomLevels: []string{"13-15"},
			}},
			Layers: []v2.WMTSLayer{{
				Identifier: "layeridentifier",
				Title:      "My layer title",
				Abstract:   "My layer abstract",
				Styles: []v2.WMTSLayerStyle{{
					Identifier: "default",
					Legend: v2.StyleLegend{
						BlobKey: "resources/images/owner/dataset/mylegendpicture.png",
					},
				}},
				Source: v2.WMTSLayerSource{
					Wms: v2.SourceWMS{
						URL:    smoothoperatormodel.URL{URL: mustParseURL("https://my-wms-mapserver/mapserver?MAP_RESOLUTION=92")},
						Layers: []string{"layer1identifier"},
						Styles: nil,
					},
				},
			}},
			Cache: &v2.WMTSCache{
				MetaSize: &v2.CacheMetaSize{
					Rows: 8,
					Cols: 8,
				},
				Azure: &v2.AzureCache{
					Container:  "tiles",
					BlobPrefix: "owner",
				},
			},
		},
	},
}

var noCacheURL = mustParseURL("https://test.example.com/owner/nocache/wmts/v1_0")
var NoCache = v2.WMTS{
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
	Spec: v2.WMTSSpec{
		Options: &v2.WMTSOptions{
			Cached:         false,
			IncludeIngress: true,
			GetFeatureInfo: false,
		},
		HorizontalPodAutoscalerPatch: &v2.HorizontalPodAutoscalerPatch{
			MinReplicas: ptr.To(int32(2)),
			MaxReplicas: ptr.To(int32(8)),
		},
		PodSpecPatch: corev1.PodSpec{
			InitContainers: []corev1.Container{
				{
					Name: "blob-download",
					EnvFrom: []corev1.EnvFromSource{{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "mysecretblobreference"},
						},
					}, {
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "mysecretblobreference"},
						},
					}},
					Resources: corev1.ResourceRequirements{},
				},
			},
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
		HealthCheck: &v2.WMTSHealthCheck{
			Querystring: "mylayer/EPSG:28992/1/1/1.png",
			Mimetype:    "image/png",
		},
		IngressRouteURLs: smoothoperatormodel.IngressRouteURLs{{
			URL: smoothoperatormodel.URL{URL: noCacheURL},
		}},
		Service: v2.WMTSService{
			BaseURL: smoothoperatormodel.URL{
				URL: noCacheURL,
			},
			Title:             "My service title",
			Abstract:          "My service abstract",
			AccessConstraints: nil,
			TileMatrixSets: []v2.TileMatrixSet{{
				CRS:        "EPSG:28992",
				ZoomLevels: []string{"0-12"},
			}, {
				CRS:        "EPSG:28992",
				ZoomLevels: []string{"13-15"},
			}},
			Layers: []v2.WMTSLayer{{
				Identifier: "layeridentifier",
				Title:      "My layer title",
				Abstract:   "My layer abstract",
				Styles: []v2.WMTSLayerStyle{{
					Identifier: "default",
					Legend: v2.StyleLegend{
						BlobKey: "resources/images/owner/dataset/mylegendpicture.png",
					},
				}},
				Source: v2.WMTSLayerSource{
					Wms: v2.SourceWMS{
						URL:    smoothoperatormodel.URL{URL: mustParseURL("https://my-wms-mapserver/mapserver?MAP_RESOLUTION=92")},
						Layers: []string{"layer1identifier"},
						Styles: nil,
					},
				},
			}},
		},
	},
}

var featureInfoURL = mustParseURL("https://test.example.com/owner/featureinfo/wmts/v1_0")
var FeatureInfo = v2.WMTS{ //nolint:dupl
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
	Spec: v2.WMTSSpec{
		Options: &v2.WMTSOptions{
			Cached:         true,
			IncludeIngress: true,
			GetFeatureInfo: true,
		},
		HorizontalPodAutoscalerPatch: &v2.HorizontalPodAutoscalerPatch{
			MinReplicas: ptr.To(int32(2)),
			MaxReplicas: ptr.To(int32(8)),
		},
		PodSpecPatch: corev1.PodSpec{
			InitContainers: []corev1.Container{
				{
					Name: "blob-download",
					EnvFrom: []corev1.EnvFromSource{{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "mysecretblobreference"},
						},
					}, {
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "mysecretblobreference"},
						},
					}},
					Resources: corev1.ResourceRequirements{},
				},
			},
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
		HealthCheck: &v2.WMTSHealthCheck{
			Querystring: "mylayer/EPSG:28992/1/1/1.png",
			Mimetype:    "image/png",
		},
		IngressRouteURLs: smoothoperatormodel.IngressRouteURLs{{
			URL: smoothoperatormodel.URL{URL: featureInfoURL},
		}},
		Service: v2.WMTSService{
			BaseURL: smoothoperatormodel.URL{
				URL: featureInfoURL,
			},
			Title:             "My service title",
			Abstract:          "My service abstract",
			AccessConstraints: nil,
			TileMatrixSets: []v2.TileMatrixSet{{
				CRS:        "EPSG:28992",
				ZoomLevels: []string{"0-12"},
			}, {
				CRS:        "EPSG:28992",
				ZoomLevels: []string{"13-15"},
			}},
			Layers: []v2.WMTSLayer{{
				Identifier: "layeridentifier",
				Title:      "My layer title",
				Abstract:   "My layer abstract",
				Styles: []v2.WMTSLayerStyle{{
					Identifier: "default",
					Legend: v2.StyleLegend{
						BlobKey: "resources/images/owner/dataset/mylegendpicture.png",
					},
				}},
				Source: v2.WMTSLayerSource{
					Wms: v2.SourceWMS{
						URL:    smoothoperatormodel.URL{URL: mustParseURL("https://my-wms-mapserver/mapserver?MAP_RESOLUTION=92")},
						Layers: []string{"layer1identifier"},
						Styles: nil,
					},
				},
			}},
			Cache: &v2.WMTSCache{
				MetaSize: &v2.CacheMetaSize{
					Rows: 8,
					Cols: 8,
				},
				Azure: &v2.AzureCache{
					Container:  "tiles",
					BlobPrefix: "owner",
				},
			},
		},
	},
}

func mustParseURL(input string) *url.URL {
	result, _ := url.Parse(input)
	return result
}
