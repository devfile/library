package utils

import (
	"fmt"

	schema "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

var Exposures = [...]schema.EndpointExposure{schema.PublicEndpointExposure, schema.InternalEndpointExposure, schema.NoneEndpointExposure}

// getRandomExposure returns a random exposure value
func getRandomExposure() schema.EndpointExposure {
	return Exposures[GetRandomNumber(len(Exposures))-1]
}

//var Protocols = [...]schema.EndpointProtocol{schema.HTTPEndpointProtocol, schema.HTTPSEndpointProtocol, schema.WSEndpointProtocol, schema.WSSEndpointProtocol, schema.TCPEndpointProtocol, schema.UDPEndpointProtocol}
var Protocols = [...]schema.EndpointProtocol{schema.HTTPEndpointProtocol, schema.WSEndpointProtocol, schema.TCPEndpointProtocol, schema.UDPEndpointProtocol}

// getRandomProtocol returns a random protocol value
func getRandomProtocol() schema.EndpointProtocol {
	return Protocols[GetRandomNumber(len(Protocols))-1]
}

// CreateEndpoints createa and returns a randon numebr of endpoints in a schema structure
func CreateEndpoints() []schema.Endpoint {

	numEndpoints := GetRandomNumber(5)
	endpoints := make([]schema.Endpoint, numEndpoints)

	for i := 0; i < numEndpoints; i++ {

		endpoint := schema.Endpoint{}

		endpoint.Name = GetRandomString(GetRandomNumber(15)+5, false)
		LogInfoMessage(fmt.Sprintf("   ....... add endpoint %d name  : %s", i, endpoint.Name))

		endpoint.TargetPort = GetRandomNumber(9999)
		LogInfoMessage(fmt.Sprintf("   ....... add endpoint %d targetPort: %d", i, endpoint.TargetPort))

		if GetBinaryDecision() {
			endpoint.Exposure = getRandomExposure()
			LogInfoMessage(fmt.Sprintf("   ....... add endpoint %d exposure: %s", i, endpoint.Exposure))
		}

		if GetBinaryDecision() {
			endpoint.Protocol = getRandomProtocol()
			LogInfoMessage(fmt.Sprintf("   ....... add endpoint %d protocol: %s", i, endpoint.Protocol))
		}

		endpoint.Secure = GetBinaryDecision()
		LogInfoMessage(fmt.Sprintf("   ....... add endpoint %d secure: %t", i, endpoint.Secure))

		if GetBinaryDecision() {
			endpoint.Path = "/Path_" + GetRandomString(GetRandomNumber(10)+3, false)
			LogInfoMessage(fmt.Sprintf("   ....... add endpoint %d path: %s", i, endpoint.Path))
		}

		endpoints[i] = endpoint

	}

	return endpoints
}
