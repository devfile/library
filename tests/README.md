# Devfile Parser Library Tests

## About

The tests use the go language and are intended to test every apsect of the parser for every schema attribute. Some basic aspects of the tests:

* A first test (parser_v200_schema_test.go) feeds pre-created devfiles to the parser to ensure the parser can parse all attribues and return an appropriate error when the devfile contains an error. This test is not currently available. 
* A second set of tests (parser_v200_verify_test.go) create devfile content at runtime:
    * Devfile content is randomly generated and as a result the tests are designed to run multiple times.
    * Parser functions covered:
        * Read an existing devfile.
        * Write a new devfile.
        * Modify content of a devfile.
        * Multi-threaded access to the parser.
    * The tests use the devfile schema to create a structure containing expected content for a devfile. These structures are compared with those returned by the parser. 
    * sigs.k8s.io/yaml is used to write out devfiles.
    * github.com/google/go-cmp/cmp is used to compare structures.

## Current tests:

The tests using pre-created devfiles are not currently available (update in progress due to schema changes)

The tests which generate devfiles with random content at run time currently cover the following properties and items.

* Commands: 
    * Exec
    * Composite
    * Apply
* Components 
    * Container
    * Volume
* Projects
* Starter Projects
    
## Test structure

* From this repository
    - `test/v2/libraryTest/library-test.go`: The go unit test program
    - `test/v2/utils/library/*-utils.go` : utilites, used by the test, which contain functions uniqiue to the library tests. Mostly contain the code to modify and check devfile content.
* From the [api respository](https://github.com/devfile/api/tree/master/test/v200/utils/common)
    - `test/v200/utils/common/*-utils.go` : utilites, used by the test, which are also used by the api tests. Mostly contain the code to generate valid devfile content.

## Running the tests

1. Go to directory /tests/v2/libraryTest
1. Run ```go test``` or ```go test -v```
1. The test creates the following files:
    1. ```./tmp/test.log``` contains log output from the tests.
    1. ```./tmp/library_test/Test_*.yaml``` are the devfiles which are randomly generated at runtime. The file name matches the name of the test function which resulted in them being created.
    1. If a test detects an error when comparing properties returned by the parser with expected properties
        *  ```./tmp/library_test/Test_*_<property id>_Parser.yaml``` - property as returned by the parser
        *  ```./tmp/library_test/Test_*_<property id>_Test.yaml``` - property as expected by the test

Note: each run of the test removes the existing conents of the ```./tmp``` directory



