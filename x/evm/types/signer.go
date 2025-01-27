package types

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	evmapi "github.com/hetu-project/hetu-hub/v1/api/ethermint/evm/v1"
)

var (
	_ TxDataV2 = &evmapi.LegacyTx{}
	_ TxDataV2 = &evmapi.AccessListTx{}
	_ TxDataV2 = &evmapi.DynamicFeeTx{}
)

// supportedTxs holds the Ethereum transaction types
// supported by Evmos.
// Use a function to return a new pointer and avoid
// possible reuse or racing conditions when using the same pointer
var supportedTxs = map[string]func() TxDataV2{
	"/ethermint.evm.v1.DynamicFeeTx": func() TxDataV2 { return &evmapi.DynamicFeeTx{} },
	"/ethermint.evm.v1.AccessListTx": func() TxDataV2 { return &evmapi.AccessListTx{} },
	"/ethermint.evm.v1.LegacyTx":     func() TxDataV2 { return &evmapi.LegacyTx{} },
}

// getEVMSender extracts the sender address from the signature values using the latest signer for the given chainID.
func getEVMSender(txData TxDataV2) (common.Address, error) {
	signer := ethtypes.LatestSignerForChainID(txData.GetChainID())
	from, err := signer.Sender(ethtypes.NewTx(txData.AsEthereumData()))
	if err != nil {
		return common.Address{}, err
	}
	return from, nil
}

// GetEVMSigners is the custom function to get signers on Ethereum transactions
// Gets the signer's address from the Ethereum tx signature
func GetEVMSigners(msg protov2.Message) ([][]byte, error) {
	msgEthTx, ok := msg.(*evmapi.MsgEthereumTx)
	if !ok {
		return nil, fmt.Errorf("invalid type, expected MsgEthereumTx and got %T", msg)
	}

	txDataFn, found := supportedTxs[msgEthTx.Data.TypeUrl]
	if !found {
		return nil, fmt.Errorf("invalid TypeUrl %s", msgEthTx.Data.TypeUrl)
	}
	txData := txDataFn()

	// msgEthTx.Data is a message (DynamicFeeTx, LegacyTx or AccessListTx)
	if err := msgEthTx.Data.UnmarshalTo(txData); err != nil {
		return nil, err
	}

	sender, err := getEVMSender(txData)
	if err != nil {
		return nil, err
	}

	return [][]byte{sender.Bytes()}, nil
}

// TxDataV2 implements the Ethereum transaction tx structure. It is used
// solely to define the custom logic for getting signers on Ethereum transactions.
type TxDataV2 interface {
	GetChainID() *big.Int
	AsEthereumData() ethtypes.TxData

	ProtoReflect() protoreflect.Message
}