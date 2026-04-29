package mapproxy

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	pdoknlv2 "github.com/pdok/mapproxy-operator/api/v2"
	"github.com/pdok/mapproxy-operator/internal/controller/constants"
	"gopkg.in/yaml.v3"
	"k8s.io/utils/ptr"
)

func GetInclude(obj *pdoknlv2.WMTS) (string, error) {
	result := templateInclude
	stringBuilder := strings.Builder{}
	ingressRouteUrls := obj.GetIngressRouteUrls()
	for _, url := range ingressRouteUrls {
		path := url.Path
		stringBuilder.WriteString(fmt.Sprintf(`  "^%s(/1\.0\.0)?/[Ww][Mm][Tt][Ss][Cc]apabilities\.xml" => "/WMTSCapabilities.xml",%s`, path, "\n"))
		for _, layer := range obj.Spec.Service.Layers {
			if len(layer.Styles) > 0 {
				stringBuilder.WriteString(fmt.Sprintf(`  "^%s/%s/legend.png" => "/images/%s",%s`, path, layer.Identifier, layer.Styles[0].Legend.GetBlobKeyName(), "\n"))
			}
		}
		stringBuilder.WriteString(fmt.Sprintf(`  "^%s/(.*)" => "/mapproxy/wmts/$1",%s`, path, "\n"))
	}
	stringBuilder.WriteString(`  "^/mapproxy/.*" => "/hide_direct_url"` + "\n")

	rewriteRules := stringBuilder.String()
	result = strings.ReplaceAll(result, "$REWRITERULES", rewriteRules)
	return result, nil
}

func GetResponse(_ *pdoknlv2.WMTS) (string, error) {
	return responseLua, nil
}

func GetMapproxyConfig(obj *pdoknlv2.WMTS) (MapproxyConfig, error) {
	return MapproxyConfig{
		Services: Services{
			Wmts: ServiceWMTS{
				Kvp:                true,
				Restful:            true,
				FeatureinfoFormats: nil,
			},
		},
		Layers:  getMapproxyLayers(obj),
		Caches:  getMapproxyCaches(obj),
		Sources: getMapproxySources(obj),
		Grids:   getMapproxyGrids(obj),
		Globals: getMapproxyGlobals(obj),
	}, nil
}

