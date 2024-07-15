// Copyright (c) 2013, 2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package util_test

import (
	"math"
	"testing"

	"github.com/harbi-network/harbid/domain/consensus/utils/constants"
)

func TestAmountCreation(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		valid    bool
		expected Amount
	}{
		// Positive tests.
		{
			name:     "zero",
			amount:   0,
			valid:    true,
			expected: 0,
		},
		{
			name:     "max producible",
			amount:   29e9,
			valid:    true,
			expected: Amount(constants.MaxLeor),
		},
		{
			name:     "one hundred",
			amount:   100,
			valid:    true,
			expected: 100 * constants.LeorPerHarb,
		},
		{
			name:     "fraction",
			amount:   0.01234567,
			valid:    true,
			expected: 1234567,
		},
		{
			name:     "rounding up",
			amount:   54.999999999999943157,
			valid:    true,
			expected: 55 * constants.LeorPerHarb,
		},
		{
			name:     "rounding down",
			amount:   55.000000000000056843,
			valid:    true,
			expected: 55 * constants.LeorPerHarb,
		},

		// Negative tests.
		{
			name:   "not-a-number",
			amount: math.NaN(),
			valid:  false,
		},
		{
			name:   "-infinity",
			amount: math.Inf(-1),
			valid:  false,
		},
		{
			name:   "+infinity",
			amount: math.Inf(1),
			valid:  false,
		},
	}

	for _, test := range tests {
		a, err := NewAmount(test.amount)
		switch {
		case test.valid && err != nil:
			t.Errorf("%v: Positive test Amount creation failed with: %v", test.name, err)
			continue
		case !test.valid && err == nil:
			t.Errorf("%v: Negative test Amount creation succeeded (value %v) when should fail", test.name, a)
			continue
		}

		if a != test.expected {
			t.Errorf("%v: Created amount %v does not match expected %v", test.name, a, test.expected)
			continue
		}
	}
}

func TestAmountUnitConversions(t *testing.T) {
	tests := []struct {
		name      string
		amount    Amount
		unit      AmountUnit
		converted float64
		s         string
	}{
		{
			name:      "MHAR",
			amount:    Amount(constants.MaxLeor),
			unit:      AmountMegaHAR,
			converted: 29000,
			s:         "29000 MHAR",
		},
		{
			name:      "kHAR",
			amount:    44433322211100,
			unit:      AmountKiloHAR,
			converted: 444.33322211100,
			s:         "444.333222111 kHAR",
		},
		{
			name:      "HAR",
			amount:    44433322211100,
			unit:      AmountHAR,
			converted: 444333.22211100,
			s:         "444333.222111 HAR",
		},
		{
			name:      "mHAR",
			amount:    44433322211100,
			unit:      AmountMilliHAR,
			converted: 444333222.11100,
			s:         "444333222.111 mHAR",
		},
		{

			name:      "μHAR",
			amount:    44433322211100,
			unit:      AmountMicroHAR,
			converted: 444333222111.00,
			s:         "444333222111 μHAR",
		},
		{

			name:      "leor",
			amount:    44433322211100,
			unit:      AmountLeor,
			converted: 44433322211100,
			s:         "44433322211100 Leor",
		},
		{

			name:      "non-standard unit",
			amount:    44433322211100,
			unit:      AmountUnit(-1),
			converted: 4443332.2211100,
			s:         "4443332.22111 1e-1 HAR",
		},
	}

	for _, test := range tests {
		f := test.amount.ToUnit(test.unit)
		if f != test.converted {
			t.Errorf("%v: converted value %v does not match expected %v", test.name, f, test.converted)
			continue
		}

		s := test.amount.Format(test.unit)
		if s != test.s {
			t.Errorf("%v: format '%v' does not match expected '%v'", test.name, s, test.s)
			continue
		}

		// Verify that Amount.ToHAR works as advertised.
		f1 := test.amount.ToUnit(AmountHAR)
		f2 := test.amount.ToHAR()
		if f1 != f2 {
			t.Errorf("%v: ToHAR does not match ToUnit(AmountHAR): %v != %v", test.name, f1, f2)
		}

		// Verify that Amount.String works as advertised.
		s1 := test.amount.Format(AmountHAR)
		s2 := test.amount.String()
		if s1 != s2 {
			t.Errorf("%v: String does not match Format(AmountHAR): %v != %v", test.name, s1, s2)
		}
	}
}

func TestAmountMulF64(t *testing.T) {
	tests := []struct {
		name string
		amt  Amount
		mul  float64
		res  Amount
	}{
		{
			name: "Multiply 0.1 HAR by 2",
			amt:  100e5, // 0.1 HAR
			mul:  2,
			res:  200e5, // 0.2 HAR
		},
		{
			name: "Multiply 0.2 HAR by 0.02",
			amt:  200e5, // 0.2 HAR
			mul:  1.02,
			res:  204e5, // 0.204 HAR
		},
		{
			name: "Round down",
			amt:  49, // 49 Leor
			mul:  0.01,
			res:  0,
		},
		{
			name: "Round up",
			amt:  50, // 50 Leor
			mul:  0.01,
			res:  1, // 1 Leor
		},
		{
			name: "Multiply by 0.",
			amt:  1e8, // 1 HAR
			mul:  0,
			res:  0, // 0 HAR
		},
		{
			name: "Multiply 1 by 0.5.",
			amt:  1, // 1 Leor
			mul:  0.5,
			res:  1, // 1 Leor
		},
		{
			name: "Multiply 100 by 66%.",
			amt:  100, // 100 Leor
			mul:  0.66,
			res:  66, // 66 Leor
		},
		{
			name: "Multiply 100 by 66.6%.",
			amt:  100, // 100 Leor
			mul:  0.666,
			res:  67, // 67 Leor
		},
		{
			name: "Multiply 100 by 2/3.",
			amt:  100, // 100 Leor
			mul:  2.0 / 3,
			res:  67, // 67 Leor
		},
	}

	for _, test := range tests {
		a := test.amt.MulF64(test.mul)
		if a != test.res {
			t.Errorf("%v: expected %v got %v", test.name, test.res, a)
		}
	}
}
