package v1

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"

	v2 "github.com/pdok/mapproxy-operator/api/v2"
	"github.com/pdok/smooth-operator/model"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

func (src *WMTS) ToV2() (v2.WMTS, error) {
	baseURL := getIngressBaseURL(src)
	cached := true
	if src.Spec.Options.Cached != nil {
		cached = *src.Spec.Options.Cached
	}

	includeIngress := true
	if src.Spec.Options.IncludeIngress != nil {
		includeIngress = *src.Spec.Options.IncludeIngress
	}

	metaSize := "[9,9]"
	if src.Spec.Options.MetaSize != nil {
		metaSize = *src.Spec.Options.MetaSize
	}

	result := v2.WMTS{
		TypeMeta:   v2.TypeMetaWMTS,
		ObjectMeta: src.ObjectMeta,
		Spec: v2.WMTSSpec{
			Options: &v2.WMTSOptions{
				Cached:         cached,
				IncludeIngress: includeIngress,
				GetFeatureInfo: src.Spec.Service.GetFeatureInfo,
			},
			Lifecycle:                    nil,
			HorizontalPodAutoscalerPatch: getHPAPatch(src),
			PodSpecPatch:                 getPodSpecPatch(),
			HealthCheck: &v2.WMTSHealthCheck{
				Querystring: src.Spec.Kubernetes.HealthCheck.QueryString,
				Mimetype:    src.Spec.Kubernetes.HealthCheck.Mimetype,
			},
			IngressRouteURLs: model.IngressRouteURLs{{
				URL: model.URL{URL: baseURL},
			}},
			Service: v2.WMTSService{
				BaseURL:           model.URL{URL: baseURL},
				Title:             src.Spec.Service.Title,
				Abstract:          src.Spec.Service.Abstract,
				AccessConstraints: nil,
				TileMatrixSets:    getTileMatrixSets(src),
				Layers:            getLayers(src),
				Cache: v2.WMTSCache{
					MetaSize: metaSize,
					Azure: v2.AzureCache{
						Container:  "public",
						BlobPrefix: src.Spec.Service.BlobPath,
					},
				},
			},
		},
		Status: model.OperatorStatus{},
	}
	return result, nil
}

func getHPAPatch(src *WMTS) *v2.HorizontalPodAutoscalerPatch {
	var result *v2.HorizontalPodAutoscalerPatch
	if src.Spec.Kubernetes.Autoscaling != nil {
		result = &v2.HorizontalPodAutoscalerPatch{
			MinReplicas: ptr.To(int32(src.Spec.Kubernetes.Autoscaling.MinReplicas)), //nolint:gosec
			MaxReplicas: ptr.To(int32(src.Spec.Kubernetes.Autoscaling.MaxReplicas)), //nolint:gosec
			Metrics:     nil,
			Behavior:    nil,
		}
	}
	return result
}

func getPodSpecPatch() corev1.PodSpec {
	secretRef := "mysecretblobname"

	return corev1.PodSpec{
		InitContainers: []corev1.Container{{
			Name: "blob-download",
			EnvFrom: []corev1.EnvFromSource{{
				Prefix: "",
				ConfigMapRef: &corev1.ConfigMapEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: secretRef},
					Optional:             nil,
				},
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: secretRef},
					Optional:             nil,
				},
			}},
		}},
		Containers: []corev1.Container{{
			Name: "mapproxy",
			Env: []corev1.EnvVar{{
				Name: "AZURE_STORAGE_CONNECTION_STRING",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: secretRef},
						Key:                  "AZURE_STORAGE_CONNECTION_STRING",
						Optional:             nil,
					},
				},
			}},
			Resources: corev1.ResourceRequirements{Limits: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("21Mi")}},
		}},
	}
}

func getIngressBaseURL(src *WMTS) *url.URL {
	wmtsPath := src.Spec.Service.WmtsPath
	url, _ := model.ParseURL("https://service.pdok.nl/" + wmtsPath)
	return url
}

func getTileMatrixSets(src *WMTS) []v2.TileMatrixSet {
	result := make([]v2.TileMatrixSet, 0)
	for _, srsSet := range src.Spec.Service.SupportedSrs {
		tileMatrixSet := v2.TileMatrixSet{
			CRS:        srsSet.Srs,
			ZoomLevels: getZoomLevels(srsSet),
		}
		result = append(result, tileMatrixSet)
	}
	return result
}

type intRange struct {
	minval int
	maxval int
}

func (i *intRange) String() string {
	if i.minval == i.maxval {
		return strconv.Itoa(i.minval)
	}
	return fmt.Sprintf("%d-%d", i.minval, i.maxval)
}

func getZoomLevels(srsSet SupportedSrs) []string {
	result := make([]string, 0)
	if srsSet.ZoomLevels == nil {
		return nil
	}

	if len(srsSet.ZoomLevels) == 0 {
		return result
	}

	zoomLevelInts := make([]int, 0)
	for _, zoomLevel := range srsSet.ZoomLevels {
		zoomLevelInt, _ := strconv.Atoi(zoomLevel)
		zoomLevelInts = append(zoomLevelInts, zoomLevelInt)
	}

	sort.Ints(zoomLevelInts)
	ranges := make([]intRange, 0)
	i := 0
	for i < len(zoomLevelInts) {
		if i == len(zoomLevelInts) {
			ranges = append(ranges, intRange{zoomLevelInts[i], zoomLevelInts[i]})
			break
		}
		// We can collapse subsequent zoomlevels into a range
		if i < len(zoomLevelInts)-1 && zoomLevelInts[i+1]-zoomLevelInts[i] == 1 {
			minVal := zoomLevelInts[i]
			j := i
			for j < len(zoomLevelInts)-1 {
				if zoomLevelInts[j+1]-zoomLevelInts[j] == 1 {
					j++
				} else {
					break
				}
			}
			maxVal := zoomLevelInts[j]
			ranges = append(ranges, intRange{minVal, maxVal})
			i = j + 1
		} else {
			ranges = append(ranges, intRange{zoomLevelInts[i], zoomLevelInts[i]})
			i++
		}
	}

	for _, rang := range ranges {
		result = append(result, rang.String())
	}

	return result
}

func getLayers(src *WMTS) []v2.WMTSLayer {
	result := make([]v2.WMTSLayer, 0)

	for _, layer := range src.Spec.Service.Layers {
		sourceURLString := layer.WmsSource.URL
		sourceURL, _ := model.ParseURL(sourceURLString)

		styles := make([]string, 0)
		if len(layer.WmsSource.Styles) > 0 {
			styles = layer.WmsSource.Styles
		}

		newLayer := v2.WMTSLayer{
			Identifier: layer.Name,
			Title:      layer.Title,
			Abstract:   layer.Abstract,
			Styles: []v2.WMTSLayerStyle{{
				Identifier: "default",
				Legend: v2.StyleLegend{
					BlobKey: fmt.Sprintf("resources/images/%s/%s/%s", src.Spec.General.DatasetOwner, src.Spec.General.Dataset, layer.LegendFile),
				},
			}},
			Source: v2.WMTSLayerSource{
				Wms: v2.SourceWMS{
					URL:         model.URL{URL: sourceURL},
					Transparent: layer.WmsSource.Transparent,
					Layers:      layer.WmsSource.Layers,
					Styles:      styles,
				},
			},
		}
		result = append(result, newLayer)
	}
	return result
}
