package types

import (
	"math/big"
)

// AddBigIntString adds two big integer strings
func AddBigIntString(a, b string) string {
	aInt, _ := new(big.Int).SetString(a, 10)
	bInt, _ := new(big.Int).SetString(b, 10)
	result := new(big.Int).Add(aInt, bInt)
	return result.String()
}

// SubBigIntString subtracts two big integer strings
func SubBigIntString(a, b string) string {
	aInt, _ := new(big.Int).SetString(a, 10)
	bInt, _ := new(big.Int).SetString(b, 10)
	// Ensure the result is not negative
	if aInt.Cmp(bInt) < 0 {
		return "0"
	}
	result := new(big.Int).Sub(aInt, bInt)
	return result.String()
}
