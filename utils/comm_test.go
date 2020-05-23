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
	input := "1.001"
	decimal := 8
	res := FormatTokenAmount(input, decimal)
	t.Log(res)
}
