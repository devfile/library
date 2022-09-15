//
// Copyright 2022 Red Hat, Inc.
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

package common

import (
	"testing"

	"github.com/devfile/api/v2/pkg/attributes"
)

func TestFilterDevfileObject(t *testing.T) {

	tests := []struct {
		name       string
		attributes attributes.Attributes
		options    DevfileOptions
		wantFilter bool
	}{
		{
			name: "Filter with one key",
			attributes: attributes.Attributes{}.FromStringMap(map[string]string{
				"firstString":  "firstStringValue",
				"secondString": "secondStringValue",
			}),
			options: DevfileOptions{
				Filter: map[string]interface{}{
					"firstString": "firstStringValue",
				},
			},
			wantFilter: true,
		},
		{
			name: "Filter with two keys",
			attributes: attributes.Attributes{}.FromStringMap(map[string]string{
				"firstString":  "firstStringValue",
				"secondString": "secondStringValue",
			}),
			options: DevfileOptions{
				Filter: map[string]interface{}{
					"firstString":  "firstStringValue",
					"secondString": "secondStringValue",
				},
			},
			wantFilter: true,
		},
		{
			name: "Filter with missing key",
			attributes: attributes.Attributes{}.FromStringMap(map[string]string{
				"firstString":  "firstStringValue",
				"secondString": "secondStringValue",
			}),
			options: DevfileOptions{
				Filter: map[string]interface{}{
					"missingkey": "firstStringValue",
				},
			},
			wantFilter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterIn, err := FilterDevfileObject(tt.attributes, tt.options)
			// Unexpected error
			if err != nil {
				t.Errorf("TestFilterDevfileObject() unexpected error: %v", err)
			} else if filterIn != tt.wantFilter {
				t.Errorf("TestFilterDevfileObject() error: expected %v got %v", tt.wantFilter, filterIn)
			}
		})
	}
}
