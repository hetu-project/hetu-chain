package types

// Subnet represents a subnet in the network
type Subnet struct {
	Netuid             uint16            `json:"netuid" yaml:"netuid"`
	Owner              string            `json:"owner" yaml:"owner"`
	LockAmount         string            `json:"lock_amount" yaml:"lock_amount"`
	BurnedTao          string            `json:"burned_tao" yaml:"burned_tao"`
	Pool               string            `json:"pool" yaml:"pool"`
	Params             map[string]string `json:"params" yaml:"params"`
	FirstEmissionBlock uint64            `json:"first_emission_block" yaml:"first_emission_block"` // 首次排放区块号
}
