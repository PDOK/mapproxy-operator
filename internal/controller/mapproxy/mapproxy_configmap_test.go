package mapproxy

import (
	"testing"

	"github.com/pdok/mapproxy-operator/internal/controller/utils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestCache(t *testing.T) {
	wmts := utils.Cache
	mapproxyConfig, err := GetMapproxyConfig(&wmts)
	assert.NoError(t, err)

	var expectedConfig MapproxyConfig
	err = yaml.Unmarshal([]byte(expectedConfigStringCache), &expectedConfig)
	assert.NoError(t, err)
	assert.True(t, mapproxyConfig.Equal(&expectedConfig))
}

func TestNoCache(t *testing.T) {
	wmts := utils.NoCache
	mapproxyConfig, err := GetMapproxyConfig(&wmts)
	assert.NoError(t, err)

	var expectedConfig MapproxyConfig
	err = yaml.Unmarshal([]byte(expectedConfigStringNoCache), &expectedConfig)
	assert.NoError(t, err)
	assert.True(t, mapproxyConfig.Equal(&expectedConfig))
}

func TestFeatureInfo(t *testing.T) {
	wmts := utils.FeatureInfo
	mapproxyConfig, err := GetMapproxyConfig(&wmts)
	assert.NoError(t, err)

	var expectedConfig MapproxyConfig
	err = yaml.Unmarshal([]byte(expectedConfigStringFeatureInfo), &expectedConfig)
	assert.NoError(t, err)
	assert.True(t, mapproxyConfig.Equal(&expectedConfig))
}

