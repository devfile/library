package v2

import (
	"reflect"
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	"github.com/kylelemons/godebug/pretty"
)

func TestGetAttributes(t *testing.T) {

	tests := []struct {
		name           string
		devfilev2      *DevfileV2
		wantAttributes attributes.Attributes
		wantErr        bool
	}{
		{
			name: "Schema 2.0.0 does not have attributes",
			devfilev2: &DevfileV2{
				v1alpha2.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{
						SchemaVersion: "2.0.0",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Schema 2.1.0 has attributes",
			devfilev2: &DevfileV2{
				v1alpha2.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{
						SchemaVersion: "2.1.0",
					},
					DevWorkspaceTemplateSpec: v1alpha2.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
							Attributes: attributes.Attributes{}.PutString("key1", "value1").PutString("key2", "value2"),
						},
					},
				},
			},
			wantAttributes: attributes.Attributes{}.PutString("key1", "value1").PutString("key2", "value2"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attributes, err := tt.devfilev2.GetAttributes()
			if tt.wantErr == (err == nil) {
				t.Errorf("TestGetAttributes error - %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				if !reflect.DeepEqual(attributes, tt.wantAttributes) {
					t.Errorf("TestGetAttributes error - actual does not equal expected, difference at %+v", pretty.Compare(attributes, tt.wantAttributes))
				}
			}
		})
	}
}

func TestUpdateAttributes(t *testing.T) {

	nestedValue := map[string]interface{}{
		"key1.1": map[string]interface{}{
			"key1.1.1": "value1.1.1",
		},
	}

	tests := []struct {
		name           string
		devfilev2      *DevfileV2
		key            string
		value          interface{}
		wantAttributes attributes.Attributes
		wantErr        bool
	}{
		{
			name: "Schema 2.0.0 does not have attributes",
			devfilev2: &DevfileV2{
				v1alpha2.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{
						SchemaVersion: "2.0.0",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Schema 2.1.0 has the top-level key attribute",
			devfilev2: &DevfileV2{
				v1alpha2.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{
						SchemaVersion: "2.1.0",
					},
					DevWorkspaceTemplateSpec: v1alpha2.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
							Attributes: attributes.Attributes{}.PutString("key1", "value1").PutString("key2", "value2"),
						},
					},
				},
			},
			key:            "key1",
			value:          nestedValue,
			wantAttributes: attributes.Attributes{}.Put("key1", nestedValue, nil).PutString("key2", "value2"),
		},
		{
			name: "Schema 2.1.0 does not have the top-level key attribute",
			devfilev2: &DevfileV2{
				v1alpha2.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{
						SchemaVersion: "2.1.0",
					},
					DevWorkspaceTemplateSpec: v1alpha2.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
							Attributes: attributes.Attributes{}.PutString("key1", "value1").PutString("key2", "value2"),
						},
					},
				},
			},
			key:     "key_invalid",
			value:   nestedValue,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.devfilev2.UpdateAttributes(tt.key, tt.value)
			if tt.wantErr == (err == nil) {
				t.Errorf("TestUpdateAttributes error - %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				attributes, err := tt.devfilev2.GetAttributes()
				if err != nil {
					t.Errorf("TestUpdateAttributes error - %+v", err)
					return
				}
				if !reflect.DeepEqual(attributes, tt.wantAttributes) {
					t.Errorf("TestUpdateAttributes mismatch error - expected %+v, actual %+v", tt.wantAttributes, attributes)
				}
			}
		})
	}
}

func TestAddAttributes(t *testing.T) {

	nestedValue := map[string]interface{}{
		"key3.1": map[string]interface{}{
			"key3.1.1": "value3.1.1",
		},
	}

	tests := []struct {
		name           string
		devfilev2      *DevfileV2
		key            string
		value          interface{}
		wantAttributes attributes.Attributes
		wantErr        bool
	}{
		{
			name: "Schema 2.0.0 does not have attributes",
			devfilev2: &DevfileV2{
				v1alpha2.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{
						SchemaVersion: "2.0.0",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Schema 2.1.0 has attributes",
			devfilev2: &DevfileV2{
				v1alpha2.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{
						SchemaVersion: "2.1.0",
					},
					DevWorkspaceTemplateSpec: v1alpha2.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
							Attributes: attributes.Attributes{}.PutString("key1", "value1").PutString("key2", "value2"),
						},
					},
				},
			},
			key:            "key3",
			value:          nestedValue,
			wantAttributes: attributes.Attributes{}.PutString("key1", "value1").Put("key3", nestedValue, nil).PutString("key2", "value2"),
		},
		{
			name: "If Schema 2.1.0 has an attribute already present, it should overwrite",
			devfilev2: &DevfileV2{
				v1alpha2.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{
						SchemaVersion: "2.1.0",
					},
					DevWorkspaceTemplateSpec: v1alpha2.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
							Attributes: attributes.Attributes{}.PutString("key1", "value1").PutString("key2", "value2"),
						},
					},
				},
			},
			key:            "key2",
			value:          "value2new",
			wantAttributes: attributes.Attributes{}.PutString("key1", "value1").PutString("key2", "value2new"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.devfilev2.AddAttributes(tt.key, tt.value)
			if tt.wantErr == (err == nil) {
				t.Errorf("TestAddAttributes error - %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				attributes, err := tt.devfilev2.GetAttributes()
				if err != nil {
					t.Errorf("TestAddAttributes error - %+v", err)
					return
				}
				if !reflect.DeepEqual(attributes, tt.wantAttributes) {
					t.Errorf("TestAddAttributes mismatch error - expected %+v, actual %+v", tt.wantAttributes, attributes)
				}
			}
		})
	}
}
