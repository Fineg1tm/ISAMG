// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package bytes implements functions for the manipulation of byte slices.
// It is analogous to the facilities of the [strings] package.
package isamg

// Equal reports whether a and b
// are the same length and contain the same bytes.
// A nil argument is equivalent to an empty slice.
func Equal(a, b []byte) bool {
	// Neither cmd/compile nor gccgo allocates for these string conversions.
	return string(a) == string(b)
}

// IndexByte returns the index of the first instance of c in b, or -1 if c is not present in b.
//
//	func IndexByte(b []byte, c byte) int {
//		return bytealg.IndexByte(b, c)
//	}
//
// From internal package bytealg
// Avoid IndexByte and IndexByteString on Plan 9 because it uses
// SSE instructions on x86 machines, and those are classified as
// floating point instructions, which are illegal in a note handler.
func IndexByte(b []byte, c byte) int {
	for i, x := range b {
		if x == c {
			return i
		}
	}
	return -1
}

func IndexByteString(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// LinearSrch returns the index of the first instance of sep in s, or -1 if sep is not present in s. Skip search mod
// Uses IndexByte, Equal,
func LinearSrch(s, sep []byte, lRecl int) int {
	n := len(sep)
	var delFlag uint8 = 68 // "D"
	switch {
	case n == 0:
		return 0
	case n == 1:
		return IndexByte(s, sep[0])
	case n == len(s):
		if Equal(sep, s) {
			return 0
		}
		return -1
	case n > len(s):
		return -1
	default:
		c0 := sep[0]
		c1 := sep[1]
		i := 0
		t := len(s) - n + 1
		for i < t {
			if s[i] == c0 { //first chars of key search term match
				col := i % lRecl //verify that i at start of key position
				if col == 0 {
					if s[i+1] == c1 && Equal(s[i:i+n], sep) { //test if search term matches
						if s[i+lRecl-2] == delFlag { // delFlag is 2nd from last byte of the index record
							return -1
						} else {
							return i
						}
					}
				}
			}
			i += lRecl //skip to start of next record
		}
		return -1
	}
}