var expectedConfigStringCache = `
services:
    wmts:
        kvp: true
        restful: true
layers:
    - name: layeridentifier
      title: My layer title
      tile_sources:
        - layeridentifier-28992-cache
        - layeridentifier-28992-cache
caches:
    layeridentifier-28992-cache:
        sources:
            - layeridentifier-28992-source
        grids:
            - EPSG:28992
        disable_storage: false
        cache:
            type: azureblob
            directory: owner/layeridentifier/28992/
            container_name: tiles
sources:
    layeridentifier-28992-source:
        type: wms
        wms_opts:
            map: true
            featureinfo: false
            version: 1.3.0
        supported_srs:
            - EPSG:28992
        coverage:
            srs: EPSG:28992
            bbox:
                - -101552
                - 210360
                - 352820
                - 887600
        min_res: 0.42
        max_res: 0.0525
        req:
            layers: layer1identifier
            url: https://my-wms-mapserver/mapserver?MAP_RESOLUTION=92
            transparent: true
grids:
    EPSG:28992:
        tile_size:
            - 256
            - 256
        origin: nw
        srs: EPSG:28992
        bbox:
            - -285401.92
            - 22598.08
            - 595401.92
            - 903401.92
        bbox_srs: EPSG:28992
        res:
            - 3440.64
            - 1720.32
            - 860.16
            - 430.08
            - 215.04
            - 107.52
            - 53.76
            - 26.88
            - 13.44
            - 6.72
            - 3.36
            - 1.68
            - 0.84
            - 0.42
            - 0.21
            - 0.105
            - 0.0525
globals:
    cache:
        meta_buffer: 360
        base_dir: /srv/mapproxy/cache_data
        lock_dir: /srv/mapproxy/cache_data/locks
        tile_lock_dir: /srv/mapproxy/cache_data/tile_locks
        meta_size:
            - 8
            - 8
    image:
        resampling_method: bilinear
        paletted: false
        formats:
            png24:
                format: image/png
                transparent: true
`
var expectedConfigStringNoCache = `services:
    wmts:
        kvp: true
        restful: true
layers:
    - name: layeridentifier
      title: My layer title
      tile_sources:
        - layeridentifier-28992-cache
        - layeridentifier-28992-cache
caches:
    layeridentifier-28992-cache:
        sources:
            - layeridentifier-28992-source
        grids:
            - EPSG:28992
        disable_storage: true
sources:
    layeridentifier-28992-source:
        type: wms
        wms_opts:
            map: true
            featureinfo: false
            version: 1.3.0
        supported_srs:
            - EPSG:28992
        coverage:
            srs: EPSG:28992
            bbox:
                - -101552
                - 210360
                - 352820
                - 887600
        min_res: 0.42
        max_res: 0.0525
        req:
            layers: layer1identifier
            url: https://my-wms-mapserver/mapserver?MAP_RESOLUTION=92
            transparent: true
grids:
    EPSG:28992:
        tile_size:
            - 256
            - 256
        origin: nw
        srs: EPSG:28992
        bbox:
            - -285401.92
            - 22598.08
            - 595401.92
            - 903401.92
        bbox_srs: EPSG:28992
        res:
            - 3440.64
            - 1720.32
            - 860.16
            - 430.08
            - 215.04
            - 107.52
            - 53.76
            - 26.88
            - 13.44
            - 6.72
            - 3.36
            - 1.68
            - 0.84
            - 0.42
            - 0.21
            - 0.105
            - 0.0525
globals:
    cache:
        meta_buffer: 360
        meta_size:
            - 1
            - 1
    image:
        resampling_method: bilinear
        paletted: false
        formats:
            png24:
                format: image/png
                transparent: true
`
var expectedConfigStringFeatureInfo = `services:
    wmts:
        kvp: true
        restful: true
        featureinfo_formats:
            - mimetype: text/html
              suffix: html
            - mimetype: text/xml
              suffix: xml
            - mimetype: application/json
              suffix: json
            - mimetype: text/plain
              suffix: txt
layers:
    - name: layeridentifier
      title: My layer title
      tile_sources:
        - layeridentifier-28992-cache
        - layeridentifier-28992-cache
caches:
    layeridentifier-28992-cache:
        sources:
            - layeridentifier-28992-source
            - layeridentifier-28992-source-featureinfo
        grids:
            - EPSG:28992
        disable_storage: false
        cache:
            type: azureblob
            directory: owner/layeridentifier/28992/
            container_name: tiles
sources:
    layeridentifier-28992-source:
        type: wms
        wms_opts:
            map: true
            featureinfo: false
            version: 1.3.0
        supported_srs:
            - EPSG:28992
        coverage:
            srs: EPSG:28992
            bbox:
                - -101552
                - 210360
                - 352820
                - 887600
        min_res: 0.42
        max_res: 0.0525
        req:
            layers: layer1identifier
            url: https://my-wms-mapserver/mapserver?MAP_RESOLUTION=92
            transparent: true
    layeridentifier-28992-source-featureinfo:
        type: wms
        wms_opts:
            map: false
            featureinfo: true
            version: 1.3.0
        supported_srs:
            - EPSG:28992
        coverage:
            srs: EPSG:28992
            bbox:
                - -101552
                - 210360
                - 352820
                - 887600
        min_res: 0.42
        max_res: 0.0525
        req:
            layers: layer1identifier
            url: https://my-wms-mapserver/mapserver?MAP_RESOLUTION=92&feature_count=5
            transparent: true
grids:
    EPSG:28992:
        tile_size:
            - 256
            - 256
        origin: nw
        srs: EPSG:28992
        bbox:
            - -285401.92
            - 22598.08
            - 595401.92
            - 903401.92
        bbox_srs: EPSG:28992
        res:
            - 3440.64
            - 1720.32
            - 860.16
            - 430.08
            - 215.04
            - 107.52
            - 53.76
            - 26.88
            - 13.44
            - 6.72
            - 3.36
            - 1.68
            - 0.84
            - 0.42
            - 0.21
            - 0.105
            - 0.0525
globals:
    cache:
        meta_buffer: 360
        base_dir: /srv/mapproxy/cache_data
        lock_dir: /srv/mapproxy/cache_data/locks
        tile_lock_dir: /srv/mapproxy/cache_data/tile_locks
        meta_size:
            - 8
            - 8
    image:
        resampling_method: bilinear
        paletted: false
        formats:
            png24:
                format: image/png
                transparent: true`
