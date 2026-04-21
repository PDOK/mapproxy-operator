package types //nolint:revive

type HashedConfigMapNames struct {
	CapabilitiesGenerator string
	KvpToRestful          string
	Mapproxy              string
}
type Images struct {
	ApacheExporterImage        string
	CapabilitiesGeneratorImage string
	KvpToRestfulImage          string
	MapproxyImage              string
	MultiToolImage             string
}
