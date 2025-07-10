package types

type Subnet struct {
	Netuid     uint16            `json:"netuid"`
	Owner      string            `json:"owner"`
	LockAmount string            `json:"lock_amount"`
	BurnedTao  string            `json:"burned_tao"`
	Pool       string            `json:"pool"`
	Params     map[string]string `json:"params"`
}
