package tests

import (
	"testing"

	schema "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
)

func Test_ExecCommand(t *testing.T) {
	testContent := TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.ExecCommandType}
	testContent.CreateWithParser = false
	testContent.EditContent = false
	testContent.FileName = GetDevFileName()
	err := RunTest(testContent)
	if err != nil {
		t.Fatalf("ERROR : %v", err)
	}
	// RunMultiThreadTest(testContent)
}
func Test_ExecCommandEdit(t *testing.T) {
	testContent := TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.ExecCommandType}
	testContent.CreateWithParser = false
	testContent.EditContent = true
	testContent.FileName = GetDevFileName()
	err := RunTest(testContent)
	if err != nil {
		t.Fatalf("ERROR : %v", err)
	}
	//RunMultiThreadTest(testContent)
}

func Test_ExecCommandParserCreate(t *testing.T) {
	testContent := TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.ExecCommandType}
	testContent.CreateWithParser = true
	testContent.EditContent = false
	testContent.FileName = GetDevFileName()
	err := RunTest(testContent)
	if err != nil {
		t.Fatalf("ERROR : %v", err)
	}
	//RunMultiThreadTest(testContent)
}

func Test_ExecCommandEditParserCreate(t *testing.T) {
	testContent := TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.ExecCommandType}
	testContent.CreateWithParser = true
	testContent.EditContent = true
	testContent.FileName = GetDevFileName()
	err := RunTest(testContent)
	if err != nil {
		t.Fatalf("ERROR : %v", err)
	}
	//RunMultiThreadTest(testContent)
}

func Test_CompositeCommand(t *testing.T) {
	testContent := TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.CompositeCommandType}
	testContent.CreateWithParser = false
	testContent.EditContent = false
	testContent.FileName = GetDevFileName()
	err := RunTest(testContent)
	if err != nil {
		t.Fatalf("ERROR : %v", err)
	}
	//RunMultiThreadTest(testContent)
}
func Test_CompositeCommandEdit(t *testing.T) {
	testContent := TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.CompositeCommandType}
	testContent.CreateWithParser = false
	testContent.EditContent = true
	testContent.FileName = GetDevFileName()
	err := RunTest(testContent)
	if err != nil {
		t.Fatalf("ERROR : %v", err)
	}
	//RunMultiThreadTest(testContent)
}

func Test_CompositeCommandParserCreate(t *testing.T) {
	testContent := TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.CompositeCommandType}
	testContent.CreateWithParser = true
	testContent.EditContent = false
	testContent.FileName = GetDevFileName()
	err := RunTest(testContent)
	if err != nil {
		t.Fatalf("ERROR : %v", err)
	}
	//RunMultiThreadTest(testContent)
}

func Test_CompositeCommandEditParserCreate(t *testing.T) {
	testContent := TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.CompositeCommandType}
	testContent.CreateWithParser = true
	testContent.EditContent = true
	testContent.FileName = GetDevFileName()
	err := RunTest(testContent)
	if err != nil {
		t.Fatalf("ERROR : %v", err)
	}
	//RunMultiThreadTest(testContent)
}

func Test_MultiCommand(t *testing.T) {
	testContent := TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.ExecCommandType, schema.CompositeCommandType}
	testContent.CreateWithParser = true
	testContent.EditContent = true
	testContent.FileName = GetDevFileName()
	err := RunTest(testContent)
	if err != nil {
		t.Fatalf("ERROR : %v", err)
	}
	//RunMultiThreadTest(testContent)
}

func Test_ContainerComponent(t *testing.T) {
	testContent := TestContent{}
	testContent.ComponentTypes = []schema.ComponentType{schema.ContainerComponentType}
	testContent.CreateWithParser = false
	testContent.EditContent = false
	testContent.FileName = GetDevFileName()
	err := RunTest(testContent)
	if err != nil {
		t.Fatalf("ERROR : %v", err)
	}
	//RunMultiThreadTest(testContent)
}

