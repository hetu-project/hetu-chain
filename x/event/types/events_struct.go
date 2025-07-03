// Copyright 2024 Hetu Project
// This file is part of the Hetu Network packages.

package types

type SubnetRegisteredEvent struct {
	Netuid                uint16
	Owner                 string
	Name                  string
	Github                string
	Discord               string
	Website               string
	Description           string
	Kappa                 string // 用 string 存储大数
	BondsPenalty          string
	BondsMovingAverage    string
	AlphaLow              string
	AlphaHigh             string
	AlphaSigmoidSteepness string
	TxHash                string
	BlockNumber           uint64
}

type SubnetMultiParamUpdatedEvent struct {
	Netuid      uint16
	Params      []string
	Values      []string
	TxHash      string
	BlockNumber uint64
}

type TaoStakedEvent struct {
	Netuid      uint16
	Staker      string
	Amount      string
	TxHash      string
	BlockNumber uint64
}

type TaoUnstakedEvent struct {
	Netuid      uint16
	Staker      string
	Amount      string
	TxHash      string
	BlockNumber uint64
}
