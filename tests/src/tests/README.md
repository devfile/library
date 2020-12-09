# Devfile Parser Library Tests

## About

The tests use the go language and are intended to test every apsect of the parser for every schema attribute. Some basic aspects of the tests:

* A first test (parser_v200_schema_test.go) feeds pre-created devfiles to the parser to ensure the parser can parse all attribues and return an approproate error when the devfile contains an error.
* A second set of tests (parser_v200_verify_test.go) create devfile content at runtime:
    * Devfile content is randomly generated and as a result the tests are designed to run multiple times.
    * Parser functions covered:
        * Read an existing devfile.
        * Write a new devfile.
        * Modify Content of a devfile.
        * Multi-threaded access to the parser.
    * The tests use the devfile schema to create a structure containing expected content for a devfile. These structures are compared with those returned by the parser. 
    * sigs.k8s.io/yaml is used to write out devfiles.
    * github.com/google/go-cmp/cmp is used to compare structures.

## Current tests:

The tests using pre-created devfiles are complete (but update in progress due to schema changes)

The tests which generate devfiles with random content at run time currently cover the following properties and items.

* Commands: 
    * Exec
    * Composite
* Components 
    * Container
    * Volume

## Running the tests

1. Go to directory tests/src/tests
1. Run ```go test``` or ```go test -v```
1. The test creates the following files: 
    1. ```./tmp/test.log``` contains log output from the tests.
    1. ```./tmp/parser_v200_verify_test/Test_*.yaml``` are the devfiles which are randomly generated at runtime. The file name matches the name of the test function which resulted in them being created. 
    1. ```./tmp/parser_v200_schema_test/*.yaml``` are the pre-created devfiles.

Note: each run of the test removes the existing conents of the ```./tmp``` directory 

## Anatomy of parser_v200_verify_test.go test

Each test in ```parser_v200_verify_test.go``` sets values in a test structure which defines the test to run (additions will new made for new properties as support is added):

    type TestContent struct {
	    CommandTypes     []schema.CommandType
	    ComponentTypes   []schema.ComponentType
	    FileName         string
	    CreateWithParser bool
	    EditContent      bool
    }


The test then uses one (or both) of two functions to run a test

* For a single thread test:
    * ```func runTest(testContent TestContent, t *testing.T)```
* For a multi-thread test: 
    * ```func runMultiThreadTest(testContent TestContent, t *testing.T)```  

An example test:

    func Test_MultiCommand(t *testing.T) {
	    testContent := TestContent{}
	    testContent.CommandTypes = []schema.CommandType{schema.ExecCommandType, schema.CompositeCommandType}
	    testContent.CreateWithParser = true
	    testContent.EditContent = true
	    testContent.FileName = GetDevFileName()
	    runTest(testContent, t)
	    runMultiThreadTest(testContent, t)
    }

Note: ```GetDevFileName()``` is a function which returns the name of a temporary file to create which uses the name of the test file as a subdirectory and the name of test function function as file name. In this example it returns ```./tmp/parser_v200_verify_test/Test_MultiCommand```

There are also some constants which control execution of the tests:

    const numThreads = 5		// Number of threads used by multi-thread tests
    const maxCommands = 10 		// The maximum number of commands to include in a generated devfile
    const maxComponents = 10	// The maximum number of components to include in a generated devfile

## Basic principles of the tests which randomly generates devfiles

* Each devfile is created in a schema structure.
* Which attributes are set and the values used are randomized.
* A helper structure is used to track each created object, these are kept in a maps.
* Once the schema structure is complete it is written in one of two ways.
    * using the sigs.k8s.io/yaml.
    * using the parser.
* Once the devfile is created on disk the parser is used to read and validate it. 
* If editing the devfile
    * each object is retrieved, modified and written back to the pasrer
    * the parser is used to write the devfile to disk
    * the parser is then used to read and validate the modified devfile.
* Each object in then devfile is then retrieved from the parser and checked
    * The object returned by the parser is compared to the equivalent object tracked in the test.
        * if the obejcst do not match the test fails
    * If the parser returns more or fewer objects than expected, the test fails.   

## Updating tests

### Files
* ```parser_v200_verify_test.go``` contains the tests
* ```test-utils.go``` provides property agnostic functions for use by the tests and other utils
* ```<property>-test-utils.go``` for example ```command-test-utils.go```, provides property related functions 

### Adding, modifying attributes of existing properties.

In the ```<property>-test-utils.go``` files there are:
*  ```set<Item>Values``` functions.
    * for example in ```command-test-utils.go``` :
        * ```func setExecCommandValues(execCommand *schema.ExecCommand)``` 
        * ```func setCompositeCommandValues(compositeCommand *schema.CompositeCommand)``` 
* These may use utility functions to set more complex attributes. 
* Modify these functions to add/modify test for new/changed attributes.  

### Add new item to an existing property.

For example add support for apply command to existing command support:

1. In ```command-test-utils.go```
    * add functions:
        * ```func createApplyCommand() *schema.ApplyCommand```
            * creates the apply command object and calls setApplyCommandValues to add attribute values
        * ```func setApplyCommandValues(applyCommand *schema.ApplyCommand)``` 
            * randomly set attribute values in the provided apply command object
        * follow the implementation of other similar functions.
    * modify:
        * ```type GenericCommand struct```
            * add a variabale to store the address of a schema.ApplyCommand object 
            * the value is set by the test for later comparision with that returned by the parser. 
            * one of these structures is created for each command created.
        * ```func generateCommand(command *schema.Command, genericCommand *GenericCommand)```
            * add logic to call createApplyCommand if commandType indicates such.
            * store the generated ApplyCommand object in GenericCommand structure.
        * ```func (devfile *TestDevfile) UpdateCommand(command *schema.Command) error```
            * add logic to call setApplyCommandValues if commandType indicates such.
            * update the ApplyCommand object in GenericCommand structure.
        * ```func (devfile TestDevfile) VerifyCommands(commands map[string]schema.Command)```
            * add logic to compare an apply command object returned by the parser with that stored in the GenericCommand structure
