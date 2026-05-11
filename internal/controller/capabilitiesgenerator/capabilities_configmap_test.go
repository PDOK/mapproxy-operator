package capabilitiesgenerator

import (
	"testing"

	"github.com/pdok/mapproxy-operator/internal/controller/utils"
	config2 "github.com/pdok/ogc-capabilities-generator/pkg/config"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestCache(t *testing.T) {
	wmts := utils.Cache
	actualConfig, err := MapWMTSToCapabilitiesGeneratorInput(&wmts)
	assert.NoError(t, err)
	var expectedConfig config2.Config
	err = yaml.Unmarshal([]byte(expectedConfigStringCache), &expectedConfig)
	assert.NoError(t, err)
	assert.Equal(t, *expectedConfig.Services.WMTS100Config, *actualConfig.Services.WMTS100Config)
}

func TestNoCache(t *testing.T) {
	wmts := utils.NoCache
	actualConfig, err := MapWMTSToCapabilitiesGeneratorInput(&wmts)
	assert.NoError(t, err)
	var expectedConfig config2.Config
	err = yaml.Unmarshal([]byte(expectedConfigStringNoCache), &expectedConfig)
	assert.NoError(t, err)
	assert.Equal(t, *expectedConfig.Services.WMTS100Config, *actualConfig.Services.WMTS100Config)
}

func TestFeatureInfo(t *testing.T) {
	wmts := utils.FeatureInfo
	actualConfig, err := MapWMTSToCapabilitiesGeneratorInput(&wmts)
	assert.NoError(t, err)
	var expectedConfig config2.Config
	err = yaml.Unmarshal([]byte(expectedConfigStringFeatureInfo), &expectedConfig)
	assert.NoError(t, err)
	assert.Equal(t, *expectedConfig.Services.WMTS100Config, *actualConfig.Services.WMTS100Config)
}

var expectedConfigStringCache = `
global:
    prefix: wmts
    namespace: ""
    onlineResourceUrl: ""
    path: ""
    version: ""
    additionalSchemaLocations: ""
services:
    wmts100:
        filename: /output/WMTSCapabilities.xml
        definition:
            capabilities:
                space: ""
                local: ""
            namespaces:
                xmlns: ""
                common: ""
                xlink: ""
                xsi: ""
                gml: ""
                version: ""
                schemaLocation: ""
            serviceIdentification:
                title: My service title
                abstract: My service abstract
                keywords: null
                serviceType: ""
                serviceTypeVersion: ""
                fees: none
                accessConstraints: none
            serviceProvider: null
            operationsMetadata:
                operationsMetadata:
                    space: ""
                    local: ""
                operation:
                    - name: GetCapabilities
                      dcp:
                        http:
                            get:
                                type: simple
                                href: https://test.example.com/owner/cache/wmts/v1_0?
                                constraint:
                                    - name: GetEncoding
                                      allowedValues:
                                        value:
                                            - KVP
                    - name: GetTile
                      dcp:
                        http:
                            get:
                                type: simple
                                href: https://test.example.com/owner/cache/wmts/v1_0?
                                constraint:
                                    - name: GetEncoding
                                      allowedValues:
                                        value:
                                            - KVP
            contents:
                layer:
                    - title: My layer title
                      abstract: My layer abstract
                      wgs84BoundingBox:
                        lowerCorner: -1.65729160235 48.0405018704
                        upperCorner: 12.4317272654 56.1105896442
                      identifier: layeridentifier
                      metadata: null
                      style:
                        - title: null
                          abstract: null
                          keywords: null
                          identifier: default
                          legendUrl:
                            - format: image/png
                              href: https://test.example.com/owner/cache/wmts/v1_0/layeridentifier/legend.png
                          isDefault: null
                      format:
                        - image/png
                      tileMatrixSetLink:
                        - tileMatrixSet: EPSG:28992
                        - tileMatrixSet: EPSG:28992
                      resourceUrl:
                        - format: image/png
                          resourceType: tile
                          template: https://test.example.com/owner/cache/wmts/v1_0/layeridentifier/{TileMatrixSet}/{TileMatrix}/{TileCol}/{TileRow}.png
                tileMatrixSet:
                    - identifier: EPSG:28992
                      supportedCrs: ""
                      tileMatrix:
                        - identifier: "00"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "01"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "02"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "03"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "04"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "05"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "06"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "07"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "08"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "09"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "10"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "11"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "12"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                    - identifier: EPSG:28992
                      supportedCrs: ""
                      tileMatrix:
                        - identifier: "00"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "01"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "02"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "03"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "04"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "05"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "06"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "07"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "08"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "09"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "10"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "11"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "12"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "13"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "14"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "15"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
            serviceMetadataUrl:
                href: https://test.example.com/owner/cache/wmts/v1_0/WMTSCapabilities.xml

`
var expectedConfigStringNoCache = `
global:
    prefix: wmts
    namespace: ""
    onlineResourceUrl: ""
    path: ""
    version: ""
    additionalSchemaLocations: ""
services:
    wmts100:
        filename: /output/WMTSCapabilities.xml
        definition:
            capabilities:
                space: ""
                local: ""
            namespaces:
                xmlns: ""
                common: ""
                xlink: ""
                xsi: ""
                gml: ""
                version: ""
                schemaLocation: ""
            serviceIdentification:
                title: My service title
                abstract: My service abstract
                keywords: null
                serviceType: ""
                serviceTypeVersion: ""
                fees: none
                accessConstraints: none
            serviceProvider: null
            operationsMetadata:
                operationsMetadata:
                    space: ""
                    local: ""
                operation:
                    - name: GetCapabilities
                      dcp:
                        http:
                            get:
                                type: simple
                                href: https://test.example.com/owner/nocache/wmts/v1_0?
                                constraint:
                                    - name: GetEncoding
                                      allowedValues:
                                        value:
                                            - KVP
                    - name: GetTile
                      dcp:
                        http:
                            get:
                                type: simple
                                href: https://test.example.com/owner/nocache/wmts/v1_0?
                                constraint:
                                    - name: GetEncoding
                                      allowedValues:
                                        value:
                                            - KVP
            contents:
                layer:
                    - title: My layer title
                      abstract: My layer abstract
                      wgs84BoundingBox:
                        lowerCorner: -1.65729160235 48.0405018704
                        upperCorner: 12.4317272654 56.1105896442
                      identifier: layeridentifier
                      metadata: null
                      style:
                        - title: null
                          abstract: null
                          keywords: null
                          identifier: default
                          legendUrl:
                            - format: image/png
                              href: https://test.example.com/owner/nocache/wmts/v1_0/layeridentifier/legend.png
                          isDefault: null
                      format:
                        - image/png
                      tileMatrixSetLink:
                        - tileMatrixSet: EPSG:28992
                        - tileMatrixSet: EPSG:28992
                      resourceUrl:
                        - format: image/png
                          resourceType: tile
                          template: https://test.example.com/owner/nocache/wmts/v1_0/layeridentifier/{TileMatrixSet}/{TileMatrix}/{TileCol}/{TileRow}.png
                tileMatrixSet:
                    - identifier: EPSG:28992
                      supportedCrs: ""
                      tileMatrix:
                        - identifier: "00"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "01"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "02"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "03"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "04"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "05"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "06"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "07"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "08"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "09"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "10"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "11"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "12"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                    - identifier: EPSG:28992
                      supportedCrs: ""
                      tileMatrix:
                        - identifier: "00"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "01"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "02"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "03"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "04"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "05"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "06"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "07"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "08"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "09"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "10"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "11"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "12"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "13"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "14"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "15"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
            serviceMetadataUrl:
                href: https://test.example.com/owner/nocache/wmts/v1_0/WMTSCapabilities.xml
`

var expectedConfigStringFeatureInfo = `
global:
    prefix: wmts
    namespace: ""
    onlineResourceUrl: ""
    path: ""
    version: ""
    additionalSchemaLocations: ""
services:
    wmts100:
        filename: /output/WMTSCapabilities.xml
        definition:
            capabilities:
                space: ""
                local: ""
            namespaces:
                xmlns: ""
                common: ""
                xlink: ""
                xsi: ""
                gml: ""
                version: ""
                schemaLocation: ""
            serviceIdentification:
                title: My service title
                abstract: My service abstract
                keywords: null
                serviceType: ""
                serviceTypeVersion: ""
                fees: none
                accessConstraints: none
            serviceProvider: null
            operationsMetadata:
                operationsMetadata:
                    space: ""
                    local: ""
                operation:
                    - name: GetCapabilities
                      dcp:
                        http:
                            get:
                                type: simple
                                href: https://test.example.com/owner/featureinfo/wmts/v1_0?
                                constraint:
                                    - name: GetEncoding
                                      allowedValues:
                                        value:
                                            - KVP
                    - name: GetTile
                      dcp:
                        http:
                            get:
                                type: simple
                                href: https://test.example.com/owner/featureinfo/wmts/v1_0?
                                constraint:
                                    - name: GetEncoding
                                      allowedValues:
                                        value:
                                            - KVP
                    - name: GetFeatureInfo
                      dcp:
                        http:
                            get:
                                type: simple
                                href: https://test.example.com/owner/featureinfo/wmts/v1_0?
                                constraint:
                                    - name: GetEncoding
                                      allowedValues:
                                        value:
                                            - KVP
            contents:
                layer:
                    - title: My layer title
                      abstract: My layer abstract
                      wgs84BoundingBox:
                        lowerCorner: -1.65729160235 48.0405018704
                        upperCorner: 12.4317272654 56.1105896442
                      identifier: layeridentifier
                      metadata: null
                      style:
                        - title: null
                          abstract: null
                          keywords: null
                          identifier: default
                          legendUrl:
                            - format: image/png
                              href: https://test.example.com/owner/featureinfo/wmts/v1_0/layeridentifier/legend.png
                          isDefault: null
                      format:
                        - image/png
                      infoFormat:
                        - text/html
                        - text/xml
                        - application/json
                        - text/plain
                      tileMatrixSetLink:
                        - tileMatrixSet: EPSG:28992
                        - tileMatrixSet: EPSG:28992
                      resourceUrl:
                        - format: image/png
                          resourceType: tile
                          template: https://test.example.com/owner/featureinfo/wmts/v1_0/layeridentifier/{TileMatrixSet}/{TileMatrix}/{TileCol}/{TileRow}.png
                        - format: text/html
                          resourceType: FeatureInfo
                          template: https://test.example.com/owner/featureinfo/wmts/v1_0/layeridentifier/{TileMatrixSet}/{TileMatrix}/{TileCol}/{TileRow}/{I}/{J}.html
                        - format: text/xml
                          resourceType: FeatureInfo
                          template: https://test.example.com/owner/featureinfo/wmts/v1_0/layeridentifier/{TileMatrixSet}/{TileMatrix}/{TileCol}/{TileRow}/{I}/{J}.xml
                        - format: application/json
                          resourceType: FeatureInfo
                          template: https://test.example.com/owner/featureinfo/wmts/v1_0/layeridentifier/{TileMatrixSet}/{TileMatrix}/{TileCol}/{TileRow}/{I}/{J}.json
                        - format: text/plain
                          resourceType: FeatureInfo
                          template: https://test.example.com/owner/featureinfo/wmts/v1_0/layeridentifier/{TileMatrixSet}/{TileMatrix}/{TileCol}/{TileRow}/{I}/{J}.txt
                tileMatrixSet:
                    - identifier: EPSG:28992
                      supportedCrs: ""
                      tileMatrix:
                        - identifier: "00"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "01"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "02"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "03"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "04"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "05"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "06"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "07"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "08"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "09"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "10"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "11"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "12"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                    - identifier: EPSG:28992
                      supportedCrs: ""
                      tileMatrix:
                        - identifier: "00"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "01"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "02"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "03"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "04"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "05"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "06"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "07"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "08"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "09"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "10"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "11"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "12"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "13"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "14"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
                        - identifier: "15"
                          scaleDenominator: ""
                          topLeftCorner: ""
                          tileWidth: ""
                          tileHeight: ""
                          matrixWidth: ""
                          matrixHeight: ""
            serviceMetadataUrl:
                href: https://test.example.com/owner/featureinfo/wmts/v1_0/WMTSCapabilities.xml
`
