// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2025, Berachain Foundation. All rights reserved.
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

package math

import (
	"encoding/binary"
	"math/big"
	"strconv"

	"github.com/berachain/beacon-kit/errors"
	"github.com/berachain/beacon-kit/primitives/common"
	"github.com/berachain/beacon-kit/primitives/encoding/hex"
	"github.com/berachain/beacon-kit/primitives/math/log"
	"github.com/berachain/beacon-kit/primitives/math/pow"
	"github.com/ethereum/go-ethereum/params"
)

type (
	// U64 represents a 64-bit unsigned integer that is both SSZ and JSON
	// marshallable. We marshal U64 as hex strings in JSON in order to keep the
	// execution client apis happy, and we marshal U64 as little-endian in SSZ
	// to be
	// compatible with the spec.
	U64 uint64

	// Gwei is a denomination of 1e9 Wei represented as a U64.
	Gwei = U64

	// Slot as per the Ethereum 2.0 Specification:
	// https://github.com/ethereum/consensus-specs/blob/dev/specs/phase0/beacon-chain.md#custom-types
	Slot = U64

	// CommitteeIndex as per the Ethereum 2.0 Specification:
	// https://github.com/ethereum/consensus-specs/blob/dev/specs/phase0/beacon-chain.md#custom-types
	CommitteeIndex = U64

	// ValidatorIndex as per the Ethereum 2.0  Specification:
	// https://github.com/ethereum/consensus-specs/blob/dev/specs/phase0/beacon-chain.md#custom-types
	ValidatorIndex = U64

	// Epoch as per the Ethereum 2.0 Specification:
	// https://github.com/ethereum/consensus-specs/blob/dev/specs/phase0/beacon-chain.md#custom-types
	Epoch = U64
)

// -------------------------- JSONMarshallable -------------------------

// MarshalText implements encoding.TextMarshaler.
func (u U64) MarshalText() ([]byte, error) {
	return hex.MarshalText(u.Unwrap())
}

// UnmarshalJSON implements json.Unmarshaler.
func (u *U64) UnmarshalJSON(input []byte) error {
	strippedInput, err := hex.ValidateQuotedString(input)
	if err != nil {
		return err
	}
	return u.UnmarshalText(strippedInput)
}

func U64FromString(id string) (U64, error) {
	u64, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, err
	}

	return U64(u64), nil
}

// ---------------------------------- Hex ----------------------------------

// UnmarshalText implements encoding.TextUnmarshaler.
func (u *U64) UnmarshalText(input []byte) error {
	dec, err := hex.UnmarshalUint64Text(input)
	if err != nil {
		return err
	}
	*u = U64(dec)
	return nil
}

// String returns the string representation of the U64.
func (u U64) Base10() string {
	return strconv.FormatUint(uint64(u), 10)
}

// ----------------------- U64 Mathematical Methods -----------------------

// Unwrap returns a copy of the underlying uint64 value of U64.
func (u U64) Unwrap() uint64 {
	return uint64(u)
}

// UnwrapPtr returns a pointer to the underlying uint64 value of U64.
func (u U64) UnwrapPtr() *uint64 {
	return (*uint64)(&u)
}

// Get the power of 2 for given input, or the closest higher power of 2 if the
// input is not a power of 2. Commonly used for "how many nodes do I need for a
// bottom tree layer fitting x elements?"
// Example: 0->1, 1->1, 2->2, 3->4, 4->4, 5->8, 6->8, 7->8, 8->8, 9->16.
//
// https://github.com/ethereum/consensus-specs/blob/dev/ssz/merkle-proofs.md#helper-functions
func (u U64) NextPowerOfTwo() U64 {
	return pow.NextPowerOfTwo(u)
}

// Get the power of 2 for given input, or the closest lower power of 2 if the
// input is not a power of 2. The zero case is a placeholder and not used for
// math with generalized indices. Commonly used for "what power of two makes up
// the root bit of the generalized index?"
// Example: 0->1, 1->1, 2->2, 3->2, 4->4, 5->4, 6->4, 7->4, 8->8, 9->8.
//
// https://github.com/ethereum/consensus-specs/blob/dev/ssz/merkle-proofs.md#helper-functions
func (u U64) PrevPowerOfTwo() U64 {
	return pow.PrevPowerOfTwo(u)
}

// ILog2Ceil returns the ceiling of the base 2 logarithm of the U64.
func (u U64) ILog2Ceil() uint8 {
	return log.ILog2Ceil(u)
}

// ILog2Floor returns the floor of the base 2 logarithm of the U64.
func (u U64) ILog2Floor() uint8 {
	return log.ILog2Floor(u)
}

// -------------------------------- SSZ ---------------------------------

// HashTreeRoot returns the hash tree root of the U64.
func (u U64) HashTreeRoot() common.Root {
	var root common.Root
	binary.LittleEndian.PutUint64(root[:], u.Unwrap())
	return root
}

// ---------------------------- Gwei Methods ----------------------------

var ErrGweiOverflow = errors.New("gwei from big.Int overflows")

// GweiFromWei returns the value of Wei in Gwei.
func GweiFromWei(i *big.Int) (Gwei, error) {
	intToGwei := big.NewInt(0).SetUint64(params.GWei)
	i.Div(i, intToGwei)
	if !i.IsUint64() {
		// a Gwei amount >= (2**64) * (10**9) or negative would not
		// be representable as uint64. This should not happen but
		// we still guard against a serialization bug or other mishap.
		return 0, ErrGweiOverflow
	}
	return Gwei(i.Uint64()), nil
}

// ToWei converts a value from Gwei to Wei.
func (u Gwei) ToWei() *U256 {
	return (&U256{}).Mul(NewU256(uint64(u)), NewU256(params.GWei))
}
