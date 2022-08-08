package parser

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/devfile/library/pkg/util"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	k8yaml "sigs.k8s.io/yaml"
)

func TestReadKubernetesYaml(t *testing.T) {
	const serverIP = "127.0.0.1:9080"
	var data []byte

	fs := filesystem.DefaultFs{}
	absPath, err := util.GetAbsPath("../../../tests/yamls/resources.yaml")
	if err != nil {
		t.Error(err)
		return
	}

	data, err = fs.ReadFile(absPath)
	if err != nil {
		t.Error(err)
		return
	}

	// Mocking the YAML file endpoint on a very basic level
	testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err = w.Write(data)
		if err != nil {
			t.Errorf("Unexpected error while writing data: %v", err)
		}
	}))
	// create a listener with the desired port.
	l, err := net.Listen("tcp", serverIP)
	if err != nil {
		t.Errorf("Unexpected error while creating listener: %v", err)
		return
	}

	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	testServer.Listener.Close()
	testServer.Listener = l

	testServer.Start()
	defer testServer.Close()

	tests := []struct {
		name                string
		src                 YamlSrc
		fs                  filesystem.Filesystem
		wantErr             bool
		wantDeploymentNames []string
		wantServiceNames    []string
		wantRouteNames      []string
		wantOtherNames      []string
	}{
		{
			name: "Read the YAML from the URL",
			src: YamlSrc{
				URL: "http://" + serverIP,
			},
			fs:                  filesystem.DefaultFs{},
			wantDeploymentNames: []string{"deploy-sample"},
			wantServiceNames:    []string{"service-sample"},
			wantRouteNames:      []string{"route-sample"},
			wantOtherNames:      []string{"pvc-sample"},
		},
		{
			name: "Read the YAML from the Path",
			src: YamlSrc{
				Path: "../../../tests/yamls/resources.yaml",
			},
			fs:                  filesystem.DefaultFs{},
			wantDeploymentNames: []string{"deploy-sample"},
			wantServiceNames:    []string{"service-sample"},
			wantRouteNames:      []string{"route-sample"},
			wantOtherNames:      []string{"pvc-sample"},
		},
		{
			name: "Read the YAML from the Data",
			src: YamlSrc{
				Data: data,
			},
			fs:                  filesystem.DefaultFs{},
			wantDeploymentNames: []string{"deploy-sample"},
			wantServiceNames:    []string{"service-sample"},
			wantRouteNames:      []string{"route-sample"},
			wantOtherNames:      []string{"pvc-sample"},
		},
		{
			name: "Bad URL",
			src: YamlSrc{
				URL: "http://badurl",
			},
			fs:      filesystem.DefaultFs{},
			wantErr: true,
		},
		{
			name: "Bad Path",
			src: YamlSrc{
				Path: "$%^&",
			},
			fs:      filesystem.DefaultFs{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployments, services, routes, others, err := ReadKubernetesYaml(tt.src, tt.fs)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: %v", err)
				return
			}
			for _, deploy := range deployments {
				assert.Contains(t, tt.wantDeploymentNames, deploy.Name)
			}
			for _, svc := range services {
				assert.Contains(t, tt.wantServiceNames, svc.Name)
			}
			for _, route := range routes {
				assert.Contains(t, tt.wantRouteNames, route.Name)
			}
			for _, other := range others {
				pvc := corev1.PersistentVolumeClaim{}
				err = k8yaml.Unmarshal(other, &pvc)
				if err != nil {
					t.Error(err)
				}
				assert.Contains(t, tt.wantOtherNames, pvc.Name)
			}
		})
	}
}
