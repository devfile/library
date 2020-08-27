package validate

import (
	"fmt"
	"reflect"
	"testing"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/parser/pkg/devfile/parser/data/common"
)

func TestValidateComponents(t *testing.T) {

	t.Run("No components present", func(t *testing.T) {

		// Empty components
		components := []common.DevfileComponent{}

		got := ValidateComponents(components)
		want := fmt.Errorf(ErrorNoComponents)

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got: '%v', want: '%v'", got, want)
		}
	})

	t.Run("Container type of component present", func(t *testing.T) {

		components := []common.DevfileComponent{
			{
				Container: &v1.Container{
					Name: "container",
				},
			},
		}

		got := ValidateComponents(components)

		if got != nil {
			t.Errorf("Not expecting an error: '%v'", got)
		}
	})
}
