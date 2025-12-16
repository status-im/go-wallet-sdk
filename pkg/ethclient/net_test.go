package ethclient_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
	mock_ethclient "github.com/status-im/go-wallet-sdk/pkg/ethclient/mock"
)

func TestNetMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRPC := mock_ethclient.NewMockRPCClient(ctrl)

	client := ethclient.NewClient(mockRPC)

	// Test NetVersion
	netVersionJSON := `"1"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "net_version").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(netVersionJSON), result)
		})

	version, err := client.NetVersion(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "1", version)

	// Test NetListening
	listeningJSON := `true`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "net_listening").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(listeningJSON), result)
		})

	listening, err := client.NetListening(context.Background())
	assert.NoError(t, err)
	assert.True(t, listening)

	// Test NetPeerCount
	peerCountJSON := `"0x2"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "net_peerCount").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(peerCountJSON), result)
		})

	peerCount, err := client.NetPeerCount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), peerCount)
}
