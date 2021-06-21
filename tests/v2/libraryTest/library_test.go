package api

import (
	"testing"

	schema "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	commonUtils "github.com/devfile/api/v2/test/v200/utils/common"
	libraryUtils "github.com/devfile/library/tests/v2/utils/library"
)

// TestContent - structure used by a test to configure the tests to run
type TestContent struct {
	CommandTypes   []schema.CommandType
	ComponentTypes []schema.ComponentType
	FileName       string
	EditContent    bool
}

func Test_ExecCommand(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.ExecCommandType}
	testContent.AddEvents = commonUtils.GetBinaryDecision()
	testContent.EditContent = false
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}
func Test_ExecCommandEdit(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.ExecCommandType}
	testContent.AddEvents = commonUtils.GetBinaryDecision()
	testContent.EditContent = true
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_ApplyCommand(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.ApplyCommandType}
	testContent.AddEvents = commonUtils.GetBinaryDecision()
	testContent.EditContent = false
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_ApplyCommandEdit(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.ApplyCommandType}
	testContent.AddEvents = commonUtils.GetBinaryDecision()
	testContent.EditContent = true
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}

func Test_CompositeCommand(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.CompositeCommandType}
	testContent.AddEvents = commonUtils.GetBinaryDecision()
	testContent.EditContent = false
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)
}
func Test_CompositeCommandEdit(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.CompositeCommandType}
	testContent.AddEvents = commonUtils.GetBinaryDecision()
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
	testContent.AddEvents = commonUtils.GetBinaryDecision()
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

func Test_Everything(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.ExecCommandType, schema.CompositeCommandType, schema.ApplyCommandType}
	testContent.ComponentTypes = []schema.ComponentType{schema.ContainerComponentType, schema.KubernetesComponentType, schema.OpenshiftComponentType, schema.VolumeComponentType}
	testContent.ProjectTypes = []schema.ProjectSourceType{schema.GitProjectSourceType, schema.ZipProjectSourceType}
	testContent.StarterProjectTypes = []schema.ProjectSourceType{schema.GitProjectSourceType, schema.ZipProjectSourceType}
	testContent.AddEvents = true
	testContent.EditContent = false
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)

}

func Test_EverythingEdit(t *testing.T) {
	testContent := commonUtils.TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.ExecCommandType, schema.CompositeCommandType, schema.ApplyCommandType}
	testContent.ComponentTypes = []schema.ComponentType{schema.ContainerComponentType, schema.KubernetesComponentType, schema.OpenshiftComponentType, schema.VolumeComponentType}
	testContent.ProjectTypes = []schema.ProjectSourceType{schema.GitProjectSourceType, schema.ZipProjectSourceType}
	testContent.StarterProjectTypes = []schema.ProjectSourceType{schema.GitProjectSourceType, schema.ZipProjectSourceType}
	testContent.AddEvents = true
	testContent.EditContent = true
	testContent.FileName = commonUtils.GetDevFileName()
	libraryUtils.RunTest(testContent, t)
	libraryUtils.RunMultiThreadTest(testContent, t)

}
