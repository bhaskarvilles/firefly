// Copyright © 2021 Kaleido, Inc.
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

package retry

import (
	"context"
	"testing"
	"time"
)

func TestRetryEventuallyOk(t *testing.T) {
	r := Retry{
		MaximumDelay: 3 * time.Microsecond,
		InitialDelay: 1 * time.Microsecond,
	}
	r.Do(context.Background(), func(i int) (retry bool) {
		return i < 10
	})
}

func TestRetryDeadlineTimeout(t *testing.T) {
	r := Retry{
		MaximumDelay: 3 * time.Microsecond,
		InitialDelay: 0,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
	defer cancel()
	r.Do(ctx, func(i int) (retry bool) {
		return true
	})
}

func TestNegativeDelayExit(t *testing.T) {
	r := Retry{
		InitialDelay: -1,
	}
	r.Do(context.Background(), func(i int) (retry bool) {
		return true
	})
}
