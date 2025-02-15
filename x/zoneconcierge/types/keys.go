package types

const (
	// ModuleName defines the module name
	ModuleName = "zoneconcierge"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_zoneconcierge"

	// Version defines the current version the IBC module supports
	Version = "zoneconcierge-1"

	// PortID is the default port id that module binds to
	PortID = "zoneconcierge"
)

var (
	PortKey           = []byte{0x11} // PortKey defines the key to store the port ID in store
	ChainInfoKey      = []byte{0x12} // ChainInfoKey defines the key to store the chain info for each CZ in store
	CanonicalChainKey = []byte{0x13} // CanonicalChainKey defines the key to store the canonical chain for each CZ in store
	ForkKey           = []byte{0x14} // ForkKey defines the key to store the forks for each CZ in store
	EpochChainInfoKey = []byte{0x15} // EpochChainInfoKey defines the key to store each epoch's latests chain info for each CZ in store
	FinalizedEpochKey = []byte{0x16} // FinalizedEpochKey defines the key to store the last finalised epoch
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
