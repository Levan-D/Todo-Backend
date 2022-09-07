package utils

import (
	"github.com/shopspring/decimal"
	"math"
	"math/rand"
)

func GenerateRandomFloat(min float64, max float64) float64 {
	gen := min + rand.Float64()*(max-min)
	return math.Round(gen*100) / 100
}

func GenerateRandomDecimal(min float64, max float64) decimal.Decimal {
	gen := min + rand.Float64()*(max-min)
	return decimal.NewFromFloat(math.Round(gen*100) / 100)
}
