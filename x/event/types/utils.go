package types

import (
	"math/big"
)

// AddBigIntString adds two base-10 big integer strings and returns the decimal string result.
// Invalid inputs are treated as "0" (not recommended for financial logic).
func AddBigIntString(a, b string) string {
	aInt, _ := new(big.Int).SetString(a, 10)
	bInt, _ := new(big.Int).SetString(b, 10)
	result := new(big.Int).Add(aInt, bInt)
	return result.String()
}

// SubBigIntString subtracts b from a for base-10 big integer strings and clamps negatives to "0".
// Invalid inputs are treated as "0" (not recommended for financial logic).
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
