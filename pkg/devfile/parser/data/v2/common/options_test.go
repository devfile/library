package common

import (
	"testing"

	"github.com/devfile/api/pkg/attributes"
)

func TestFilterDevfileObject(t *testing.T) {

	tests := []struct {
		name       string
		attributes attributes.Attributes
		options    DevfileOptions
		wantFilter bool
		wantErr    bool
	}{
		{
			name: "Case 1: Filter with one key",
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
			wantErr:    false,
		},
		{
			name: "Case 2: Filter with two keys",
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
			wantErr:    false,
		},
		{
			name: "Case 3: Filter with missing key",
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
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterIn, err := FilterDevfileObject(tt.attributes, tt.options)
			if !tt.wantErr && err != nil {
				t.Errorf("TestFilterDevfileObject unexpected error - %v", err)
			} else if tt.wantErr && err == nil {
				t.Errorf("TestFilterDevfileObject wanted error got nil")
			} else if filterIn != tt.wantFilter {
				t.Errorf("TestFilterDevfileObject error - expected %v got %v", tt.wantFilter, filterIn)
			}
		})
	}
}
