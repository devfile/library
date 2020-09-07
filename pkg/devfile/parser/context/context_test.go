package parser

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPopulateFromBytes(t *testing.T) {
	tests := []struct {
		name        string
		dataFunc    func() []byte
		expectError bool
	}{
		{
			name:        "valid data passed",
			dataFunc:    validJsonRawContent200,
			expectError: false,
		},
		{
			name:        "invalid data passed",
			dataFunc:    invalidJsonRawContent200,
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := w.Write(tt.dataFunc())
				if err != nil {
					t.Error(err)
				}
			}))
			var (
				d = DevfileCtx{
					url: testServer.URL,
				}
			)
			defer testServer.Close()
			err := d.PopulateFromURL()
			if tt.expectError && err == nil {
				t.Errorf("expected error, didn't get one")
			} else if !tt.expectError && err != nil {
				t.Errorf("unexpected error '%v'", err)
			}
		})
	}
}

func TestPopulateFromInvalidURL(t *testing.T) {
	t.Run("Populate from invalid URL", func(t *testing.T) {
		var (
			d = DevfileCtx{
				url: "blah",
			}
		)

		err := d.PopulateFromURL()

		if err == nil {
			t.Errorf("expected an error, didn't get one")
		}
	})
}

func invalidJsonRawContent200() []byte {
	return []byte(InvalidDevfileContent)
}
