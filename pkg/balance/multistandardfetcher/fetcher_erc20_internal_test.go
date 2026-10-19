package multistandardfetcher

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3"
)

func TestERC20Job_OmittedZeroBalanceDoesNotAllocate(t *testing.T) {
	job := buildERC20Job(common.Address{1}, []ContractAddress{{2}}, true)
	zero := multicall3.IMulticall3Result{Success: true, ReturnData: make([]byte, 32)}
	allocs := testing.AllocsPerRun(100, func() {
		_, _ = job.CallResultFn(zero)
	})
	assert.Zero(t, allocs)
}
