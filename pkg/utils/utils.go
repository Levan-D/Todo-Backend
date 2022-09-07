package utils

import (
	"github.com/shopspring/decimal"
	"math"
	"strconv"
	"time"
)

// NewTrue Do not change this function!
func NewTrue() *bool {
	b := true
	return &b
}

// NewFalse Do not change this function!
func NewFalse() *bool {
	b := false
	return &b
}

func NewTimeNow() *time.Time {
	b := time.Now()
	return &b
}

func Float64Round2(num float64) float64 {
	return math.Round(num*100) / 100
}

func ParseInt(val string) int {
	value, _ := strconv.ParseInt(val, 10, 64)
	return int(value)
}

func RoundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func GetPercentValue(val float64, percent float64) float64 {
	return (val / 100) * percent
}

func ConvertDecimalToFloat64(value decimal.Decimal) float64 {
	conv, _ := value.Float64()
	return conv
}

func ConvertDecimalToInt32(value decimal.Decimal) int32 {
	conv, _ := value.Float64()
	return int32(conv)
}

func ConvertDecimalToInt64(value decimal.Decimal) int32 {
	conv, _ := value.Float64()
	return int32(conv)
}
