package types

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	// 验证总供应量：2100万 HETU = 21,000,000 * 10^18 aHETU
	expectedTotalSupply, _ := math.NewIntFromString("21000000000000000000000000")
	require.Equal(t, expectedTotalSupply, params.TotalSupply, "总供应量应该是2100万 HETU")

	// 验证默认区块奖励：1 HETU = 10^18 aHETU
	expectedBlockEmission, _ := math.NewIntFromString("1000000000000000000")
	require.Equal(t, expectedBlockEmission, params.DefaultBlockEmission, "默认区块奖励应该是1 HETU")

	// 验证其他参数
	require.Equal(t, "ahetu", params.MintDenom, "代币单位应该是 ahetu")
	require.True(t, params.EnableBlockInflation, "区块通胀应该默认启用")

	t.Logf("总供应量: %s aHETU", params.TotalSupply.String())
	t.Logf("默认区块奖励: %s aHETU", params.DefaultBlockEmission.String())

	// 验证数值的正确性
	require.Equal(t, "21000000000000000000000000", params.TotalSupply.String(), "总供应量字符串表示")
	require.Equal(t, "1000000000000000000", params.DefaultBlockEmission.String(), "默认区块奖励字符串表示")
}