1. In ```parser_v200_verify_test.go```
    * add new tests. for example:
        * Test_ApplyCommand -  CreateWithParser set to false, EditContent set to false
        * Test_CreateApplyCommand -  CreateWithParser set to true, EditContent set to false
        * Test_EditApplyCommand -  CreateWithParser set to false, EditContent set to true
        * Test_CreateEditApplyCommand -  CreateWithParser set to true, EditContent set to true
    * modify existing test to include Apply commands
        * Test_MultiCommand 
        * Test_Everything
    * add logic to ```runTest(testContent TestContent, t *testing.T)``` for creatin and editing the     

### Add new property

Using existing support for commands as an illustration, any new property support added should follow the same structure: 

1. ```command-test-utils.go```:
    * Specific to commands
    * Commands require support for 5 different command types:
        * Exec
        * Appy (to be implemented)
        * Composite
        * VSCodeLaunch (to be implemented)
        * VSCodeTask (to be implemented)
    * Each of these command-types have equivalent functions:    
        * ```func create<command-type>Command() *schema.<command-type>```
            * creates the command object and calls ```set<command-type>CommandValues``` to add attribute values
            * for example see: ```func createExecCommand(execCommand *schema.ExecCommand)```
        * ```func set<command-type>CommandValues(project-sourceProject *schema.<project-source>)```
            * sets random attributes into the provided object 
            * for example see: ```func setExecCommandValues(execCommand *schema.ExecCommand)```
    *  GenericCommand structure is created, one per command to track its content, for example:        
        ```
        type GenericCommand struct {
	        Name                   string
	        Verified               bool
	        CommandType            schema.CommandType
	        ExecCommandSchema      *schema.ExecCommand
	        SchemaCompositeCommand *schema.CompositeCommand
        }
        ``` 

    * Functions general to all commands  
        * ```func (devfile *TestDevfile) addCommand(commandType schema.CommandType) string```
            * main entry point for a test to add a command
            * maintains the array of commands in the schema structure
            * creates a GenericCommand object to store information about the created command
            * calls generateCommand() 
            * saves the GenericCommand object in a map for access later   
        * ```func generateCommand(command *schema.Command, genericCommand *GenericCommand)```
            * includes logic to call the ```create<Command-Type>Command``` function for the command-Type of the supplied command object.
            * stores a refrence to the generated command object in GenericCommand structure.
        * ```func (devfile *TestDevfile) UpdateCommand(command *schema.Command) error```
            * includes logic to call set<commad-type>CommandValues for each commandType.
            * update the equiavalent object in GenericCommand structure.
        * ```func (devfile TestDevfile) VerifyCommands(commands map[string]schema.Command)```
            * add logic to compare an each command object returned by the parser with that stored in the GenericCommand structure.
1. ```test-utils.go```
    * Includes code for all properties
    * Support for keeping a map of command objects
        * Command map is in ```type TestDevfile struct```
            * usees command id as the key.
        * functions to add and get from the map
            * ```func (devfile *TestDevfile) MapCommand(command GenericCommand)```
            * ```func (devfile *TestDevfile) GetCommand(id string) *GenericCommand```  
    * ```func (devfile *TestDevfile) CreateDevfile(useParser bool)```
        * Includes code when using the parser to create the devfile 
            * for commands: code required to add command objects to the the parser.
    * ```func (devfile TestDevfile) Verify()``` 
        * Includes code to get object from the paser and verify their content.
        * For commands code is required to: 
            1. Retrieve each command from the parser
            1. Use command id to obtain the GenericCommand object which matches
            1. compare the command structure returned by the parser with the command structure saved in the GenericCommand object.
    * ```func (devfile TestDevfile) EditCommands() error```
        * Specific to command objects.
            1. Ensure devfile is written to disk
            1. use parser to read devfile and get all command object
            1.  for each command call:
                *  ```func (devfile *TestDevfile) UpdateCommand(command *schema.Command) error``` 
            1. when all commands have been updated, use parser to write the updated devfile to disk
1. ```parser-v200-test.go```
    * ```type TestContent struct```
        * includes an array of command types: ```CommandTypes     []schema.CommandType``` 
    * ```func Test_ExecCommand(t *testing.T)```
        1. Creates a TestContent object
        1. Adds a single entry array containg schema.ExecCommandType to the array of command types
        1. Calls runTest for a single thread test
        1. Calls runMultiThreadTest for a multi-thread test.
    * See also
        *```func Test_<string>ExecCommand(t *testing.T)``` 
        *```func Test_MultiCommand(t *testing.T)```
        *```func Test_Everything(t *testing.T)```
    * add logic to ```func runTest(testContent TestContent, t *testing.T)``` includes logic
        1. Add commands to the test
        2. Starts edits of commands if required. 


#### Code flow

Create, modify and verify an exec command: 
1. parser_v200_verify_test.Test_ExecCommand
    1. parser-v200-test.runTest
        1. command-test-utils.AddCommand
            1. command-test-utils.GenerateCommand
                1.  command-test-utils.createExecCommand
                    1. command-test-utils.setExecCommandValues
            1. test-utils.MapCommand
        1. test-utils.CreateDevfile
        1. test-utils.EditCommands
            1.  command-test-utils.UpdateCommand
                1. command-test-utils.setExecCommandValues
        1. test-utils.Verify
            1. command-test-utils.VerifyCommands

            




