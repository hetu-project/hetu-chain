package types

import (
	fmt "fmt"

	"github.com/ethereum/go-ethereum/common"
	erc20api "github.com/hetu-project/hetu-hub/v1/api/evmos/erc20/v1"
	protov2 "google.golang.org/protobuf/proto"
)

// GetERC20Signers gets the signer's address from the Ethereum tx signature
func GetERC20Signers(msg protov2.Message) ([][]byte, error) {
	msgConvERC20, ok := msg.(*erc20api.MsgConvertERC20)
	if !ok {
		return nil, fmt.Errorf("invalid type, expected MsgConvertERC20 and got %T", msg)
	}

	// The sender on the msg is a hex address
	sender := common.HexToAddress(msgConvERC20.Sender)

	return [][]byte{sender.Bytes()}, nil
}
