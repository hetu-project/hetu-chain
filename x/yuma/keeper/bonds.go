package keeper

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/hetu-project/hetu/v1/x/yuma/types"
)

// updateBonds 更新bonds矩阵 - Yuma共识的核心部分
func (k Keeper) updateBonds(ctx sdk.Context, netuid uint16, weights [][]sdk.Dec,
	active []bool, neurons []types.NeuronInfo) [][]sdk.Dec {

	n := len(neurons)
	bonds := make([][]sdk.Dec, n)
	params := k.GetParams(ctx)

	// 初始化bonds矩阵
	for i := 0; i < n; i++ {
		bonds[i] = make([]sdk.Dec, n)
		// 从存储中获取历史bonds数据
		if len(neurons[i].Bonds) > 0 && len(neurons[i].Bonds[0]) > 0 {
			for j, bondValue := range neurons[i].Bonds[0] {
				if j < n {
					bonds[i][j] = sdk.NewDec(int64(bondValue))
				}
			}
		}
	}

	// 计算当前epoch的激励变化
	incentiveVariance := k.calculateIncentiveVariance(ctx, netuid, neurons)

	// 计算alpha值
	alpha := k.calculateAlpha(params, incentiveVariance)

	// 更新bonds矩阵
	for i := 0; i < n; i++ {
		if !active[i] {
			continue
		}

		// 应用bonds更新公式: B_{ij}^{new} = α * W_{ij} + (1-α) * B_{ij}^{old}
		for j := 0; j < n; j++ {
			if active[j] {
				// 获取权重值
				weight := sdk.ZeroDec()
				if len(weights) > i && len(weights[i]) > j {
					weight = weights[i][j]
				}

				// 应用bonds更新公式
				oldBond := bonds[i][j]
				newBond := alpha.Mul(weight).Add(sdk.OneDec().Sub(alpha).Mul(oldBond))
				bonds[i][j] = newBond
			}
		}

		// 归一化bonds行
		bonds[i] = k.normalizeBondsRow(bonds[i], active)
	}

	// 应用bonds惩罚
	bonds = k.applyBondsPenalty(bonds, params.BondsPenalty, active)

	// 应用rho和kappa变换
	bonds = k.applyRhoKappa(bonds, params.Rho, params.Kappa, active)

	return bonds
}

// calculateAlpha 计算alpha值（liquid alpha机制）
func (k Keeper) calculateAlpha(params types.Params, incentiveVariance sdk.Dec) sdk.Dec {
	if !params.LiquidAlphaEnabled {
		// 如果liquid alpha未启用，使用固定alpha
		return params.AdjustmentAlpha.Quo(sdk.NewDec(65536)) // 归一化到[0,1]
	}

	// Liquid alpha: 根据激励变化动态调整alpha
	alphaHigh := params.AlphaHigh.Quo(sdk.NewDec(65536))
	alphaLow := params.AlphaLow.Quo(sdk.NewDec(65536))

	// 使用sigmoid函数将激励变化映射到alpha范围
	normalizedVariance := k.sigmoid(incentiveVariance)

	// alpha = alpha_low + (alpha_high - alpha_low) * sigmoid(variance)
	alphaRange := alphaHigh.Sub(alphaLow)
	alpha := alphaLow.Add(alphaRange.Mul(normalizedVariance))

	return alpha
}

// calculateIncentiveVariance 计算激励变化
func (k Keeper) calculateIncentiveVariance(ctx sdk.Context, netuid uint16, neurons []types.NeuronInfo) sdk.Dec {
	n := len(neurons)
	if n == 0 {
		return sdk.ZeroDec()
	}

	// 计算当前激励的方差
	var incentives []sdk.Dec
	totalIncentive := sdk.ZeroDec()

	for _, neuron := range neurons {
		if neuron.Active {
			incentives = append(incentives, neuron.Incentive)
			totalIncentive = totalIncentive.Add(neuron.Incentive)
		}
	}

	if len(incentives) == 0 {
		return sdk.ZeroDec()
	}

	// 计算平均值
	meanIncentive := totalIncentive.Quo(sdk.NewDec(int64(len(incentives))))

	// 计算方差
	variance := sdk.ZeroDec()
	for _, incentive := range incentives {
		diff := incentive.Sub(meanIncentive)
		variance = variance.Add(diff.Mul(diff))
	}

	variance = variance.Quo(sdk.NewDec(int64(len(incentives))))

	return variance
}

