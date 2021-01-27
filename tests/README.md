# Devfile Parser Library Tests

## About

The tests use the go language and are intended to test every apsect of the parser for every schema attribute. Some basic aspects of the tests:

* A first test (schemaTest/schema_test.go) feeds pre-created devfiles to the parser to ensure the parser can parse all attribues and return an appropriate error when the devfile contains an error. 
* A second set of test (parserTest/parser_api_test.go) creates devfile content at runtime:
    * Devfile content is randomly generated and as a result the tests are designed to run multiple times.
    * Parser functions covered:
        * Read an existing devfile.
        * Write a new devfile.
        * Modify content of a devfile.
        * Multi-threaded access to the parser.
    * The tests use the devfile schema to create a structure containing expected content for a devfile. These structures are compared with those returned by the parser. 
    * sigs.k8s.io/yaml or the parser is used to write out devfiles.
    * github.com/google/go-cmp/cmp is used to compare structures.

## Current tests:

The tests using pre-created devfiles cover all attributes and most parsing errors of the 2.0.0 schema.

The tests which generate devfiles with random content at run time currently cover the following properties and items.

* Commands: 
    * Exec
    * Composite
* Components 
    * Container
    * Volume
    
## Running the tests 

### schema_test

1. Go to directory tests/v2/schemaTest
1. Run ```go test``` or ```go test -v```
1. The test creates the following files:
    1. ```./tmp/test.log``` contains log output from the tests.
    1. ```./tmp/<property>_tests/*.yaml``` are the devfiles created for the tests. The file name matches the name of the test function which resulted in them being created.

Note: each run of the test removes the existing contents of the ```./tmp``` directory

### parser_api_test

1. Go to directory tests/v2/parserTest
1. Run ```go test``` or ```go test -v```
1. The test creates the following files: 
    1. ```./tmp/test.log``` contains log output from the tests.
    1. ```./tmp/parser_api_test/Test_*.yaml``` are the devfiles which are randomly generated at runtime. The file name matches the name of the test function which resulted in them being created.
    1. If a test detects an error when comparing properties returned by the parser with expected properties
        *  ```./tmp/parser_v200_schema_test/Test_*_<property id>_Parser.yaml``` - property as returned by the parser
        *  ```./tmp/parser_v200_schema_test/Test_*_<property id>_Test.yaml``` - property as expected by the test

Note: each run of the test removes the existing contents of the ```./tmp``` directory 

## schema_test in detail

The API tests are intended to provide a comprehensive verification of the devfile schemas. This includes:
- Ensuring every possible attribute is valid.
- Ensuring all optional attributes are indeed optional.
- Ensuring any possible specification errors are invalidated by the schema. For example:
    - Missing mandatory attributes.
    - Multiple use of a one-of attribute.
    - Attribute values of the wrong type.

### Test structure

- ```test/v2/devfiles``` : contains yaml snippets which are used to generate yaml files for the tests. The names of the sub-directories and files should reflect their purpose.
- ```test/v2/schemaTest/schema-test.go``` : the go unit test program.
- ```test/v2/json``` :  contains the json files which define the tests which the test program will run:
    - ```test-xxxxxxx.json``` : these files are the top level json files, they define the schema to verify and the test files to run.

### Adding Tests

#### Add a test for a new schema file

1. Create a new ```test/v2/json/test-<schema name>.json``` file for the schema. In the json file  specify the location of the schema to test (relative to the root directory of the repository), and a list of the existing tests to use. If the generated yaml files require a schemaVersion attribute include its value in the json file. See - *link to sample schema to be added*
1. Run the test

#### Add a test for a schema changes

