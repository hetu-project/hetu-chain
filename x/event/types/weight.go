package types

type ValidatorWeight struct {
	Netuid    uint16            `json:"netuid"`
	Validator string            `json:"validator"`
	Weights   map[string]uint64 `json:"weights"`
}