// sigmoid sigmoid激活函数
func (k Keeper) sigmoid(x sdk.Dec) sdk.Dec {
	// sigmoid(x) = 1 / (1 + exp(-x))
	// 为避免数值溢出，使用近似实现

	if x.GT(sdk.NewDec(10)) {
		return sdk.OneDec()
	}
	if x.LT(sdk.NewDec(-10)) {
		return sdk.ZeroDec()
	}

	// 使用泰勒展开近似exp(-x)
	negX := x.Neg()
	expNegX := k.exponential(negX)

	// sigmoid = 1 / (1 + exp(-x))
	denominator := sdk.OneDec().Add(expNegX)
	return sdk.OneDec().Quo(denominator)
}

// normalizeBondsRow 归一化bonds行
func (k Keeper) normalizeBondsRow(bondsRow []sdk.Dec, active []bool) []sdk.Dec {
	total := sdk.ZeroDec()
	n := len(bondsRow)

	// 计算总和（只考虑活跃神经元）
	for j := 0; j < n; j++ {
		if j < len(active) && active[j] {
			total = total.Add(bondsRow[j])
		}
	}

	// 归一化
	if total.GT(sdk.ZeroDec()) {
		for j := 0; j < n; j++ {
			if j < len(active) && active[j] {
				bondsRow[j] = bondsRow[j].Quo(total)
			} else {
				bondsRow[j] = sdk.ZeroDec()
			}
		}
	}

	return bondsRow
}

// exponential 计算指数函数的改进版本
func (k Keeper) exponential(x sdk.Dec) sdk.Dec {
	if x.IsZero() {
		return sdk.OneDec()
	}

	// 处理极值情况
	if x.GT(sdk.NewDec(20)) {
		return sdk.NewDec(int64(math.MaxInt64))
	}
	if x.LT(sdk.NewDec(-20)) {
		return sdk.ZeroDec()
	}

	// 使用泰勒展开: e^x = 1 + x + x^2/2! + x^3/3! + ...
	result := sdk.OneDec()
	term := sdk.OneDec()

	for i := 1; i <= 15; i++ { // 使用更多项以提高精度
		term = term.Mul(x).Quo(sdk.NewDec(int64(i)))
		result = result.Add(term)

		// 如果项变得非常小，提前退出
		if term.Abs().LT(sdk.NewDecWithPrec(1, 18)) {
			break
		}
	}

	return result
}

// applyBondsPenalty 应用bonds惩罚
func (k Keeper) applyBondsPenalty(bonds [][]sdk.Dec, penalty sdk.Dec, active []bool) [][]sdk.Dec {
	n := len(bonds)
	if n == 0 || penalty.IsZero() {
		return bonds
	}

	// 计算惩罚系数
	penaltyFactor := sdk.OneDec().Sub(penalty.Quo(sdk.NewDec(65536))) // 归一化

	for i := 0; i < n; i++ {
		if !active[i] {
			continue
		}

		for j := 0; j < len(bonds[i]); j++ {
			if j < len(active) && active[j] {
				// 应用惩罚: bond = bond * (1 - penalty_factor)
				bonds[i][j] = bonds[i][j].Mul(penaltyFactor)
			}
		}
	}

	return bonds
}

// applyRhoKappa 应用rho和kappa变换
func (k Keeper) applyRhoKappa(bonds [][]sdk.Dec, rho, kappa sdk.Dec, active []bool) [][]sdk.Dec {
	n := len(bonds)
	if n == 0 {
		return bonds
	}

	// Rho: 控制bonds的稀疏性
	// Kappa: 控制bonds的饱和性

	for i := 0; i < n; i++ {
		if !active[i] {
			continue
		}

		for j := 0; j < len(bonds[i]); j++ {
			if j < len(active) && active[j] {
				bond := bonds[i][j]

				// 应用rho变换: bond = bond^(1/rho)
				if rho.GT(sdk.ZeroDec()) && bond.GT(sdk.ZeroDec()) {
					// 简化实现: bond = bond * (1 + 1/rho)
					rhoFactor := sdk.OneDec().Add(sdk.OneDec().Quo(rho))
					bond = bond.Mul(rhoFactor)
				}

				// 应用kappa饱和: bond = min(bond, kappa)
				if kappa.GT(sdk.ZeroDec()) && bond.GT(kappa) {
					bond = kappa
				}

				bonds[i][j] = bond
			}
		}
	}

	return bonds
}
