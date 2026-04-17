package types

type HashedConfigMapNames struct {
	CapabilitiesGenerator string
	Mapproxy              string
	KvpToRestful          string
}
type Images struct {
	MultiToolImage             string
	CapabilitiesGeneratorImage string
	KvpToRestfulImage          string
	MapproxyImage              string
	ApacheExporterImage        string
}
