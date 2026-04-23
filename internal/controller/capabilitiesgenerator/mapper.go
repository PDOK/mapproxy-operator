package capabilitiesgenerator

import (
	"encoding/xml"
	"fmt"

	v2 "github.com/pdok/mapproxy-operator/api/v2"
	"github.com/pdok/ogc-specifications/pkg/wmts100"
	"github.com/pdok/ogc-specifications/pkg/wsc110"

	capabilitiesgenerator "github.com/pdok/ogc-capabilities-generator/pkg/config"
)

func MapWMTSToCapabilitiesGeneratorInput(wmts *v2.WMTS) (*capabilitiesgenerator.Config, error) {
	accessContraints := wmts.Spec.Service.AccessConstraints
	accessConstraintsString := "none"
	if accessContraints != nil {
		accessConstraintsString = accessContraints.String()
	}

	result := capabilitiesgenerator.Config{
		Global: capabilitiesgenerator.Global{
			Prefix: "wmts",
		},
		Services: capabilitiesgenerator.Services{
			WMTS100Config: &capabilitiesgenerator.WMTS100Config{
				Filename: "/output/WMTSCapabilities.xml",
				Wmts100: wmts100.GetCapabilitiesResponse{
					XMLName:    xml.Name{},
					Namespaces: wmts100.Namespaces{},
					ServiceIdentification: wmts100.ServiceIdentification{
						Title:             wmts.Spec.Service.Title,
						Abstract:          wmts.Spec.Service.Abstract,
						Fees:              "none",
						AccessConstraints: accessConstraintsString,
					},
					ServiceProvider: nil,
					OperationsMetadata: &wmts100.OperationsMetadata{
						Operation: []wmts100.Operation{getCapabilitiesOperation(wmts), getTileOperation(wmts)},
					},
					Contents: getContents(wmts),
					ServiceMetadataURL: &wmts100.ServiceMetadataURL{
						Href: wmts.Spec.Service.BaseURL.String() + "/WMTSCapabilities.xml",
					},
				},
			},
		},
	}

	featureOperation := getFeatureInfoOperation(wmts)
	if featureOperation != nil {
		result.Services.WMTS100Config.Wmts100.OperationsMetadata.Operation = append(result.Services.WMTS100Config.Wmts100.OperationsMetadata.Operation, *featureOperation)
	}

	return &result, nil
}

type DCP struct {
	HTTP struct {
		Get  *wmts100.Method `xml:"ows:Get,omitempty" yaml:"get,omitempty"`
		Post *wmts100.Method `xml:"ows:Post,omitempty" yaml:"post,omitempty"`
	} `xml:"ows:HTTP" yaml:"http"`
}

type Method struct {
	Type       string       `xml:"xlink:type,attr" yaml:"type"`
	Href       string       `xml:"xlink:href,attr" yaml:"href"`
	Constraint []Constraint `xml:"ows:Constraint" yaml:"constraint"`
}

type Constraint struct {
	Name          string        `xml:"name,attr" yaml:"name"`
	AllowedValues AllowedValues `xml:"ows:AllowedValues" yaml:"allowedValues"`
}

type AllowedValues struct {
	Value []string `xml:"ows:Value" yaml:"value"`
}

func getCapabilitiesOperation(wmts *v2.WMTS) wmts100.Operation {
	getMethod := getEncodingGetMethod(wmts)

	return wmts100.Operation{
		Name: "GetCapabilities",
		DCP: DCP{HTTP: struct {
			Get  *wmts100.Method `xml:"ows:Get,omitempty" yaml:"get,omitempty"`
			Post *wmts100.Method `xml:"ows:Post,omitempty" yaml:"post,omitempty"`
		}{Get: libMethodFromLocalMethod(&getMethod), Post: nil}},
		Parameter:   nil,
		Constraints: nil,
	}
}

func getTileOperation(wmts *v2.WMTS) wmts100.Operation {
	getMethod := getEncodingGetMethod(wmts)

	return wmts100.Operation{
		Name: "GetTile",
		DCP: DCP{HTTP: struct {
			Get  *wmts100.Method `xml:"ows:Get,omitempty" yaml:"get,omitempty"`
			Post *wmts100.Method `xml:"ows:Post,omitempty" yaml:"post,omitempty"`
		}{Get: libMethodFromLocalMethod(&getMethod), Post: nil}},
		Parameter:   nil,
		Constraints: nil,
	}
}

func getFeatureInfoOperation(wmts *v2.WMTS) *wmts100.Operation {
	if !wmts.Spec.Options.GetFeatureInfo {
		return nil
	}

	getMethod := getEncodingGetMethod(wmts)

	return &wmts100.Operation{
		Name: "GetFeatureInfo",
		DCP: DCP{HTTP: struct {
			Get  *wmts100.Method `xml:"ows:Get,omitempty" yaml:"get,omitempty"`
			Post *wmts100.Method `xml:"ows:Post,omitempty" yaml:"post,omitempty"`
		}{Get: libMethodFromLocalMethod(&getMethod), Post: nil}},
		Parameter:   nil,
		Constraints: nil,
	}
}

func getEncodingGetMethod(wmts *v2.WMTS) Method {
	return Method{
		Type: "simple",
		Href: wmts.Spec.Service.BaseURL.String() + "?",
		Constraint: []Constraint{{
			Name: "GetEncoding",
			AllowedValues: AllowedValues{
				Value: []string{"KVP"},
			},
		}},
	}
}

