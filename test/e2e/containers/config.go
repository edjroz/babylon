package containers

// ImageConfig contains all images and their respective tags
// needed for running e2e tests.
type ImageConfig struct {
	RelayerRepository string
	RelayerTag        string
}

//nolint:deadcode
const (
	// name of babylon container produced by running `make localnet-build-env`
	BabylonContainerName = "babylonchain/babylond"

	relayerRepository = "informalsystems/hermes"
	relayerTag        = "1.3.0"
)

// Returns ImageConfig needed for running e2e test.
// If isUpgrade is true, returns images for running the upgrade
// If isFork is true, utilizes provided fork height to initiate fork logic
func NewImageConfig() ImageConfig {
	config := ImageConfig{
		RelayerRepository: relayerRepository,
		RelayerTag:        relayerTag,
	}

	return config
}
