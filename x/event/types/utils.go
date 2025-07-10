package types

import (
	"math/big"
)

// AddBigIntString 将两个字符串形式的大整数相加
func AddBigIntString(a, b string) string {
	bigA, _ := new(big.Int).SetString(a, 10)
	bigB, _ := new(big.Int).SetString(b, 10)
	result := new(big.Int).Add(bigA, bigB)
	return result.String()
}

// SubBigIntString 将两个字符串形式的大整数相减
func SubBigIntString(a, b string) string {
	bigA, _ := new(big.Int).SetString(a, 10)
	bigB, _ := new(big.Int).SetString(b, 10)
	result := new(big.Int).Sub(bigA, bigB)
	// 确保结果不为负数
	if result.Sign() < 0 {
		return "0"
	}
	return result.String()
}
