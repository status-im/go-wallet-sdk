package ethclient

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	mock_ethclient "github.com/status-im/go-wallet-sdk/pkg/ethclient/mock"
)

func TestWeb3Methods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRPC := mock_ethclient.NewMockRPCClient(ctrl)

	client := &Client{
		rpcClient: mockRPC,
	}

	// Test Web3ClientVersion
	clientVersionJSON := `"Geth/v1.10.26-stable-979fc968/linux-amd64/go1.19.2"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "web3_clientVersion").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(clientVersionJSON), result)
		})

	version, err := client.Web3ClientVersion(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "Geth/v1.10.26-stable-979fc968/linux-amd64/go1.19.2", version)

}
