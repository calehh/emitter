package emitter

import (
	ethcommon "github.com/ethereum/go-ethereum/common"
)

type ChainInfo struct {
	RPC            string
	ChainID        int64
	FilterContract []ContractInfo
}

type ContractInfo struct {
	Address   ethcommon.Address
	TopicList []Topic
}

type Event struct {
	Type        string
	TxHash      string
	Contract    ethcommon.Address
	Timestamps  int64
	Sender      ethcommon.Address
	BlockHeight uint64
	Data        interface{}
}

type HeightPersist interface {
	GetTraceHeight(chainId int64) (int64, error)
	UpdateTraceHeight(chainId int64, height int64) error
}
