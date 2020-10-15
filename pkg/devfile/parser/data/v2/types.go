package v2

import (
	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
)

type DevfileV2 struct {
	v1.Devfile
}

// // NewDevfileData returns relevant devfile struct for the provided API version
// func NewDevfileData(version string) (obj DevfileData, err error) {

// 	// Fetch devfile struct type from map
// 	devfileType, ok := apiVersionToDevfileStruct[supportedApiVersion(version)]
// 	if !ok {
// 		errMsg := fmt.Sprintf("devfile type not present for apiVersion '%s'", version)
// 		return obj, fmt.Errorf(errMsg)
// 	}

// 	return reflect.New(devfileType).Interface().(DevfileData), nil
// }
