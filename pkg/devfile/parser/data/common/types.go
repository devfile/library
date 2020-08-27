package common

import (
	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
)

// DevfileMetadata metadata for devfile
type DevfileMetadata struct {

	// Name Optional devfile name
	Name string `json:"name,omitempty"`

	// Version Optional semver-compatible version
	Version string `json:"version,omitempty"`

	// Manifest optional URL to remote Deployment Manifest
	Manifest string `json:"alpha.deployment-manifest,omitempty"`
}

// DevfileComponent component specified in devfile
type DevfileComponent struct {

	// Allows adding and configuring workspace-related containers
	Container *v1.Container `json:"container,omitempty"`

	// Allows importing into the workspace the Kubernetes resources defined in a given manifest. For example this allows reusing the Kubernetes definitions used to deploy some runtime components in production.
	Kubernetes *v1.KubernetesComponent `json:"kubernetes,omitempty"`

	// Allows importing into the workspace the OpenShift resources defined in a given manifest. For example this allows reusing the OpenShift definitions used to deploy some runtime components in production.
	Openshift *v1.OpenshiftComponent `json:"openshift,omitempty"`

	// Allows specifying the definition of a volume shared by several other components
	Volume *v1.Volume `json:"volume,omitempty"`

	// Allows specifying a dockerfile to initiate build
	Dockerfile *Dockerfile `json:"dockerfile,omitempty"`
}

// Configuration
type Configuration struct {
	CookiesAuthEnabled bool   `json:"cookiesAuthEnabled,omitempty"`
	Discoverable       bool   `json:"discoverable,omitempty"`
	Path               string `json:"path,omitempty"`

	// The is the low-level protocol of traffic coming through this endpoint. Default value is "tcp"
	Protocol string `json:"protocol,omitempty"`
	Public   bool   `json:"public,omitempty"`

	// The is the URL scheme to use when accessing the endpoint. Default value is "http"
	Scheme string `json:"scheme,omitempty"`
	Secure bool   `json:"secure,omitempty"`
	Type   string `json:"type,omitempty"`
}

type Dockerfile struct {
	// Mandatory name that allows referencing the Volume component in Container volume mounts or inside a parent
	Name string `json:"name"`

	// Mandatory path to source code
	Source *Source `json:"source"`

	// Mandatory path to dockerfile
	DockerfileLocation string `json:"dockerfileLocation"`

	// Mandatory destination to registry to push built image
	Destination string `json:"destination,omitempty"`
}

type Source struct {
	// Mandatory path to local source directory folder
	SourceDir string `json:"sourceDir"`

	// Mandatory path to source repository hosted locally or on cloud
	Location string `json:"location"`
}
