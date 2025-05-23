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

package common_test

import (
	"strings"
	"testing"

	"github.com/berachain/beacon-kit/primitives/bytes"
	"github.com/berachain/beacon-kit/primitives/common"
	"github.com/berachain/beacon-kit/primitives/encoding/hex"
	"github.com/stretchr/testify/require"
)

func TestNewRootFromHex(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		input       func() string
		expectedErr error
	}{
		{
			name: "EmptyString",
			input: func() string {
				return ""
			},
			expectedErr: hex.ErrEmptyString,
		},
		{
			name: "ShortSize",
			input: func() string {
				return hex.Prefix + strings.Repeat("f", 2*common.RootSize-2)
			},
			expectedErr: bytes.ErrIncorrectLength,
		},
		{
			name: "RightSize",
			input: func() string {
				return hex.Prefix + strings.Repeat("f", 2*common.RootSize)
			},
			expectedErr: nil,
		},
		{
			name: "LongSize",
			input: func() string {
				return hex.Prefix + strings.Repeat("f", 2*common.RootSize+2)
			},
			expectedErr: bytes.ErrIncorrectLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var err error
			f := func() {
				input := tt.input()
				_, err = common.NewRootFromHex(input)
			}
			require.NotPanics(t, f)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRoot_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		input       []byte
		expectedErr error
	}{
		{
			name:        "nil input",
			input:       nil,
			expectedErr: bytes.ErrIncorrectLength,
		},
		{
			name:        "empty input",
			input:       []byte(``),
			expectedErr: bytes.ErrIncorrectLength,
		},
		{
			name:        "short input of 1 byte",
			input:       []byte{0x01},
			expectedErr: bytes.ErrIncorrectLength,
		},
		{
			name:        "short input of just quotes",
			input:       []byte(`""`),
			expectedErr: hex.ErrEmptyString,
		},
		{
			name:        "valid input",
			input:       []byte(`"0x6969696969696969696969696969696969696969696969696969696969696969"`),
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var (
				r     common.Root
				err   error
				input = tt.input
			)

			f := func() {
				err = r.UnmarshalJSON(input)
			}
			require.NotPanics(t, f)

			if tt.expectedErr != nil {
				require.ErrorContains(t, err, tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
