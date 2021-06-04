// Copyright © 2021 Kaleido, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package broadcast

import (
	"context"

	"github.com/kaleido-io/firefly/internal/log"
	"github.com/kaleido-io/firefly/pkg/fftypes"
)

func (bm *broadcastManager) handleOrganizationBroadcast(ctx context.Context, msg *fftypes.Message, data []*fftypes.Data) (valid bool, err error) {
	l := log.L(ctx)

	var org fftypes.Organization
	valid = bm.getSystemBroadcastPayload(ctx, msg, data, &org)
	if !valid {
		return false, nil
	}

	if err = org.Validate(ctx, true); err != nil {
		l.Warnf("Unable to process organization broadcast %s - validate failed: %s", msg.Header.ID, err)
		return false, nil
	}

	signingIdentity := org.Identity
	if org.Parent != "" {
		signingIdentity = org.Parent
		parent, err := bm.database.GetOrganization(ctx, org.Parent)
		if err != nil {
			return false, err // We only return database errors
		}
		if parent == nil {
			l.Warnf("Unable to process organization broadcast %s - parent identity not found: %s", msg.Header.ID, org.Parent)
			return false, nil
		}
	}

	id, err := bm.identity.Resolve(ctx, signingIdentity)
	if err != nil {
		l.Warnf("Unable to process organization broadcast %s - resolve identity failed: %s", msg.Header.ID, err)
		return false, nil
	}

	if msg.Header.Author != id.OnChain {
		l.Warnf("Unable to process organization broadcast %s - incorrect signature. Expected=%s Received=%s", msg.Header.ID, id.OnChain, msg.Header.Author)
		return false, nil
	}

	existing, err := bm.database.GetOrganization(ctx, org.Identity)
	if err != nil {
		return false, err // We only return database errors
	}
	if existing != nil {
		if existing.Parent != org.Parent {
			l.Warnf("Unable to process organization broadcast %s - mismatch with existing %v", msg.Header.ID, existing.ID)
			return false, nil
		}
		org.ID = nil // we keep the existing ID
	}

	if err = bm.database.UpsertOrganization(ctx, &org, true); err != nil {
		return false, err
	}

	return true, nil
}