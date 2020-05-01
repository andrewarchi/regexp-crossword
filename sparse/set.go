// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sparse implements sparse sets.
package sparse

// A Set is a sparse set of uint32 values.
// https://research.swtch.com/sparse
type Set struct {
	dense  []uint32
	sparse []uint32
}

// NewSet returns a new Set with a given maximum size.
// The set can contain numbers in [0, max-1].
func NewSet(max uint32) *Set {
	return &Set{
		sparse: make([]uint32, max),
	}
}

// Init initializes a Set to have a given maximum size.
// The set can contain numbers in [0, max-1].
func (s *Set) Init(max uint32) {
	s.sparse = make([]uint32, max)
}

// Reset clears (empties) the set.
func (s *Set) Reset() {
	s.dense = s.dense[:0]
}

// Add adds x to the set if it is not already there.
func (s *Set) Add(x uint32) {
	v := s.sparse[x]
	if v < uint32(len(s.dense)) && s.dense[v] == x {
		return
	}
	n := len(s.dense)
	s.sparse[x] = uint32(n)
	s.dense = append(s.dense, x)
}

// Has reports whether x is in the set.
func (s *Set) Has(x uint32) bool {
	v := s.sparse[x]
	return v < uint32(len(s.dense)) && s.dense[v] == x
}

// Values returns the values in the set.
// The values are listed in the order in which they
// were inserted.
func (s *Set) Values() []uint32 {
	return s.dense
}

// Len returns the number of values in the set.
func (s *Set) Len() int {
	return len(s.dense)
}
