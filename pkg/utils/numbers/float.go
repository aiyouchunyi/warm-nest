package numbers

import (
	"github.com/shopspring/decimal"
)

func FloorAmountFloat(f float64, precision int) string {
	return Floor(decimal.NewFromFloat(f).String(), precision)
}
