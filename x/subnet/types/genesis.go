package types

// DefaultGenesisState returns the default genesis state for the subnet module.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState() *GenesisState {
	return &GenesisState{}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	// No validation needed for this module's genesis state
	return nil
}

// GenesisState defines the subnet module's genesis state.
type GenesisState struct {
	// No genesis state for this module
}

// Reset implements proto.Message
func (gs *GenesisState) Reset() {
	*gs = GenesisState{}
}

// String implements proto.Message
func (gs *GenesisState) String() string {
	return "SubnetGenesisState{}"
}

// ProtoMessage implements proto.Message
func (*GenesisState) ProtoMessage() {}
