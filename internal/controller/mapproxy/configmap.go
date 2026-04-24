package mapproxy

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	pdoknlv2 "github.com/pdok/mapproxy-operator/api/v2"
	"sigs.k8s.io/yaml"
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

func GetMapproxyConfig(obj *pdoknlv2.WMTS) (string, error) {
	mapproxyConfig := MapproxyConfig{
		Services: Services{
			Wmts: ServiceWMTS{
				Kvp:                true,
				Restful:            true,
				FeatureinfoFormats: nil,
			},
		},
		Layers:  make([]Layer, 0),
		Caches:  make(map[string]Cache),
		Sources: make(map[string]Source),
		Grids:   make(map[string]Grid),
		Globals: getMapproxyGlobals(obj),
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

	metaSize := ""
	if obj.Spec.Options.Cached {
		metaSize = "[2,2]"
		if obj.Spec.Service.Cache.MetaSize != "" {
			metaSize = obj.Spec.Service.Cache.MetaSize
		}

		result.Cache.BaseDir = to.Ptr("/srv/mapproxy/cache_data")
		result.Cache.LockDir = to.Ptr("/srv/mapproxy/cache_data/locks")
		result.Cache.TileLockDir = to.Ptr("/srv/mapproxy/cache_data/tile_locks")
	} else {
		metaSize = "[1,1]"
		if obj.Spec.Service.Cache.MetaSize != "" {
			metaSize = obj.Spec.Service.Cache.MetaSize
		}
	}

	// string to separate ints
	splitMetaSize := strings.Split(metaSize, ",")
	elem1, _ := strconv.Atoi(splitMetaSize[0][1:])
	elem2, _ := strconv.Atoi(splitMetaSize[1][0 : len(splitMetaSize[1])-1])
	result.Cache.MetaSize = []int{elem1, elem2}

	return result
}

type MapproxyConfig struct {
	Services Services          `yaml:"services"`
	Layers   []Layer           `yaml:"layers"`
	Caches   map[string]Cache  `yaml:"caches"`
	Sources  map[string]Source `yaml:"sources"`
	Grids    map[string]Grid   `yaml:"grids"`
	Globals  Globals           `yaml:"globals"`
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
	Sources        []string `yaml:"sources"`
	Grids          []string `yaml:"grids"`
	DisableStorage bool     `yaml:"disable_storage"`
}

type Source struct {
	Type         string         `yaml:"type"`
	WMSOpts      string         `yaml:"wms_opts"`
	SupportedSrs []string       `yaml:"supported_srs"`
	Coverage     SourceCoverage `yaml:"coverage"`
	MinRes       float64        `yaml:"min_res"`
	MaxRes       float64        `yaml:"max_res"`
	Req          SourceReq      `yaml:"req"`
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
	Layers      string `yaml:"layers"`
	Url         string `yaml:"url"`
	Styles      string `yaml:"styles"`
	Transparent bool   `yaml:"transparent"`
}

type Grid struct {
	TileSize []float64 `yaml:"tile_size"`
	Origin   string    `yaml:"origin"`
	Srs      string    `yaml:"srs"`
	Bbox     []float64 `yaml:"bbox"`
	Res      []float64 `yaml:"res"`
}
