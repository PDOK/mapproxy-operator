package constants

const (
	ApacheExporterName        = "apache-exporter"
	BlobDownloadName          = "blob-download"
	CapabilitiesGeneratorName = "capabilities-generator"
	KvpToRestfulName          = "wmts-kvp-to-restful"
	MapproxyName              = "mapproxy"

	MapserverPortNr int32 = 80
	ApachePortNr    int32 = 9117

	BaseVolumeName = "base"
	DataVolumeName = "data"

	configSuffix                             = "-config"
	ConfigMapCapabilitiesGeneratorVolumeName = CapabilitiesGeneratorName
	LighttpdVolumeName                       = "lighttpd"
	MapproxyVolumeName                       = "mapproxy"

	BlobsTilesBucket = "tiles"
)
