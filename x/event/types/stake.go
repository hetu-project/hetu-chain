package types

// ValidatorStake represents a validator's stake on a given subnet.
// Amount is a string-encoded integer in Alpha's smallest unit (no decimals).
type ValidatorStake struct {
	Netuid    uint16 `json:"netuid"`
	Validator string `json:"validator"`
	Amount    string `json:"amount"`
}

// Delegation represents a staker's delegated stake to a validator on a subnet.
// Amount is a string-encoded integer in Alpha's smallest unit (no decimals).
type Delegation struct {
	Netuid    uint16 `json:"netuid"`
	Validator string `json:"validator"`
	Staker    string `json:"staker"`
	Amount    string `json:"amount"`
}