1. Modify an existing yaml snippet from ``test/v2/devfiles``` or create a new one.
1. If appropriate create a new snippet for any possible error cases, for example to omit a required attribute.
1. If a new yaml snippet was created add a test which uses the snippet to the appropriate `json/xxxxxx-tests.json` file. Be careful to ensure the file name used for the test is unique for all tests - this is the name used for the yaml file which is generated for the test. For failure scenarios you may need to run the test first to set the outcome correctly.
1. If a new  `json/xxxxxx-tests.json` file is created, any existing `test-xxxxxxx.json` files must be updated to use the new file.

#### Add test for a new schema version

1. Copy and rename the `test/v2` directory for the new version, for example `test\v21`
1. Modify the copied tests as needed for the new version as decsribed above.
1. Add `test/v21/schemaTest/tmp` to the .gitignore file.
1. Run the test



## parser_api_test in detail 

### Anatomy of the tests

Each test in ```parser_api_test.go``` sets values in a test structure which defines the test to run (additions will new made for new properties as support is added):

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

### Basic principles of the tests

* Each devfile is created in a schema structure.
* Which attributes are set and the values used are randomized.
    * For example, the number of commands included in a devfile is randomly generated.
    * For example, attribute values are set to randomized strings, numbers or binary.
    * For example, a particular optional attribute has a 50% chance of being uncluded in a devfiles. 
    * Repeated tests give more variety and wider coverage. 
* Once the schema structure is complete it is written in one of two ways.
    * using the sigs.k8s.io/yaml.
    * using the parser.
* Once the devfile is created on disk the parser is used to read and validate it. 
* If editing the devfile
    * each object is retrieved, modified and written back to the parser
    * the parser is used to write the devfile to disk
    * the parser is then used to read and validate the modified devfile.
* Each array of objects in then devfile are then retrieved from the parser and compared. If this fails:
    * Each object returned by the parser is compared to the equivalent object tracked in the test.
        * if the obejcts do not match the test fails
            * Files are output with the content of each object.
     * If the parser returns more or fewer objects than expected, the test fails.   

### Updating tests

#### Files
* ```parser_api_test.go``` contains the tests
* ```test-utils.go``` provides property agnostic functions for use by the tests and other utils
* ```<property>-test-utils.go``` for example ```command-test-utils.go```, provides property related functions 

#### Adding, modifying attributes of existing properties.

In the ```<property>-test-utils.go``` files there are:
*  ```set<Item>Values``` functions.
    * for example in ```command-test-utils.go``` :
        * ```func setExecCommandValues(execCommand *schema.ExecCommand)``` 
        * ```func setCompositeCommandValues(compositeCommand *schema.CompositeCommand)``` 
* These may use utility functions to set more complex attributes. 
* Modify these functions to add/modify test for new/changed attributes.  

#### Add new item to an existing property.

For example add support for apply command to existing command support:

1. In ```command-test-utils.go```
    * add functions:
        * ```func setApplyCommandValues(applyCommand *schema.ApplyCommand)``` 
            * randomly set attribute values in the provided apply command object
        * ```func createApplyCommand() *schema.ApplyCommand```
            * creates the apply command object and calls setApplyCommandValues to add attribute values
        * follow the implementation of other similar functions.
    * modify:
       * ```func generateCommand(command *schema.Command, genericCommand *GenericCommand)```
            * add logic to call createApplyCommand if commandType indicates such.
        * ```func (devfile *TestDevfile) UpdateCommand(command *schema.Command) error```
            * add logic to call setApplyCommandValues if commandType indicates such.
1. In ```parser_v200_verify_test.go```
    * add new tests. for example:
        * Test_ApplyCommand -  CreateWithParser set to false, EditContent set to false
        * Test_CreateApplyCommand -  CreateWithParser set to true, EditContent set to false
        * Test_EditApplyCommand -  CreateWithParser set to false, EditContent set to true
        * Test_CreateEditApplyCommand -  CreateWithParser set to true, EditContent set to true
    * modify existing test to include Apply commands
        * Test_MultiCommand 
        * Test_Everything

#### Add new property

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
    * Functions general to all commands  
        * ```func generateCommand(command *schema.Command, genericCommand *GenericCommand)```
            * includes logic to call the ```create<Command-Type>Command``` function for the command-Type of the supplied command object.
        * ```func (devfile *TestDevfile) addCommand(commandType schema.CommandType) string```
            * main entry point for a test to add a command
            * maintains the array of commands in the schema structure
            * calls generateCommand() 
        * ```func (devfile *TestDevfile) UpdateCommand(command *schema.Command) error```
            * includes logic to call set<commad-type>CommandValues for each commandType.
        * ```func (devfile TestDevfile) VerifyCommands(parserCommands []schema.Command) error```
            * includes logic to compare the array of commands obtained from the parser with those created by the test. if the compare fails:
                * each individual command is compared.
                    * if a command compare fails, the parser version and test version of the command are oputput as yaml files  to the tmp directory 
                * a check is made to determine if the parser returned a command not known to the test or the pasrer omitted a command expected by the test.
1. ```test-utils.go```
    * ```func (devfile TestDevfile) Verify()``` 
        * includes code to get object from the paser and verify their content.
        * for commands code is required to: 
            1. Retrieve each command from the parser
            1. Use command Id to obtain the GenericCommand object which matches
            1. Compare the command structure returned by the parser with the command structure saved in the GenericCommand object.
    * ```func (devfile TestDevfile) EditCommands() error```
        * specific to command objects.
            1. Ensure devfile is written to disk
            1. Use parser to read devfile and get all command object
            1. For each command call:
                *  ```func (devfile *TestDevfile) UpdateCommand(command *schema.Command) error``` 
            1. When all commands have been updated, use parser to write the updated devfile to disk
1. ```parser-v200-test.go```
    * ```type TestContent struct```
        * includes an array of command types: ```CommandTypes     []schema.CommandType``` 
    * ```func Test_ExecCommand(t *testing.T)```
        1. Creates a TestContent object
        1. Adds a single entry array containg schema.ExecCommandType to the array of command types
        1. Calls runTest for a single thread test
        1. Calls runMultiThreadTest for a multi-thread test.
    * See also
        * ```func Test_<string>ExecCommand(t *testing.T)``` 
        * ```func Test_MultiCommand(t *testing.T)```
        * ```func Test_Everything(t *testing.T)```
    * Add logic to ```func runTest(testContent TestContent, t *testing.T)```
        1. Add commands to the test.
        2. Start edits of commands if required. 


#### Code flow

Create, modify and verify an exec command: 
1. parser_v200_verify_test.Test_ExecCommand
    1. parser-v200-test.runTest
        1. command-test-utils.AddCommand
            1. command-test-utils.GenerateCommand
                1. command-test-utils.createExecCommand
                    1. command-test-utils.setExecCommandValues
        1. test-utils.CreateDevfile
        1. test-utils.EditCommands
            1.  command-test-utils.UpdateCommand
                1. command-test-utils.setExecCommandValues
        1. test-utils.Verify
            1. command-test-utils.VerifyCommands

            




