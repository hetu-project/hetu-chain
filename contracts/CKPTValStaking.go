package contracts

import (
	_ "embed"
	"encoding/json"

	evmtypes "github.com/hetu-project/hetu/v1/x/evm/types"
)

var (
	//go:embed compiled_contracts/CKPTValStaking.json
	CKPTValStakingABI []byte //nolint: golint

	// CKPTValStakingContract is the compiled CKPTValStaking contract
	CKPTValStakingContract evmtypes.CompiledContract
)

func init() {

	err := json.Unmarshal(CKPTValStakingABI, &CKPTValStakingContract)
	if err != nil {
		panic(err)
	}

	if len(CKPTValStakingContract.Bin) == 0 {
		panic("load contract failed")
	}
}
