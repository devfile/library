package parser

import (
	"bytes"
	"io"

	"github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/devfile/library/pkg/util"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	k8yaml "sigs.k8s.io/yaml"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// YamlSrc specifies the src of the yaml in either Path, URL or Data format
type YamlSrc struct {
	// Path is a relative or absolute yaml path.
	Path string
	// URL is the URL address of the specific yaml.
	URL string
	// Data is the yaml content in []byte format.
	Data []byte
}

// ReadKubernetesYaml reads a yaml Kubernetes file from either the Path, URL or Data provided.
// It returns Deployments, Services, Routes resources as the primary Kubernetes resources.
// Other Kubernetes resources are returned as []byte type. Consumers interested in other Kubernetes resources
// are expected to Unmarshal it to the struct of the respective resource.
func ReadKubernetesYaml(src YamlSrc, fs filesystem.Filesystem) ([]appsv1.Deployment, []corev1.Service, []routev1.Route, [][]byte, error) {

	var data []byte
	var err error

	if src.URL != "" {
		data, err = util.DownloadFileInMemory(src.URL)
		if err != nil {
			return nil, nil, nil, nil, errors.Wrapf(err, "failed to download file %q", src.URL)
		}
	} else if src.Path != "" {
		absPath, err := util.GetAbsPath(src.Path)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		data, err = fs.ReadFile(absPath)
		if err != nil {
			return nil, nil, nil, nil, errors.Wrapf(err, "failed to read yaml from path %q", src.Path)
		}
	} else if len(src.Data) > 0 {
		data = src.Data
	}

	var values []interface{}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	for {
		var value interface{}
		err = dec.Decode(&value)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, nil, nil, nil, err
		}
		values = append(values, value)
	}

	var deployments []appsv1.Deployment
	var services []corev1.Service
	var routes []routev1.Route
	var otherResources [][]byte

	for _, value := range values {
		var deployment appsv1.Deployment
		var service corev1.Service
		var route routev1.Route

		byteData, err := k8yaml.Marshal(value)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		kubernetesMap := value.(map[string]interface{})
		kind := kubernetesMap["kind"]

		switch kind {
		case "Deployment":
			err = k8yaml.Unmarshal(byteData, &deployment)
			deployments = append(deployments, deployment)
		case "Service":
			err = k8yaml.Unmarshal(byteData, &service)
			services = append(services, service)
		case "Route":
			err = k8yaml.Unmarshal(byteData, &route)
			routes = append(routes, route)
		default:
			otherResources = append(otherResources, byteData)
		}

		if err != nil {
			return nil, nil, nil, nil, err
		}
	}

	return deployments, services, routes, otherResources, nil
}
