package constants

const (
	ApacheExporterName        = "apache-exporter"
	BlobDownloadName          = "blob-download"
	CapabilitiesGeneratorName = "capabilities-generator"
	KvpToRestfulName          = "wmts-kvp-to-restful"
	MapproxyName              = "mapproxy"

	MapserverPortNr    int32 = 80
	ApachePortNr       int32 = 9117
	MapproxyPortNumber int32 = 9001

	BaseVolumeName = "base"
	DataVolumeName = "data"

	ConfigMapCapabilitiesGeneratorVolumeName = CapabilitiesGeneratorName
	LighttpdVolumeName                       = "lighttpd"
	MapproxyVolumeName                       = "mapproxy"

	BlobsTilesBucket = "tiles"

	LighttpdPortName     = "lighttpd"
	ApacheExportPortName = ApacheExporterName
	KvpToRestfulPortName = KvpToRestfulName
)