func Test_ContainerComponentEdit(t *testing.T) {
	testContent := TestContent{}
	testContent.ComponentTypes = []schema.ComponentType{schema.ContainerComponentType}
	testContent.CreateWithParser = false
	testContent.EditContent = true
	testContent.FileName = GetDevFileName()
	err := RunTest(testContent)
	if err != nil {
		t.Fatalf("ERROR : %v", err)
	}
	//RunMultiThreadTest(testContent)
}

func Test_ContainerComponentCreateWithParser(t *testing.T) {
	testContent := TestContent{}
	testContent.ComponentTypes = []schema.ComponentType{schema.ContainerComponentType}
	testContent.CreateWithParser = true
	testContent.EditContent = false
	testContent.FileName = GetDevFileName()
	err := RunTest(testContent)
	if err != nil {
		t.Fatalf("ERROR : %v", err)
	}
	//RunMultiThreadTest(testContent)
}

func Test_ContainerComponentEditCreateWithParser(t *testing.T) {
	testContent := TestContent{}
	testContent.ComponentTypes = []schema.ComponentType{schema.ContainerComponentType}
	testContent.CreateWithParser = true
	testContent.EditContent = true
	testContent.FileName = GetDevFileName()
	err := RunTest(testContent)
	if err != nil {
		t.Fatalf("ERROR : %v", err)
	}
	//RunMultiThreadTest(testContent)
}

func Test_VolumeComponent(t *testing.T) {
	testContent := TestContent{}
	testContent.ComponentTypes = []schema.ComponentType{schema.VolumeComponentType}
	testContent.CreateWithParser = false
	testContent.EditContent = false
	testContent.FileName = GetDevFileName()
	err := RunTest(testContent)
	if err != nil {
		t.Fatalf("ERROR : %v", err)
	}
	//RunMultiThreadTest(testContent)
}

func Test_VolumeComponentEdit(t *testing.T) {
	testContent := TestContent{}
	testContent.ComponentTypes = []schema.ComponentType{schema.VolumeComponentType}
	testContent.CreateWithParser = false
	testContent.EditContent = true
	testContent.FileName = GetDevFileName()
	err := RunTest(testContent)
	if err != nil {
		t.Fatalf("ERROR : %v", err)
	}
	//RunMultiThreadTest(testContent)
}

func Test_VolumeComponentCreateWithParser(t *testing.T) {
	testContent := TestContent{}
	testContent.ComponentTypes = []schema.ComponentType{schema.VolumeComponentType}
	testContent.CreateWithParser = true
	testContent.EditContent = false
	testContent.FileName = GetDevFileName()
	err := RunTest(testContent)
	if err != nil {
		t.Fatalf("ERROR : %v", err)
	}
	//RunMultiThreadTest(testContent)
}

func Test_VolumeComponentEditCreateWithParser(t *testing.T) {
	testContent := TestContent{}
	testContent.ComponentTypes = []schema.ComponentType{schema.VolumeComponentType}
	testContent.CreateWithParser = true
	testContent.EditContent = true
	testContent.FileName = GetDevFileName()
	err := RunTest(testContent)
	if err != nil {
		t.Fatalf("ERROR : %v", err)
	}
	//RunMultiThreadTest(testContent)
}

func Test_MultiComponent(t *testing.T) {
	testContent := TestContent{}
	testContent.ComponentTypes = []schema.ComponentType{schema.ContainerComponentType, schema.VolumeComponentType}
	testContent.CreateWithParser = true
	testContent.EditContent = true
	testContent.FileName = GetDevFileName()
	err := RunTest(testContent)
	if err != nil {
		t.Fatalf("ERROR : %v", err)
	}
	//RunMultiThreadTest(testContent)
}

func Test_Everything(t *testing.T) {
	testContent := TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.ExecCommandType, schema.CompositeCommandType}
	testContent.ComponentTypes = []schema.ComponentType{schema.ContainerComponentType, schema.VolumeComponentType}
	testContent.CreateWithParser = true
	testContent.EditContent = true
	testContent.FileName = GetDevFileName()
	err := RunTest(testContent)
	if err != nil {
		t.Fatalf("ERROR : %v", err)
	}
	//RunMultiThreadTest(testContent)
}
