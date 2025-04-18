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

package backend

import (
	ctypes "github.com/berachain/beacon-kit/consensus-types/types"
	"github.com/berachain/beacon-kit/primitives/common"
	"github.com/berachain/beacon-kit/primitives/math"
	statedb "github.com/berachain/beacon-kit/state-transition/core/state"
)

// StateFromSlotForProof returns the beacon state of the version that was used
// to calculate the parent beacon block root, which has the empty state root in
// the latest block header. Hence we do not process the next slot.
func (b *Backend) StateFromSlotForProof(slot math.Slot) (*statedb.StateDB, math.Slot, error) {
	return b.stateFromSlotRaw(slot)
}

// GetStateRoot returns the root of the state at the given slot.
func (b *Backend) StateRootAtSlot(slot math.Slot) (common.Root, error) {
	st, resolvedSlot, err := b.stateFromSlot(slot)
	if err != nil {
		return common.Root{}, err
	}

	// As calculated by the beacon chain. Ideally, this logic
	// should be abstracted by the beacon chain.
	return st.StateRootAtIndex(resolvedSlot.Unwrap() % b.cs.SlotsPerHistoricalRoot())
}

// StateAtSlot returns the beacon state at a particular slot.
func (b *Backend) StateAtSlot(slot math.Slot) (*statedb.StateDB, error) {
	st, _, err := b.stateFromSlot(slot)
	if err != nil {
		return nil, err
	}

	return st, nil
}

// GetStateFork returns the fork of the state at the given stateID.
func (b *Backend) StateForkAtSlot(slot math.Slot) (*ctypes.Fork, error) {
	var fork *ctypes.Fork
	st, _, err := b.stateFromSlot(slot)
	if err != nil {
		return fork, err
	}
	return st.GetFork()
}
