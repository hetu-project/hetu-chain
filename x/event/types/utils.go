package types

import (
	"fmt"
	"math/big"
)

// AddBigIntString adds two base-10 big integer strings and returns the decimal string result.
// Invalid inputs are treated as "0" (not recommended for financial logic).
func AddBigIntString(a, b string) string {
	aInt := new(big.Int)
	if _, ok := aInt.SetString(a, 10); !ok {
		aInt.SetInt64(0)
	}
	bInt := new(big.Int)
	if _, ok := bInt.SetString(b, 10); !ok {
		bInt.SetInt64(0)
	}
	return new(big.Int).Add(aInt, bInt).String()
}

// AddBigIntStringWithError adds two base-10 big integer strings and returns the sum as a decimal string.
// Returns an error if either input cannot be parsed.
func AddBigIntStringWithError(a, b string) (string, error) {
	aInt := new(big.Int)
	if _, ok := aInt.SetString(a, 10); !ok {
		return "", fmt.Errorf("invalid big integer string a: %q", a)
	}
	bInt := new(big.Int)
	if _, ok := bInt.SetString(b, 10); !ok {
		return "", fmt.Errorf("invalid big integer string b: %q", b)
	}
	return new(big.Int).Add(aInt, bInt).String(), nil
}

// SubBigIntString subtracts b from a for base-10 big integer strings and clamps negatives to "0".
// Invalid inputs are treated as "0" (not recommended for financial logic).
func SubBigIntString(a, b string) string {
	aInt := new(big.Int)
	if _, ok := aInt.SetString(a, 10); !ok {
		aInt.SetInt64(0)
	}
	bInt := new(big.Int)
	if _, ok := bInt.SetString(b, 10); !ok {
		bInt.SetInt64(0)
	}
	// Ensure the result is not negative
	if aInt.Cmp(bInt) < 0 {
		return "0"
	}
	return new(big.Int).Sub(aInt, bInt).String()
}

// SubBigIntStringWithError subtracts b from a, where a and b are base-10 big-integer strings.
// If a < b, the function clamps the result to "0". Returns an error if parsing fails.
func SubBigIntStringWithError(a, b string) (string, error) {
	aInt := new(big.Int)
	if _, ok := aInt.SetString(a, 10); !ok {
		return "", fmt.Errorf("invalid big integer string a: %q", a)
	}
	bInt := new(big.Int)
	if _, ok := bInt.SetString(b, 10); !ok {
		return "", fmt.Errorf("invalid big integer string b: %q", b)
	}
	// Ensure the result is not negative
	if aInt.Cmp(bInt) < 0 {
		return "0", nil
	}
	return new(big.Int).Sub(aInt, bInt).String(), nil
}