func GetMapproxyConfigString(obj *pdoknlv2.WMTS) (string, error) {
	mapproxyConfig, err := GetMapproxyConfig(obj)
	if err != nil {
		return "", err
	}

	bytes, err := yaml.Marshal(mapproxyConfig)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func getMapproxyGlobals(obj *pdoknlv2.WMTS) Globals {
	result := Globals{
		Cache: GlobalsCache{
			MetaBuffer: 360,
		},
		Image: GlobalsImage{
			ResamplingMethod: "bilinear",
			Paletted:         false,
			Formats: GlobalsImageFormats{
				Png24: Png24{
					Format:      "image/png",
					Transparent: true,
				},
			},
		},
	}

	var metaSize string
	if obj.Spec.Options.Cached {
		metaSize = "[2,2]"
		if obj.Spec.Service.Cache.MetaSize != nil {
			metaSize = *obj.Spec.Service.Cache.MetaSize
		}

		result.Cache.BaseDir = to.Ptr("/srv/mapproxy/cache_data")
		result.Cache.LockDir = to.Ptr("/srv/mapproxy/cache_data/locks")
		result.Cache.TileLockDir = to.Ptr("/srv/mapproxy/cache_data/tile_locks")
	} else {
		metaSize = "[1,1]"
		if obj.Spec.Service.Cache.MetaSize != nil {
			metaSize = *obj.Spec.Service.Cache.MetaSize
		}
	}

	// string to separate ints
	splitMetaSize := strings.Split(metaSize, ",")
	elem1, _ := strconv.Atoi(splitMetaSize[0][1:])
	elem2, _ := strconv.Atoi(splitMetaSize[1][0 : len(splitMetaSize[1])-1])
	result.Cache.MetaSize = []int{elem1, elem2}

	return result
}

func getMapproxyLayers(obj *pdoknlv2.WMTS) []Layer {
	result := make([]Layer, 0)
	for _, wmtsLayer := range obj.Spec.Service.Layers {
		myLayer := Layer{
			Name:        wmtsLayer.Identifier,
			Title:       wmtsLayer.Title,
			TileSources: make([]string, 0),
		}
		for _, tms := range obj.Spec.Service.TileMatrixSets {
			myLayer.TileSources = append(myLayer.TileSources, fmt.Sprintf("%s-%s-cache", wmtsLayer.Identifier, tms.CRS[5:]))
		}
		result = append(result, myLayer)
	}
	return result
}

func getMapproxyCaches(obj *pdoknlv2.WMTS) map[string]Cache {
	result := make(map[string]Cache)
	for _, wmtsLayer := range obj.Spec.Service.Layers {
		for _, tms := range obj.Spec.Service.TileMatrixSets {
			layerSrsName := getLayerSrsName(wmtsLayer, tms)
			cacheName := layerSrsName + "-cache"
			cache := Cache{
				Sources:        []string{layerSrsName + "-source"},
				Grids:          []string{tms.CRS},
				DisableStorage: !obj.Spec.Options.Cached,
			}
			if obj.Spec.Options.GetFeatureInfo {
				cache.Sources = append(cache.Sources, layerSrsName+"-source-featureinfo")
			}

			if obj.Spec.Options.Cached {
				cache.CacheDetails = &CacheDetails{
					Type:          "azureblob",
					Directory:     fmt.Sprintf("%s/%s/%s/", obj.Spec.Service.Cache.Azure.BlobPrefix, wmtsLayer.Identifier, tms.CRS[5:]),
					ContainerName: constants.BlobsTilesBucket,
				}
			}

			result[cacheName] = cache
		}

	}
	return result
}

func getLayerSrsName(wmtsLayer pdoknlv2.WMTSLayer, tileMatrixSet pdoknlv2.TileMatrixSet) string {
	return fmt.Sprintf("%s-%s", wmtsLayer.Identifier, tileMatrixSet.CRS[5:])
}

func getMapproxySources(obj *pdoknlv2.WMTS) map[string]Source {
	result := make(map[string]Source)
	addSourcesWithParams(obj, result, "", true, false)
	if obj.Spec.Options.GetFeatureInfo {
		addSourcesWithParams(obj, result, "-featureinfo", false, true)
	}

	return result
}

func addSourcesWithParams(obj *pdoknlv2.WMTS, sources map[string]Source, suffix string, isMap bool, featureInfo bool) {
	for _, wmtsLayer := range obj.Spec.Service.Layers {
		for _, tms := range obj.Spec.Service.TileMatrixSets {
			layerSrsName := getLayerSrsName(wmtsLayer, tms)
			sourceName := fmt.Sprintf("%s-source%s", layerSrsName, suffix)
			var url string
			if isMap {
				url = wmtsLayer.Source.Wms.URL.String()
			} else {
				url = wmtsLayer.Source.Wms.URL.String() + "&feature_count=5"
			}

			var styles *string
			if !featureInfo && len(wmtsLayer.Source.Wms.Styles) > 0 {
				styles = ptr.To(strings.Join(wmtsLayer.Source.Wms.Styles, ","))
			}

			source := Source{
				Type: "wms",
				WMSOpts: SourceWMSOpts{
					Map:         isMap,
					Featureinfo: featureInfo,
					Version:     "1.3.0",
				},
				SupportedSrs: []string{tms.CRS},
				Coverage:     SourceCoverage{},
				MinRes:       getMinRes(tms),
				MaxRes:       getMaxRes(tms),
				Req: SourceReq{
					Layers:      strings.Join(wmtsLayer.Source.Wms.Layers, ","),
					URL:         url,
					Styles:      styles,
					Transparent: wmtsLayer.Source.Wms.Transparent == nil || *wmtsLayer.Source.Wms.Transparent,
				},
			}
			switch tms.CRS {
			case "EPSG:25831": //nolint:goconst
				source.Coverage = SourceCoverage{
					Srs:  "EPSG:25831",
					Bbox: []float64{392962.82637282484, 5520233.281369477, 869237.0659525886, 6212112.288607274},
				}
			case "EPSG:3857": //nolint:goconst
				source.Coverage = SourceCoverage{
					Srs:  "EPSG:3857",
					Bbox: []float64{143380.08950252648, 6416635.003174079, 951740.3437830413, 7544320.730706658},
				}
			default:
				source.Coverage = SourceCoverage{
					Srs:  "EPSG:28992",
					Bbox: []float64{-101552, 210360, 352820, 887600},
				}
			}

			sources[sourceName] = source
		}
	}
}

func getMinRes(tileMatrixSet pdoknlv2.TileMatrixSet) *float64 {
	minZoomLevel := tileMatrixSet.GetMinZoomLevel()
	if minZoomLevel == nil {
		return nil
	}
	switch tileMatrixSet.CRS {
	case "EPSG:28992": //nolint:goconst
		divisor := math.Pow(2.0, float64(*minZoomLevel))
		return ptr.To(3440.64 / divisor)
	case "EPSG:3857": //nolint:goconst
		divisor := math.Pow(2.0, float64(*minZoomLevel))
		return ptr.To(559082264.029 * 0.00028 / divisor)
	case "EPSG:25831": //nolint:goconst
		presetList := []float64{10000000, 5000000, 2500000, 1000000, 500000, 250000, 100000, 75000, 50000, 25000, 10000, 5000, 2500, 1000, 500, 250, 100}
		return ptr.To(presetList[*minZoomLevel] * 0.00028)
	}

	return nil
}

func getMaxRes(tileMatrixSet pdoknlv2.TileMatrixSet) *float64 {
	maxZoomLevel := tileMatrixSet.GetMaxZoomLevel()
	if maxZoomLevel == nil {
		return nil
	}
	switch tileMatrixSet.CRS {
	case "EPSG:28992":
		divisor := math.Pow(2.0, float64(*maxZoomLevel+1))
		return ptr.To(3440.64 / divisor)
	case "EPSG:3857":
		divisor := math.Pow(2.0, float64(*maxZoomLevel+1))
		return ptr.To(559082264.029 * 0.00028 / divisor)
	case "EPSG:25831":
		presetList := []float64{10000000, 5000000, 2500000, 1000000, 500000, 250000, 100000, 75000, 50000, 25000, 10000, 5000, 2500, 1000, 500, 250, 100}
		return ptr.To(presetList[*maxZoomLevel] * 0.00028 * 0.5)
	}

	return nil
}

func getMapproxyGrids(obj *pdoknlv2.WMTS) map[string]Grid {
	result := make(map[string]Grid)
	for _, tms := range obj.Spec.Service.TileMatrixSets {
		switch tms.CRS {
		case "EPSG:28992":
			grid := Grid{
				TileSize: []float64{256, 256},
				Origin:   "nw",
				Srs:      "EPSG:28992",
				Bbox:     []float64{-285401.920, 22598.080, 595401.920, 903401.920},
				BboxSrs:  "EPSG:28992",
				Res:      []float64{},
			}
			i := 0
			for i < 17 {
				grid.Res = append(grid.Res, 3440.64/math.Pow(2.0, float64(i)))
				i++
			}

			result["EPSG:28992"] = grid
		case "EPSG:25831":
			grid := Grid{
				TileSize: []float64{256, 256},
				Origin:   "nw",
				Srs:      "EPSG:25831",
				BboxSrs:  "EPSG:25831",
				Bbox:     []float64{-2404683.40738879, 3997657.58466454, 4046516.592611209, 8298457.5846645385},
				Res:      []float64{2799.9999999999995, 1399.9999999999998, 699.9999999999999, 280.00, 140.00, 70.00, 27.999999999999996, 20.999999999999996, 13.999999999999998, 6.999999999999999, 2.80, 1.40, 0.7, 0.28, 0.14, 0.07, 0.028},
			}
			result["EPSG:25831"] = grid
		case "EPSG:3857":
			grid := Grid{
				TileSize: []float64{256, 256},
				Origin:   "nw",
				Srs:      "EPSG:3857",
				BboxSrs:  "EPSG:3857",
				Bbox:     []float64{-20037508.3427892, -20037508.3427892, 20037508.3427892, 20037508.3427892},
				Res:      []float64{156543.033928041, 78271.5169640204, 39135.7584820102, 19567.8792410051, 9783.93962050256, 4891.96981025128, 2445.98490512564, 1222.99245256282, 611.49622628141, 305.748113140704, 152.874056570352, 76.4370282851762, 38.2185141425881, 19.109257071294, 9.55462853564703, 4.77731426782351, 2.38865713391175, 1.19432856695587, 0.597164283477939, 0.29858214173897, 0.149291070869485, 0.0746455354347424, 0.0373227677173712, 0.0186613838586856, 0.0093306919293428},
			}
			result["EPSG:3857"] = grid
		}
	}
	return result
}

type MapproxyConfig struct { //nolint:revive
	Services Services          `yaml:"services"`
	Layers   []Layer           `yaml:"layers"`
	Caches   map[string]Cache  `yaml:"caches"`
	Sources  map[string]Source `yaml:"sources"`
	Grids    map[string]Grid   `yaml:"grids"`
	Globals  Globals           `yaml:"globals"`
}

func (m *MapproxyConfig) Equal(o *MapproxyConfig) bool {
	if !reflect.DeepEqual(m.Services, o.Services) {
		return false
	}

	if !reflect.DeepEqual(m.Layers, o.Layers) {
		return false
	}

	if !reflect.DeepEqual(m.Caches, o.Caches) {
		return false
	}

	for key, value := range m.Sources {
		other := o.Sources[key]
		if !value.Equal(&other) {
			return false
		}
	}

	if !reflect.DeepEqual(m.Grids, o.Grids) {
		return false
	}

	if !reflect.DeepEqual(m.Globals, o.Globals) {
		return false
	}

	return true
}

type Services struct {
	Wmts ServiceWMTS `yaml:"wmts"`
}

type ServiceWMTS struct {
	Kvp                bool                    `yaml:"kvp"`
	Restful            bool                    `yaml:"restful"`
	FeatureinfoFormats []WMTSFeatureInfoFormat `yaml:"featureinfo_formats,omitempty"`
}

type WMTSFeatureInfoFormat struct {
	MimeType string `yaml:"mime_type"`
	Suffix   string `yaml:"suffix"`
}

type Layer struct {
	Name        string   `yaml:"name"`
	Title       string   `yaml:"title"`
	TileSources []string `yaml:"tile_sources"`
}

type Globals struct {
	Cache GlobalsCache `yaml:"cache"`
	Image GlobalsImage `yaml:"image"`
}

type GlobalsCache struct {
	MetaBuffer  int     `yaml:"meta_buffer"`
	BaseDir     *string `yaml:"base_dir,omitempty"`
	LockDir     *string `yaml:"lock_dir,omitempty"`
	TileLockDir *string `yaml:"tile_lock_dir,omitempty"`
	MetaSize    []int   `yaml:"meta_size"`
}

type GlobalsImage struct {
	ResamplingMethod string              `yaml:"resampling_method"`
	Paletted         bool                `yaml:"paletted"`
	Formats          GlobalsImageFormats `yaml:"formats"`
}

type GlobalsImageFormats struct {
	Png24 Png24 `yaml:"png24"`
}

type Png24 struct {
	Format      string `yaml:"format"`
	Transparent bool   `yaml:"transparent"`
}

type Cache struct {
	Sources        []string      `yaml:"sources"`
	Grids          []string      `yaml:"grids"`
	DisableStorage bool          `yaml:"disable_storage"`
	CacheDetails   *CacheDetails `yaml:"cache,omitempty"`
}

type CacheDetails struct {
	Type          string `yaml:"type"`
	Directory     string `yaml:"directory"`
	ContainerName string `yaml:"containerName"`
}

type Source struct {
	Type         string         `yaml:"type"`
	WMSOpts      SourceWMSOpts  `yaml:"wms_opts"`
	SupportedSrs []string       `yaml:"supported_srs"`
	Coverage     SourceCoverage `yaml:"coverage"`
	MinRes       *float64       `yaml:"min_res,omitempty"`
	MaxRes       *float64       `yaml:"max_res,omitempty"`
	Req          SourceReq      `yaml:"req"`
}

func (s *Source) Equal(o *Source) bool {
	if s.Type != o.Type {
		return false
	}

	if !reflect.DeepEqual(s.WMSOpts, o.WMSOpts) {
		return false
	}

	if !reflect.DeepEqual(s.SupportedSrs, o.SupportedSrs) {
		return false
	}

	if !reflect.DeepEqual(s.Coverage, o.Coverage) {
		return false
	}

	if math.Abs(*s.MinRes-*o.MinRes) > 1e-8 {
		return false
	}

	if math.Abs(*s.MaxRes-*o.MaxRes) > 1e-8 {
		return false
	}

	if !reflect.DeepEqual(s.Req, o.Req) {
		return false
	}

	return true
}

type SourceWMSOpts struct {
	Map         bool   `yaml:"map"`
	Featureinfo bool   `yaml:"featureinfo"`
	Version     string `yaml:"version"`
}

type SourceCoverage struct {
	Srs  string    `yaml:"srs"`
	Bbox []float64 `yaml:"bbox"`
}

type SourceReq struct {
	Layers      string  `yaml:"layers"`
	URL         string  `yaml:"url"`
	Styles      *string `yaml:"styles,omitempty"`
	Transparent bool    `yaml:"transparent"`
}

type Grid struct {
	TileSize []float64 `yaml:"tile_size"`
	Origin   string    `yaml:"origin"`
	Srs      string    `yaml:"srs"`
	Bbox     []float64 `yaml:"bbox"`
	BboxSrs  string    `yaml:"bbox_srs"`
	Res      []float64 `yaml:"res"`
}
