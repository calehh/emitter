package emitter

import (
	"context"
	"github.com/calehh/emitter/log"
	"github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"time"
)

type EventTracer struct {
	heightPersist    HeightPersist
	checkDuration    time.Duration
	maxRequestHeight int64
	waitSyncHeight   int64
}

type Config struct {
	CheckDuration    time.Duration
	MaxRequestHeight int64
	WaitSyncHeight   int64
}

func InitEventTracer(heightPersist HeightPersist, config Config) (*EventTracer, error) {
	return &EventTracer{
		heightPersist:    heightPersist,
		checkDuration:    config.CheckDuration,
		maxRequestHeight: config.MaxRequestHeight,
		waitSyncHeight:   config.WaitSyncHeight,
	}, nil
}

func (s *EventTracer) SubscribeChainEvent(ctx context.Context, chainInfo ChainInfo, ch chan Event) error {
	toHeight := uint64(0)
	timer := time.NewTimer(s.checkDuration)
	log.Info("subscribe chain events", "chainId", chainInfo.ChainID)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			timer.Reset(s.checkDuration)
			client, err := ethclient.Dial(chainInfo.RPC)
			if err != nil {
				log.Error("ethclient dial error", "err", err)
				continue
			}
			currentHeight, err := client.BlockNumber(context.Background())
			if err != nil {
				log.Error("client blocknumber error", "err", err)
				continue
			}
			currentHeight = currentHeight - uint64(s.waitSyncHeight)
			lastHeight, err := s.heightPersist.GetTraceHeight(chainInfo.ChainID)
			if err != nil {
				log.Error("GetLastEventeHeight error", "err", err)
				continue
			}
			fromHeight := lastHeight + 1
			if currentHeight < uint64(fromHeight) {
				//log.Info("check blocks", "current", currentHeight, "from", fromHeight)
				continue
			}
			if (currentHeight - uint64(fromHeight)) > uint64(s.maxRequestHeight) {
				toHeight = uint64(fromHeight) + uint64(s.maxRequestHeight)
			} else {
				toHeight = currentHeight
			}
			log.Info("check blocks", "from", fromHeight, "to", toHeight)
			events, err := getEvents(chainInfo, big.NewInt(fromHeight), big.NewInt(int64(toHeight)))
			if err != nil {
				log.Error("getActivityEvent error", err)
				continue
			}
			for _, event := range events {
				ch <- event
			}
			err = s.heightPersist.UpdateTraceHeight(chainInfo.ChainID, int64(toHeight))
			if err != nil {
				log.Warn("heightPersist UpdateTraceHeight err", err)
			}
		}
	}
}

func getEvents(chainInfo ChainInfo, fromBlock *big.Int, toBlock *big.Int) ([]Event, error) {
	contractAddrList := make([]ethcommon.Address, 0)
	for _, contractInfo := range chainInfo.FilterContract {
		contractAddrList = append(contractAddrList, contractInfo.Address)
	}
	topicList := make([]ethcommon.Hash, 0)
	for _, contract := range chainInfo.FilterContract {
		for _, topic := range contract.TopicList {
			topicList = append(topicList, topic.GetSignature())
		}
	}
	query := ethereum.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Addresses: contractAddrList,
		Topics:    [][]ethcommon.Hash{topicList},
	}
	client, err := ethclient.Dial(chainInfo.RPC)
	if err != nil {
		return nil, err
	}
	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Error("FilterLogs error", "err", err)
		return nil, err
	}
	eventList := make([]Event, 0)

	for _, vLog := range logs {
		if len(vLog.Topics) == 0 {
			log.Error("the length of Topics is less than 1")
			continue
		}
		found := false
		for _, contractInfo := range chainInfo.FilterContract {
			if found {
				break
			}
			if contractInfo.Address != vLog.Address {
				continue
			}
			for _, topic := range contractInfo.TopicList {
				if found {
					break
				}
				if vLog.Topics[0] == topic.GetSignature() {
					found = true
					inter, unpackError := topic.Unpack(vLog)
					if unpackError != nil {
						log.Error("UnpackIntoInterface error", unpackError)
						break
					}
					sender := ethcommon.Address{}
					if len(vLog.Topics) > 1 {
						sender = ethcommon.HexToAddress(vLog.Topics[1].Hex())
					}
					eventList = append(eventList, Event{
						topic.GetName(),
						vLog.TxHash.String(),
						contractInfo.Address,
						time.Now().Unix(),
						sender,
						vLog.BlockNumber,
						inter,
					})
				}
			}
		}
	}
	return eventList, nil
}
