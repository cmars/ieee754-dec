// Copyright 2014 Casey Marshall. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decimal

import (
	"testing"
)

func TestMaxCoeff(t *testing.T) {
	d := Dec32(0x6408967f)
	coeff, exp, ok := d.Decode()
	if coeff != maxCoeff {
		t.Errorf("unexpected coeff %d", coeff)
	}
	if exp != 0 {
		t.Errorf("unexpected exp %d", exp)
	}
	if !ok {
		t.Errorf("decode failed")
	}
	if !d.Valid() {
		t.Errorf("not valid")
	}
}

func TestZero(t *testing.T) {
	d := Dec32(0x0)
	if !d.Zero() {
		t.Errorf("should be zero")
	}
}

func TestEncDec(t *testing.T) {
	testCases := []struct {
		coeff int32
		exp   int8
		ref   Dec32
		ok    bool
	}{
		// Min exp
		{2, -95, Dec32(0x42100002), true},
		// Min exp - 1
		{2, -96, failDec32, false},
		// Max exp
		{2, 96, Dec32(0x22000002), true},
		// Max exp + 1
		{2, 97, failDec32, false},
		// Max coeff
		{9999999, 0, Dec32(0x6408967f), true},
		// Max coeff + 1
		{10000000, 0, failDec32, false},
		// Max coeff fitting in 23 bits
		{8388607, 0, Dec32(0x1c0fffff), true},
		// Max coeff fitting in 23 bits + 1
		{8388608, 0, Dec32(0x60000000), true},
	}
	for i, testCase := range testCases {
		d, ok := EncodeDec32(testCase.coeff, testCase.exp)
		if ok != testCase.ok {
			t.Errorf("testCase #%d: expect ok=%s, got %s", i, testCase.ok, ok)
		}
		if ok {
			if d != testCase.ref {
				t.Errorf("testCase #%d: expect dec32=%x, got %x", i, uint32(testCase.ref), uint32(d))
			}
			coeff, exp, ok := d.Decode()
			if coeff != testCase.coeff {
				t.Errorf("testCase #%d: expect coeff=%d, got %d", i, testCase.coeff, coeff)
			}
			if exp != testCase.exp {
				t.Errorf("testCase #%d: expect exp=%d, got %d", i, testCase.exp, exp)
			}
			if ok != testCase.ok {
				t.Errorf("testCase #%d: expect ok=%s, got %s", i, testCase.ok, ok)
			}
		}
	}
}
