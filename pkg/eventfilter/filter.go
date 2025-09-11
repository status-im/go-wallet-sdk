package eventfilter

import (
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/status-im/go-wallet-sdk/pkg/eventlog"
)

type FilterClient interface {
	FilterLogs(context.Context, ethereum.FilterQuery) ([]types.Log, error)
}

func FilterTransfers(client FilterClient, config TransferQueryConfig) ([]eventlog.Event, error) {
	queries := config.ToFilterQueries()
	events := make([]eventlog.Event, 0)

	for _, query := range queries {
		logs, err := client.FilterLogs(context.Background(), query)
		if err != nil {
			return nil, err
		}
		for _, log := range logs {
			event := eventlog.ParseLog(log)
			events = append(events, event...)
		}
	}

	return events, nil
}
