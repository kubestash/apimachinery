/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		duration         string
		expectedDuration Duration
		expectedErr      error
	}{
		{
			duration:         "1y2mo3d4h5m",
			expectedDuration: Duration{Years: 1, Months: 2, Days: 3, Hours: 4, Minutes: 5},
			expectedErr:      nil,
		},
		{
			duration:         "7y11mo27d",
			expectedDuration: Duration{Years: 7, Months: 11, Days: 27},
			expectedErr:      nil,
		},
		{
			duration:         "7y27d11h",
			expectedDuration: Duration{Years: 7, Days: 27, Hours: 11},
			expectedErr:      nil,
		},
		{
			duration:         "7y11h15m",
			expectedDuration: Duration{Years: 7, Hours: 11, Minutes: 15},
			expectedErr:      nil,
		},
		{
			duration:         "11mo11d11h",
			expectedDuration: Duration{Months: 11, Days: 11, Hours: 11},
			expectedErr:      nil,
		},
		{
			duration:         "11mo11h11m",
			expectedDuration: Duration{Months: 11, Hours: 11, Minutes: 11},
			expectedErr:      nil,
		},
		{
			duration:         "5y12mo",
			expectedDuration: Duration{Years: 5, Months: 12},
			expectedErr:      nil,
		},
		{
			duration:         "5y12d",
			expectedDuration: Duration{Years: 5, Days: 12},
			expectedErr:      nil,
		},
		{
			duration:         "5y12h",
			expectedDuration: Duration{Years: 5, Hours: 12},
			expectedErr:      nil,
		},
		{
			duration:         "5y12m",
			expectedDuration: Duration{Years: 5, Minutes: 12},
			expectedErr:      nil,
		},
		{
			duration:         "6mo15d",
			expectedDuration: Duration{Months: 6, Days: 15},
			expectedErr:      nil,
		},
		{
			duration:         "6mo15h",
			expectedDuration: Duration{Months: 6, Hours: 15},
			expectedErr:      nil,
		},
		{
			duration:         "6mo15m",
			expectedDuration: Duration{Months: 6, Minutes: 15},
			expectedErr:      nil,
		},
		{
			duration:         "6d12h",
			expectedDuration: Duration{Days: 6, Hours: 12},
			expectedErr:      nil,
		},
		{
			duration:         "6d12m",
			expectedDuration: Duration{Days: 6, Minutes: 12},
			expectedErr:      nil,
		},
		{
			duration:         "12h30m",
			expectedDuration: Duration{Hours: 12, Minutes: 30},
			expectedErr:      nil,
		},
		{
			duration:         "10y",
			expectedDuration: Duration{Years: 10},
			expectedErr:      nil,
		},
		{
			duration:         "10mo",
			expectedDuration: Duration{Months: 10},
			expectedErr:      nil,
		},
		{
			duration:         "10d",
			expectedDuration: Duration{Days: 10},
			expectedErr:      nil,
		},
		{
			duration:         "10h",
			expectedDuration: Duration{Hours: 10},
			expectedErr:      nil,
		},
		{
			duration:         "10m",
			expectedDuration: Duration{Minutes: 10},
			expectedErr:      nil,
		},
		{
			duration:         "2w",
			expectedDuration: Duration{Weeks: 2},
			expectedErr:      nil,
		},
		{
			duration:         "1y2w",
			expectedDuration: Duration{Years: 1, Weeks: 2},
			expectedErr:      nil,
		},
		{
			duration:         "3w3mo",
			expectedDuration: Duration{Weeks: 3, Months: 3},
			expectedErr:      nil,
		},
		{
			duration:         "2w5d",
			expectedDuration: Duration{Weeks: 2, Days: 5},
			expectedErr:      nil,
		},
		{
			duration:         "3w11h",
			expectedDuration: Duration{Weeks: 3, Hours: 11},
			expectedErr:      nil,
		},
		{
			duration:         "4w30m",
			expectedDuration: Duration{Weeks: 4, Minutes: 30},
			expectedErr:      nil,
		},
		{
			duration:         "1y2mo3w4d5h6m",
			expectedDuration: Duration{Years: 1, Months: 2, Weeks: 3, Days: 4, Hours: 5, Minutes: 6},
			expectedErr:      nil,
		},
		{
			duration:         "10",
			expectedDuration: Duration{},
			expectedErr:      errInvalidDuration,
		},
		{
			duration:         "1mm",
			expectedDuration: Duration{},
			expectedErr:      errInvalidDuration,
		},
		{
			duration:         "mo",
			expectedDuration: Duration{},
			expectedErr:      errInvalidDuration,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			d, err := ParseDuration(test.duration)
			assert.Equal(t, test.expectedDuration, d)
			assert.True(t, errors.Is(err, test.expectedErr))
		})
	}
}
