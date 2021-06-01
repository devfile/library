# Devfile Parser Library Tests

## About

The tests use the go language and are intended to test every aspect of the parser for every schema attribute. Some basic aspects of the tests:

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
    - `tests/v2/libraryTest/library-test.go`: The go unit test program
    - `tests/v2/utils/library/*-utils.go` : utilites, used by the test, which contain functions uniqiue to the library tests. Mostly contain the code to modify and check devfile content.
* From the [api respository](https://github.com/devfile/api/tree/master/test/v200/utils/common)
    - `tests/v200/utils/common/*-utils.go` : utilites, used by the test, which are also used by the api tests. Mostly contain the code to generate valid devfile content.

## Running the tests locally

1. Go to the ```/library``` directory 
1. Run ```Make test```
1. The test creates the following files:
    1. ```./tmp/test.log``` contains log output from the tests.
    1. ```./tmp/library_test/Test_*.yaml``` are the devfiles which are randomly generated at runtime. The file name matches the name of the test function which resulted in them being created.
    1. If a test detects an error when comparing properties returned by the parser with expected properties
        *  ```./tmp/library_test/Test_*_<property id>_Parser.yaml``` - property as returned by the parser
        *  ```./tmp/library_test/Test_*_<property id>_Test.yaml``` - property as expected by the test
    1.  ```tests/v2/lib-test-coverage.html``` which is the test coverage report for the test run.  You can open this file up in a browser and inspect the results to determine the gaps you may have in testing
    
Note: each run of the test removes the existing contents of the ```./tmp``` directory

## Viewing test results from a workflow

The tests run automatically with every PR or Push action.  You can see the results in the `devfile/library` repo's `Actions` view:

1.  Select the `Validate PRs` workflow and click on the PR you're interested in
1.  To view the console output, select the `Build` job and expand the `Run Go Tests` step.  This will give you a summary of the tests that were executed and their respective status 
1.  To view the test coverage report, click on the `Summary` page and you should see an `Artifacts` section with the `lib-test-coverage-html` file available for download.



