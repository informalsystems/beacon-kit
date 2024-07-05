// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2024, Berachain Foundation. All rights reserved.
// Use of this software is governed by the Business Source License included
// in the LICENSE file of this repository and at www.mariadb.com/bsl11.
//
// ANY USE OF THE LICENSED WORK IN VIOLATION OF THIS LICENSE WILL AUTOMATICALLY
// TERMINATE YOUR RIGHTS UNDER THIS LICENSE FOR THE CURRENT AND ALL OTHER
// VERSIONS OF THE LICENSED WORK.
//
// THIS LICENSE DOES NOT GRANT YOU ANY RIGHT IN ANY TRADEMARK OR LOGO OF
// LICENSOR OR ITS AFFILIATES (PROVIDED THAT YOU MAY USE A TRADEMARK OR LOGO OF
// LICENSOR AS EXPRESSLY REQUIRED BY THIS LICENSE).
//
// TO THE EXTENT PERMITTED BY APPLICABLE LAW, THE LICENSED WORK IS PROVIDED ON
// AN “AS IS” BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
// EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
// TITLE.

package merkle

import (
	"unsafe"

	"github.com/berachain/beacon-kit/mod/errors"
	"github.com/berachain/beacon-kit/mod/primitives/pkg/crypto"
	"github.com/berachain/beacon-kit/mod/primitives/pkg/math"
	"github.com/berachain/beacon-kit/mod/primitives/pkg/merkle/zero"
	"github.com/prysmaticlabs/gohashtree"
)

const (
	// MinParallelizationSize is the minimum size of the input list that
	// should be hashed using the default method. If the input list is smaller
	// than this size, the overhead of parallelizing the hashing process is.
	//
	// TODO: This value is arbitrary and should be benchmarked to find the
	// optimal value.
	MinParallelizationSize = 5000
	// two is a constant to make the linter happy.
	two = 2
)

// RootHashFn is a function that hashes the input leaves into the output.
type RootHashFn[RootT ~[32]byte] func(output, input []RootT) error

// RootHasher is a struct that hashes the input leaves into the output.
type RootHasher[RootT ~[32]byte] struct {
	// Hasher is the underlying hasher for combi and mixins.
	crypto.Hasher[RootT]
	// rootHashFn is the underlying root hasher for the tree.
	rootHashFn RootHashFn[RootT]
}

// NewRootHasher constructs a new RootHasher.
func NewRootHasher[RootT ~[32]byte](
	hasher crypto.Hasher[RootT],
	rootHashFn RootHashFn[RootT],
) *RootHasher[RootT] {
	return &RootHasher[RootT]{
		Hasher:     hasher,
		rootHashFn: rootHashFn,
	}
}

// NewRootWithMaxLeaves constructs a Merkle tree root from a set of.
func (rh *RootHasher[RootT]) NewRootWithMaxLeaves(
	leaves []RootT,
	limit math.U64,
) (RootT, error) {
	count := math.U64(len(leaves))
	if count > limit {
		return zero.Hashes[0], errors.New("number of leaves exceeds limit")
	}
	if limit == 0 {
		return zero.Hashes[0], nil
	}
	if limit == 1 && count == 1 {
		return leaves[0], nil
	}

	return rh.NewRootWithDepth(
		leaves,
		count.NextPowerOfTwo().ILog2Ceil(),
		limit.NextPowerOfTwo().ILog2Ceil(),
	)
}

// NewRootWithDepth constructs a Merkle tree root from a set of leaves.
func (rh *RootHasher[RootT]) NewRootWithDepth(
	leaves []RootT,
	depth uint8,
	limitDepth uint8,
) (RootT, error) {
	// Short-circuit to getting memory from the buffer.
	if len(leaves) == 0 {
		return zero.Hashes[limitDepth], nil
	}

	var err error
	for i := range depth {
		layerLen := len(leaves)
		if layerLen%two == 1 {
			leaves = append(leaves, zero.Hashes[i])
		}

		if err = rh.rootHashFn(leaves, leaves); err != nil {
			return zero.Hashes[limitDepth], err
		}
		leaves = leaves[:(layerLen+1)/two]
	}

	// If something went wrong, return the zero hash of limitDepth.
	if len(leaves) != 1 {
		return zero.Hashes[limitDepth], nil
	}

	// Handle the case where the tree is not full.
	h := leaves[0]
	for j := depth; j < limitDepth; j++ {
		h = rh.Combi(h, zero.Hashes[j])
	}

	return h, nil
}

// BuildParentTreeRoots calls BuildParentTreeRootsWithNRoutines with the
// number of routines set to runtime.GOMAXPROCS(0)-1.
//
// TODO: enable parallelization.
func BuildParentTreeRoots[RootT ~[32]byte](
	outputList, inputList []RootT,
) error {
	return gohashtree.Hash(
		//#nosec:G103 // on purpose.
		*(*[][32]byte)(unsafe.Pointer(&outputList)),
		//#nosec:G103 // on purpose.
		*(*[][32]byte)(unsafe.Pointer(&inputList)),
	)
}