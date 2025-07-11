package keeper

import (
	"testing"

	"github.com/hetu-project/hetu/v1/x/yuma/types"
	"github.com/stretchr/testify/require"
)

func TestClamp(t *testing.T) {
	k := Keeper{}

	// 测试正常值
	require.Equal(t, 0.5, k.clamp(0.5, 0.0, 1.0))

	// 测试下限
	require.Equal(t, 0.0, k.clamp(-0.1, 0.0, 1.0))

	// 测试上限
	require.Equal(t, 1.0, k.clamp(1.5, 0.0, 1.0))
}

func TestNormalize(t *testing.T) {
	k := Keeper{}

	values := []float64{1.0, 2.0, 3.0}
	normalized := k.normalize(values)

	// 检查总和为1
	sum := 0.0
	for _, v := range normalized {
		sum += v
	}
	require.InDelta(t, 1.0, sum, 0.001)

	// 检查比例正确
	require.InDelta(t, 1.0/6.0, normalized[0], 0.001)
	require.InDelta(t, 2.0/6.0, normalized[1], 0.001)
	require.InDelta(t, 3.0/6.0, normalized[2], 0.001)
}

func TestMatMul(t *testing.T) {
	k := Keeper{}

	weights := [][]float64{
		{1.0, 2.0},
		{3.0, 4.0},
	}
	stake := []float64{0.5, 0.5}

	result := k.matMul(weights, stake)

	require.Len(t, result, 2)
	require.InDelta(t, 2.0, result[0], 0.001) // 1.0*0.5 + 3.0*0.5
	require.InDelta(t, 3.0, result[1], 0.001) // 2.0*0.5 + 4.0*0.5
}

func TestComputeDisabledLiquidAlpha(t *testing.T) {
	k := Keeper{}

	alpha := k.computeDisabledLiquidAlpha(0.9)
	require.InDelta(t, 0.1, alpha, 0.001)
}

func TestAlphaSigmoid(t *testing.T) {
	k := Keeper{}

	params := types.EpochParams{
		AlphaSigmoidSteepness: 10.0,
		AlphaLow:              0.01,
		AlphaHigh:             0.99,
	}

	// 测试 sigmoid 计算
	alpha := k.alphaSigmoid(0.5, 0.6, 0.4, params)
	require.Greater(t, alpha, params.AlphaLow)
	require.Less(t, alpha, params.AlphaHigh)
}
