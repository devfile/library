package parser

import (
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	apiAttributes "github.com/devfile/api/v2/pkg/attributes"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"sigs.k8s.io/yaml"

	"github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

// WriteYamlDevfile creates a devfile.yaml file
func (d *DevfileObj) WriteYamlDevfile() error {
	err := restoreKubeCompURI(d)
	if err != nil {
		return errors.Wrapf(err, "failed to restore kubernetes component uri field")
	}
	// Encode data into YAML format
	yamlData, err := yaml.Marshal(d.Data)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal devfile object into yaml")
	}
	// Write to devfile.yaml
	fs := d.Ctx.GetFs()
	if fs == nil {
		fs = filesystem.DefaultFs{}
	}
	err = fs.WriteFile(d.Ctx.GetAbsPath(), yamlData, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to create devfile yaml file")
	}

	// Successful
	klog.V(2).Infof("devfile yaml created at: '%s'", OutputDevfileYamlPath)
	return nil
}


func restoreKubeCompURI(devObj *DevfileObj) error {
	getKubeCompOptions := common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{
			ComponentType: v1.KubernetesComponentType,
		},
	}
	kubeComponents, err := devObj.Data.GetComponents(getKubeCompOptions)
	if err != nil {
		return err
	}
	for _, kubeComp := range kubeComponents {
		var keyNotFoundErr = &apiAttributes.KeyNotFoundError{Key: KubeComponentOriginalURIKey}
		uri := kubeComp.Attributes.GetString(KubeComponentOriginalURIKey, &err)
		if err != nil && err.Error() != keyNotFoundErr.Error() {
			return err
		}
		kubeComp.Kubernetes.Uri = uri
		kubeComp.Kubernetes.Inlined = ""
		delete(kubeComp.Attributes,KubeComponentOriginalURIKey)
		err = devObj.Data.UpdateComponent(kubeComp)
		if err != nil {
			return err
		}
	}
	return nil
}