// Copyright 2024 Hetu Project
// This file is part of the Hetu Network packages.

package types

import (
	"math/big"
)

type SubnetInfo struct {
	Netuid                uint16
	Owner                 string
	Name                  string
	Github                string
	Discord               string
	Website               string
	Description           string
	Kappa                 string
	BondsPenalty          string
	BondsMovingAverage    string
	AlphaLow              string
	AlphaHigh             string
	AlphaSigmoidSteepness string
	Params                map[string]string // 新增参数表
}

type SubnetParam struct {
	Netuid uint16
	Params map[string]string // 参数名到值
}

type StakeInfo struct {
	Netuid uint16
	Staker string
	Amount string
}

// 字符串大数加法
func AddBigIntString(a, b string) string {
	ai, _ := new(big.Int).SetString(a, 10)
	bi, _ := new(big.Int).SetString(b, 10)
	return new(big.Int).Add(ai, bi).String()
}

// 字符串大数减法
func SubBigIntString(a, b string) string {
	ai, _ := new(big.Int).SetString(a, 10)
	bi, _ := new(big.Int).SetString(b, 10)
	return new(big.Int).Sub(ai, bi).String()
}
