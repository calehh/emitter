package emitter

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

type Topic interface {
	GetName() string
	GetSignature() ethcommon.Hash
	Unpack(log types.Log) (interface{}, error)
}

type AbiUnPacker interface {
	UnpackIntoInterface(v interface{}, name string, data []byte) error
}

type TopicExample struct {
	ContractAbi abi.ABI
}

type EventCarMint struct {
	Owner   ethcommon.Address
	TokenId *big.Int
	Level   *big.Int
	TotalMp *big.Int
	Network *big.Int
}

func (c TopicExample) GetName() string {
	return "MINT_CAR"
}

func (c TopicExample) GetSignature() ethcommon.Hash {
	return EventSignatureHash("MINT_CAR(address,uint256,uint256,uint256,uint256)")
}

func (c TopicExample) Unpack(log types.Log) (interface{}, error) {
	var eventCarMint EventCarMint
	err := c.ContractAbi.UnpackIntoInterface(&eventCarMint, "MINT_CAR", log.Data)
	if err != nil {
		return nil, err
	}
	return eventCarMint, nil
}
