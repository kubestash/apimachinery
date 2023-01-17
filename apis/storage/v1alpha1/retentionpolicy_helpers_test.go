/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

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
			expectedDuration: Duration{Years: 7, Months: 11, Days: 27, Hours: 0, Minutes: 0},
			expectedErr:      nil,
		},
		{
			duration:         "7y27d11h",
			expectedDuration: Duration{Years: 7, Months: 0, Days: 27, Hours: 11, Minutes: 0},
			expectedErr:      nil,
		},
		{
			duration:         "7y11h15m",
			expectedDuration: Duration{Years: 7, Months: 0, Days: 0, Hours: 11, Minutes: 15},
			expectedErr:      nil,
		},
		{
			duration:         "11mo11d11h",
			expectedDuration: Duration{Years: 0, Months: 11, Days: 11, Hours: 11, Minutes: 0},
			expectedErr:      nil,
		},
		{
			duration:         "11mo11h11m",
			expectedDuration: Duration{Years: 0, Months: 11, Days: 0, Hours: 11, Minutes: 11},
			expectedErr:      nil,
		},
		{
			duration:         "5y12mo",
			expectedDuration: Duration{Years: 5, Months: 12, Days: 0, Hours: 0, Minutes: 0},
			expectedErr:      nil,
		},
		{
			duration:         "5y12d",
			expectedDuration: Duration{Years: 5, Months: 0, Days: 12, Hours: 0, Minutes: 0},
			expectedErr:      nil,
		},
		{
			duration:         "5y12h",
			expectedDuration: Duration{Years: 5, Months: 0, Days: 0, Hours: 12, Minutes: 0},
			expectedErr:      nil,
		},
		{
			duration:         "5y12m",
			expectedDuration: Duration{Years: 5, Months: 0, Days: 0, Hours: 0, Minutes: 12},
			expectedErr:      nil,
		},
		{
			duration:         "6mo15d",
			expectedDuration: Duration{Years: 0, Months: 6, Days: 15, Hours: 0, Minutes: 0},
			expectedErr:      nil,
		},
		{
			duration:         "6mo15h",
			expectedDuration: Duration{Years: 0, Months: 6, Days: 0, Hours: 15, Minutes: 0},
			expectedErr:      nil,
		},
		{
			duration:         "6mo15m",
			expectedDuration: Duration{Years: 0, Months: 6, Days: 0, Hours: 0, Minutes: 15},
			expectedErr:      nil,
		},
		{
			duration:         "6d12h",
			expectedDuration: Duration{Years: 0, Months: 0, Days: 6, Hours: 12, Minutes: 0},
			expectedErr:      nil,
		},
		{
			duration:         "6d12m",
			expectedDuration: Duration{Years: 0, Months: 0, Days: 6, Hours: 0, Minutes: 12},
			expectedErr:      nil,
		},
		{
			duration:         "12h30m",
			expectedDuration: Duration{Years: 0, Months: 0, Days: 0, Hours: 12, Minutes: 30},
			expectedErr:      nil,
		},
		{
			duration:         "10y",
			expectedDuration: Duration{Years: 10, Months: 0, Days: 0, Hours: 0, Minutes: 0},
			expectedErr:      nil,
		},
		{
			duration:         "10mo",
			expectedDuration: Duration{Years: 0, Months: 10, Days: 0, Hours: 0, Minutes: 0},
			expectedErr:      nil,
		},
		{
			duration:         "10d",
			expectedDuration: Duration{Years: 0, Months: 0, Days: 10, Hours: 0, Minutes: 0},
			expectedErr:      nil,
		},
		{
			duration:         "10h",
			expectedDuration: Duration{Years: 0, Months: 0, Days: 0, Hours: 10, Minutes: 0},
			expectedErr:      nil,
		},
		{
			duration:         "10m",
			expectedDuration: Duration{Years: 0, Months: 0, Days: 0, Hours: 0, Minutes: 10},
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
