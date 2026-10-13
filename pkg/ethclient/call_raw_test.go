package ethclient_test

import (
	"context"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"

	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
)

func newRecordingClient(t *testing.T, bodies *[]string) *ethclient.Client {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		*bodies = append(*bodies, string(body))
		var req struct {
			ID json.RawMessage `json:"id"`
		}
		_ = json.Unmarshal(body, &req)
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": "0x0102"})
	}))
	t.Cleanup(server.Close)
	rpcClient, err := rpc.Dial(server.URL)
	require.NoError(t, err)
	t.Cleanup(rpcClient.Close)
	return ethclient.NewClient(rpcClient)
}

func TestCallContractRaw_SendsTheSameRequestAsCallContract(t *testing.T) {
	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")
	data := []byte{0xde, 0xad, 0xbe, 0xef}
	callArg := json.RawMessage(`{"from":"0x1111111111111111111111111111111111111111","input":"0xdeadbeef","to":"0x2222222222222222222222222222222222222222"}`)

	for _, block := range []*big.Int{nil, big.NewInt(1234)} {
		var callBodies, rawBodies []string
		out, err := newRecordingClient(t, &callBodies).CallContract(context.Background(), ethereum.CallMsg{From: from, To: &to, Data: data}, block)
		require.NoError(t, err)
		rawOut, err := newRecordingClient(t, &rawBodies).CallContractRaw(context.Background(), callArg, block)
		require.NoError(t, err)

		require.Equal(t, callBodies, rawBodies)
		require.Equal(t, out, rawOut)
		require.Equal(t, []byte{1, 2}, rawOut)
	}
}
