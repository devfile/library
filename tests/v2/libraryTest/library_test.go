package api

import (
	"context"
	"testing"

	schema "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	commonUtils "github.com/devfile/api/v2/test/v200/utils/common"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/testingutil"
	libraryUtils "github.com/devfile/library/tests/v2/utils/library"
	kubev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_ExecCommand(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.ExecCommandType}
	testContent.EditContent = false
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}
func Test_ExecCommandEdit(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.ExecCommandType}
	testContent.EditContent = true
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_ApplyCommand(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.ApplyCommandType}
	testContent.EditContent = false
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_ApplyCommandEdit(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.ApplyCommandType}
	testContent.EditContent = true
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_CompositeCommand(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.CompositeCommandType}
	testContent.EditContent = false
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}
func Test_CompositeCommandEdit(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.CompositeCommandType}
	testContent.EditContent = true
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_MultiCommand(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.ExecCommandType,
		schema.CompositeCommandType,
		schema.ApplyCommandType}
	testContent.EditContent = true
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_ContainerComponent(t *testing.T) {

	testContent := commonUtils.TestContent{}
	testContent.ComponentTypes = []schema.ComponentType{schema.ContainerComponentType}
	testContent.EditContent = false
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_ContainerComponentEdit(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.ComponentTypes = []schema.ComponentType{schema.ContainerComponentType}
	testContent.EditContent = true
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_KubernetesComponent(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.ComponentTypes = []schema.ComponentType{schema.KubernetesComponentType}
	testContent.EditContent = false
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_KubernetesComponentEdit(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.ComponentTypes = []schema.ComponentType{schema.KubernetesComponentType}
	testContent.EditContent = true
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_OpenshiftComponent(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.ComponentTypes = []schema.ComponentType{schema.OpenshiftComponentType}
	testContent.EditContent = false
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_OpenshiftComponentEdit(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.ComponentTypes = []schema.ComponentType{schema.OpenshiftComponentType}
	testContent.EditContent = true
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_VolumeComponent(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.ComponentTypes = []schema.ComponentType{schema.VolumeComponentType}
	testContent.EditContent = false
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_VolumeComponentEdit(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.ComponentTypes = []schema.ComponentType{schema.VolumeComponentType}
	testContent.EditContent = true
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_MultiComponent(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.ComponentTypes = []schema.ComponentType{schema.ContainerComponentType, schema.KubernetesComponentType, schema.OpenshiftComponentType, schema.VolumeComponentType}
	testContent.EditContent = true
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_Projects(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.ProjectTypes = []schema.ProjectSourceType{schema.GitProjectSourceType, schema.ZipProjectSourceType}
	testContent.EditContent = true
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_StarterProjects(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.StarterProjectTypes = []schema.ProjectSourceType{schema.GitProjectSourceType, schema.ZipProjectSourceType}
	testContent.EditContent = true
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_Events(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.AddEvents = true
	testContent.EditContent = false
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_EventsEdit(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.AddEvents = true
	testContent.EditContent = true
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_Metadata(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.AddMetaData = true
	testContent.EditContent = false
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_MetadataEdit(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.AddMetaData = true
	testContent.EditContent = true
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_Parent_Local_URI(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.AddParent = true
	testContent.EditContent = false
	testContent.FileName = "Test_Parent_LocalURI.yaml"
	//copy the parent and main devfile from devfiles/samples
	libraryUtils.CopyDevfileSamples(t, []string{testContent.FileName, "Parent.yaml"})
	libraryUtils.RunParentTest(testContent, t)
	libraryUtils.RunMultiThreadedParentTest(testContent, t)
}

//Create kube client and context and set as ParserArgs for Parent Kubernetes reference test.  Corresponding main devfile is ../devfile/samples/TestParent_KubeCRD.yaml
func setClientAndContextParserArgs() *parser.ParserArgs {
	name := "testkubeparent1"
	parentSpec := schema.DevWorkspaceTemplateSpec{
		DevWorkspaceTemplateSpecContent: schema.DevWorkspaceTemplateSpecContent{
			Commands: []schema.Command{
				{
					Id: "applycommand",
					CommandUnion: schema.CommandUnion{
						Apply: &schema.ApplyCommand{
							Component: "devbuild",
							LabeledCommand: schema.LabeledCommand{
								Label: "testcontainerparent",
								BaseCommand: schema.BaseCommand{
									Group: &schema.CommandGroup{
										Kind:      schema.TestCommandGroupKind,
										IsDefault: true,
									},
								},
							},
						},
					},
				},
			},
			Components: []schema.Component{
				{
					Name: "devbuild",
					ComponentUnion: schema.ComponentUnion{
						Container: &schema.ContainerComponent{
							Container: schema.Container{
								Image: "quay.io/nodejs-12",
							},
						},
					},
				},
			},
			Projects: []schema.Project{
				{
					Name: "parentproject",
					ProjectSource: schema.ProjectSource{
						Git: &schema.GitProjectSource{
							GitLikeProjectSource: schema.GitLikeProjectSource{
								CheckoutFrom: &schema.CheckoutFrom{
									Revision: "master",
									Remote:   "origin",
								},
								Remotes: map[string]string{"origin": "https://github.com/spring-projects/spring-petclinic.git"},
							},
						},
					},
				},
				{
					Name: "parentproject2",
					ProjectSource: schema.ProjectSource{
						Zip: &schema.ZipProjectSource{
							Location: "https://github.com/spring-projects/spring-petclinic.zip",
						},
					},
				},
			},
			StarterProjects: []schema.StarterProject{
				{
					Name: "parentstarterproject",
					ProjectSource: schema.ProjectSource{
						Git: &schema.GitProjectSource{
							GitLikeProjectSource: schema.GitLikeProjectSource{
								CheckoutFrom: &schema.CheckoutFrom{
									Revision: "main",
									Remote:   "origin",
								},
								Remotes: map[string]string{"origin": "https://github.com/spring-projects/spring-petclinic.git"},
							},
						},
					},
				},
			},
			Attributes: attributes.Attributes{}.FromStringMap(map[string]string{"category": "parentDevfile", "title": "This is a parent devfile"}),
			Variables:  map[string]string{"version": "2.0.0", "tag": "parent"},
		},
	}
	testK8sClient := &testingutil.FakeK8sClient{
		DevWorkspaceResources: map[string]schema.DevWorkspaceTemplate{
			name: {
				TypeMeta: kubev1.TypeMeta{
					APIVersion: "2.1.0",
				},
				Spec: parentSpec,
			},
		},
	}
	parserArgs := parser.ParserArgs{}
	parserArgs.K8sClient = testK8sClient
	parserArgs.Context = context.Background()
	return &parserArgs
}

func Test_Parent_KubeCRD(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.AddParent = true
	testContent.EditContent = false
	testContent.FileName = "Test_Parent_KubeCRD.yaml"
	parserArgs := setClientAndContextParserArgs()
	libraryUtils.CopyDevfileSamples(t, []string{testContent.FileName})
	libraryUtils.SetParserArgs(*parserArgs)
	libraryUtils.RunParentTest(testContent, t)
	libraryUtils.RunMultiThreadedParentTest(testContent, t)
}

func Test_Parent_RegistryURL(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.AddParent = true
	testContent.EditContent = false
	testContent.FileName = "Test_Parent_RegistryURL.yaml"
	libraryUtils.CopyDevfileSamples(t, []string{testContent.FileName})
	libraryUtils.RunParentTest(testContent, t)
	libraryUtils.RunMultiThreadedParentTest(testContent, t)
}

func Test_Everything(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.CommandTypes = commonUtils.CommandTypes
	testContent.ComponentTypes = commonUtils.ComponentTypes
	testContent.ProjectTypes = commonUtils.ProjectSourceTypes
	testContent.StarterProjectTypes = commonUtils.ProjectSourceTypes
	testContent.AddEvents = true
	testContent.AddMetaData = true
	testContent.EditContent = false
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)

}

func Test_EverythingEdit(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.CommandTypes = commonUtils.CommandTypes
	testContent.ComponentTypes = commonUtils.ComponentTypes
	testContent.ProjectTypes = commonUtils.ProjectSourceTypes
	testContent.StarterProjectTypes = commonUtils.ProjectSourceTypes
	testContent.AddEvents = true
	testContent.AddMetaData = true
	testContent.EditContent = true
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)

}