func libMethodFromLocalMethod(method *Method) *wmts100.Method {
	if method == nil {
		return nil
	}

	constraints := make([]struct {
		Name          string `xml:"name,attr" yaml:"name"`
		AllowedValues struct {
			Value []string `xml:"ows:Value" yaml:"value"`
		} `xml:"ows:AllowedValues" yaml:"allowedValues"`
	}, 0)

	for _, constraint := range method.Constraint {

		newConstr := struct {
			Name          string `xml:"name,attr" yaml:"name"`
			AllowedValues struct {
				Value []string `xml:"ows:Value" yaml:"value"`
			} `xml:"ows:AllowedValues" yaml:"allowedValues"`
		}{
			Name: constraint.Name,
			AllowedValues: struct {
				Value []string `xml:"ows:Value" yaml:"value"`
			}{
				Value: constraint.AllowedValues.Value,
			},
		}
		constraints = append(constraints, newConstr)
	}

	return &wmts100.Method{
		Type:       method.Type,
		Href:       method.Href,
		Constraint: constraints,
	}
}

func getContents(wmts *v2.WMTS) wmts100.Contents {
	result := wmts100.Contents{
		Layer:         []wmts100.Layer{},
		TileMatrixSet: []wmts100.TileMatrixSet{},
	}

	for _, layer := range wmts.Spec.Service.Layers {
		layerCapabilities := getLayerCapabilities(wmts, layer)
		result.Layer = append(result.Layer, layerCapabilities)
	}

	for _, tilematrixSet := range wmts.Spec.Service.TileMatrixSets {
		result.TileMatrixSet = append(result.TileMatrixSet, getTileMatrixSetCapabilities(tilematrixSet))
	}
	return result
}

func getTileMatrixSetCapabilities(tileMatrixSet v2.TileMatrixSet) wmts100.TileMatrixSet {
	result := wmts100.TileMatrixSet{
		Identifier:   tileMatrixSet.CRS,
		SupportedCRS: "",
		TileMatrix:   nil,
	}

	if len(tileMatrixSet.ZoomLevels) > 0 {
		result.TileMatrix = make([]wmts100.TileMatrix, 0)
		maxZoomLevel := *tileMatrixSet.GetMaxZoomLevel()
		for i := range maxZoomLevel + 1 {
			s := fmt.Sprintf("%02d", i)
			result.TileMatrix = append(result.TileMatrix, wmts100.TileMatrix{Identifier: s})
		}

	}

	return result
}

func getLayerCapabilities(wmts *v2.WMTS, layer v2.WMTSLayer) wmts100.Layer {
	var result = wmts100.Layer{
		Title:    layer.Title,
		Abstract: layer.Abstract,
		WGS84BoundingBox: wsc110.WGS84BoundingBox{
			LowerCorner: wsc110.Position{-1.65729160235, 48.0405018704},
			UpperCorner: wsc110.Position{12.4317272654, 56.1105896442},
		},
		Identifier:        layer.Identifier,
		TileMatrixSetLink: getTileMatrixSetLinks(wmts),
		ResourceURL:       getResourceUrls(wmts, &layer),
	}

	if wmts.Spec.Options.GetFeatureInfo {
		result.InfoFormat = []string{"text/html", "text/xml", "application/json", "text/plain"}
	}

	for _, style := range layer.Styles {
		var legendUrl []*wmts100.LegendURL
		if style.Legend.BlobKey != "" {
			legendUrl = []*wmts100.LegendURL{{
				Format: "image/png",
				Href:   style.Legend.BlobKey,
			}}
		}

		styleCapabilities := wmts100.Style{
			Identifier: style.Identifier,
			LegendURL:  legendUrl,
		}
		result.Style = append(result.Style, styleCapabilities)
	}

	return result
}

func getResourceUrls(wmts *v2.WMTS, layer *v2.WMTSLayer) []wmts100.ResourceURL {
	layerBaseReference := wmts.Spec.Service.BaseURL.String() + "/" + layer.Identifier
	baseRowColReference := "/{TileMatrixSet}/{TileMatrix}/{TileCol}/{TileRow}"

	result := []wmts100.ResourceURL{{
		Format:       "image/png",
		ResourceType: "tile",
		Template:     layerBaseReference + baseRowColReference + ".png",
	}}

	if wmts.Spec.Options.GetFeatureInfo {
		result = append(result,
			wmts100.ResourceURL{
				Format:       "text/html",
				ResourceType: "FeatureInfo",
				Template:     layerBaseReference + baseRowColReference + "/{I}/{J}.html",
			},
			wmts100.ResourceURL{
				Format:       "text/xml",
				ResourceType: "FeatureInfo",
				Template:     layerBaseReference + baseRowColReference + "/{I}/{J}.xml",
			},
			wmts100.ResourceURL{
				Format:       "application/json",
				ResourceType: "FeatureInfo",
				Template:     layerBaseReference + baseRowColReference + "/{I}/{J}.json",
			},
			wmts100.ResourceURL{
				Format:       "text/plain",
				ResourceType: "FeatureInfo",
				Template:     layerBaseReference + baseRowColReference + "/{I}/{J}.txt",
			})
	}
	return result
}

func getTileMatrixSetLinks(wmts *v2.WMTS) []wmts100.TileMatrixSetLink {
	result := make([]wmts100.TileMatrixSetLink, 0)
	for _, tileMatrixSet := range wmts.Spec.Service.TileMatrixSets {
		result = append(result, wmts100.TileMatrixSetLink{
			TileMatrixSet: tileMatrixSet.CRS,
		})
	}

	return result
}
