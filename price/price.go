package price

import (
	"errors"
	"math/big"

	"github.com/ALTree/bigfloat"
)

var ErrNotExact error = errors.New("conversion from float to int is not exact")

// Receives a float and the number of decimals and returns an int with the decimals added
func AddDecimals(num *big.Float, dec *big.Int) *big.Int {
	numCopy := new(big.Float).Copy(num)
	numCopy.Mul(numCopy, bigfloat.Pow(big.NewFloat(10), new(big.Float).SetInt(dec)))
	result, _ := numCopy.Int(nil)
	return result
}

// Receives an int and the number of decimals and returns a float without the decimals
func RemoveDecimals(num *big.Int, dec *big.Int) *big.Float {
	result := new(big.Float).SetInt(num)
	result.Quo(result, bigfloat.Pow(big.NewFloat(10), new(big.Float).SetInt(dec)))
	return result
}

// Receives an int and multiplies it by the percentage of mult
func MultiplyPercent(num *big.Int, mult float64) *big.Int {
	floatNum := new(big.Float).SetInt(num)
	multFloat := new(big.Float).Quo(big.NewFloat(mult), big.NewFloat(100))
	result, _ := new(big.Float).Mul(floatNum, multFloat).Int(nil)
	return result
}
