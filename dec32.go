// Copyright 2014 Casey Marshall. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decimal

import "math"

// Dec32 stores a decimal32 value: a 32-bit signed decimal floating-point number
// as defined in IEEE-754-2008. Dec32 can hold a significand in the range
// of 0-9999999, multiplied by 10^exp, where -95 <= exp <= 96.
// This implementation stores the significand as a binary integer decimal.
type Dec32 uint32

const (
	// Representation:         s mmmmm xxxxxx cccccccccccccccccccc
	signMask = 0x80000000 // --+ |     |      |
	combMask = 0x7c000000 // ----+     |      |
	expMask  = 0x03f00000 // ----------+      |
	contMask = 0x000fffff // -----------------+

	coeffMaxBits = 0x00800000
	coeffMaxMask = 0x00100000

	signOffset = 31
	combOffset = 26
	expOffset  = 20

	maxCoeff = 9999999
	minExp   = -95
	maxExp   = 96
)

// Sign returns -1 if the decimal is negative, 1 if the decimal is positive.
// Zero values can be either positive or negative.
func (d Dec32) Sign() int {
	if (d & signMask) == signMask {
		return -1
	}
	return 1
}

// Zero returns whether the decimal32 represents a zero value.  Coefficient
// values greater than the maximum 9,999,999 also represent zero according to
// the IEEE-754-2008 spec.
func (d Dec32) Zero() bool {
	coeff, _, _ := d.Decode()
	return coeff == 0 || coeff > maxCoeff
}

// Valid returns whether the decimal value is well-formed according to the
// IEEE-754-2008 specification. Exponent and coefficient values beyond the
// spec limits are invalid.
func (d Dec32) Valid() bool {
	coeff, exp, ok := d.Decode()
	return ok && exp >= minExp && exp <= maxExp && coeff <= maxCoeff
}

// IsInf returns whether the decimal32 value is infinite.
func (d Dec32) IsInf() bool {
	return d.combBits() == 0x1e
}

// IsInf returns whether the decimal32 value is not-a-number (NaN).
func (d Dec32) IsNaN() bool {
	return d.combBits() == 0x1f
}

func (d Dec32) combBits() uint32 {
	return ((uint32(d) & combMask) >> combOffset)
}

func (d Dec32) expBits() uint32 {
	return ((uint32(d) & expMask) >> expOffset)
}

func (d Dec32) contBits() uint32 {
	return (uint32(d) & contMask)
}

var failDec32 = Dec32(0xffffffff)

// EncodeDec32 encodes the given coefficient and exponent into a decimal value.
func EncodeDec32(coeff int32, exp int8) (Dec32, bool) {
	var result uint32
	if coeff < 0 {
		result = result | signMask
		coeff = 0 - coeff
	}
	if coeff > maxCoeff {
		return failDec32, false
	}
	if exp < minExp || exp > maxExp {
		return failDec32, false
	}
	result |= (uint32(coeff) & contMask)
	uexp := uint8(exp)
	if (coeffMaxBits & coeff) == coeffMaxBits {
		// coefficient starts with 100
		result |= 0x60000000 | (uint32(uexp&0xc0) << 21) | ((uint32(coeff) & coeffMaxMask) << 6) | (uint32(uexp&0x3f) << 20)
	} else {
		result |= (uint32(uexp&0xc0) << 23) | ((uint32(coeff) &^ contMask) << 6) | (uint32(uexp&0x3f) << 20)
	}
	return Dec32(result), true
}

// Decode decodes a decimal32 value into its coefficient and exponent
// components, and whether the value can be decoded. Infinite, NaN and illegal
// values cannot be decoded to a coefficient and exponent.
func (d Dec32) Decode() (coeff int32, expn int8, ok bool) {
	comb, exp, cont := d.combBits(), d.expBits(), d.contBits()
	if (comb & 0x18) == 0x18 {
		coeff = int32(((comb & 0x01) << 20) | coeffMaxBits | cont)
		expn = int8(((comb << 5) & 0xc0) | exp)
		ok = true
	} else {
		coeff = int32(((comb & 0x07) << 20) | cont)
		expn = int8(((comb << 3) & 0xc0) | exp)
		ok = true
	}
	if ok {
		if (d & signMask) == signMask {
			coeff = 0 - coeff
		}
	}
	return coeff, expn, ok
}

// Float32 returns a binary floating-point approximation of the decimal value.
func (d Dec32) Float32() float32 {
	coeff, exp, ok := d.Decode()
	if !ok {
		return float32(math.NaN())
	}
	var expf float32
	if exp > 0 {
		expf = 10.0 * float32(exp)
	} else if exp < 0 {
		expf = 1.0 / (10.0 * float32(exp))
	} else {
		return float32(coeff)
	}
	return float32(coeff) * expf
}
