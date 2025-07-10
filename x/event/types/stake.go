package types

type ValidatorStake struct {
	Netuid    uint16 `json:"netuid"`
	Validator string `json:"validator"`
	Amount    string `json:"amount"`
}

type Delegation struct {
	Netuid    uint16 `json:"netuid"`
	Validator string `json:"validator"`
	Staker    string `json:"staker"`
	Amount    string `json:"amount"`
}
