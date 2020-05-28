package utils

import (
	"testing"
)

func TestUnitConversion(t *testing.T) {
	input := "123460"
	decimal := 12
	retainNum := 6
	r := UnitConversion(input, decimal, retainNum)
	t.Log(r)
}

func TestFormatTokenAmount(t *testing.T) {
	input := "1.0010000001"
	decimal := 6
	res := FormatTokenAmount(input, decimal)
	t.Log(res)

	// f, err := strconv.ParseFloat(input, 64)
	// t.Log(err)
	// t.Log(f)
	// price, err := decimal.NewFromString("11.999222")
	// t.Log(err)
	//
	// diff := decimal.NewFromFloat(1.01)
	// res := price.Mul(diff)
	// t.Log(res.String())

}
