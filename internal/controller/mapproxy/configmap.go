package mapproxy

import (
	_ "embed"
	"fmt"
	"strings"

	pdoknlv2 "github.com/pdok/mapproxy-operator/api/v2"
)

//go:embed template_include.conf
var includeTemplate string

//go:embed response.lua
var responseFile string

func GetInclude(obj *pdoknlv2.WMTS) (string, error) {
	result := includeTemplate
	stringBuilder := strings.Builder{}
	ingressRouteUrls := obj.GetIngressRouteUrls()
	for _, url := range ingressRouteUrls {
		path := url.Path
		stringBuilder.WriteString(fmt.Sprintf(`"  ^/%s(/1\.0\.0)?/[Ww][Mm][Tt][Ss][Cc]apabilities\.xml" => "/WMTSCapabilities.xml",`, path))
		for _, layer := range obj.Spec.Service.Layers {
			if len(layer.Styles) > 0 {
				stringBuilder.WriteString(fmt.Sprintf(`"  ^/%s/%s/legend.png" => "/images/%s",`, path, layer.Identifier, layer.Styles[0].Legend.GetBlobKeyName()))
			}
		}
		stringBuilder.WriteString(fmt.Sprintf(`  "^/%s/(.*)" => "/mapproxy/wmts/$1",`, path))
	}
	stringBuilder.WriteString(`  "^/mapproxy/.*" => "/hide_direct_url"`)

	rewriteRules := stringBuilder.String()
	result = strings.ReplaceAll(result, "$REWRITERULES", rewriteRules)
	return result, nil
}

func GetMapproxyConfig(obj *pdoknlv2.WMTS) (string, error) {
	return "", nil
}

func GetResponse(obj *pdoknlv2.WMTS) (string, error) {
	return responseFile, nil
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
	MetaBuffer int   `yaml:"meta_buffer"`
	MetaSize   []int `yaml:"meta_size"`
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
