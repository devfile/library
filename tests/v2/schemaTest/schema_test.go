package tests

import (
	"fmt"
	"github.com/devfile/library/tests/v2/utils"
	"strings"
	"testing"

	devfilepkg "github.com/devfile/library/pkg/devfile"
)

const (
	jsonDir      = "../json/"
	logErrorOnly = false
)

// Test_Schema: Feeds a variety of devfiles to th parse to check it parses and validates the devfile correctly.
func Test_Schema(t *testing.T) {

	passTests := 0
	totalTests := 0
	skippedTests := 0

	// Get all of the test defined in teh json file
	testsToRun, failure := utils.GetAllTests(jsonDir)
	if failure {
		t.Error(utils.LogErrorMessage(fmt.Sprintf("At least one error occurred reading tests from %s, see log for details", jsonDir)))
		totalTests++
	}

	if len(testsToRun) == 0 {
		t.Fatal(utils.LogErrorMessage("No tests were found!"))
	}

	// Run each test
	for _, testToRun := range testsToRun {

		totalTests++

		// Test may be disabled pending changes to the schema
		if testToRun.Disabled {
			t.Log(utils.LogInfoMessage(fmt.Sprintf("SKIP : %s", testToRun.FileName)))
			skippedTests++
			continue
		}

		// Create the devfile to be used for the test on disk
		yamlFile, err := testToRun.CreateTestYaml()
		if err != nil {
			t.Error(utils.LogErrorMessage(fmt.Sprintf("Error creating yaml file for test : %v", err)))
			continue
		}

		utils.LogInfoMessage(fmt.Sprintf("Parse file : " + yamlFile))

		// Parse and validate the devfile
		_, err = devfilepkg.ParseAndValidate(yamlFile)
		if err != nil {
			if testToRun.ExpectOutcome == "PASS" {
				t.Error(utils.LogErrorMessage(fmt.Sprintf("  FAIL : %s : Validate failure : %v", yamlFile, err)))
			} else if testToRun.ExpectOutcome == "" {
				t.Error(utils.LogErrorMessage(fmt.Sprintf("  FAIL : %s : No expected ouctome was set : %s  got : %v", yamlFile, testToRun.ExpectOutcome, err)))
			} else if !strings.Contains(err.Error(), testToRun.ExpectOutcome) {
				t.Error(utils.LogErrorMessage(fmt.Sprintf("  FAIL : %s : Did not fail as expected : %s  got : %v", yamlFile, testToRun.ExpectOutcome, err)))
			} else {
				passTests++
				if !logErrorOnly {
					t.Log(utils.LogInfoMessage(fmt.Sprintf("PASS : %s : %s", yamlFile, testToRun.ExpectOutcome)))
				}
			}
		} else if testToRun.ExpectOutcome == "" {
			t.Error(utils.LogErrorMessage(fmt.Sprintf("  FAIL : %s : devfile was valid - No expected ouctome was set.", yamlFile)))
		} else if testToRun.ExpectOutcome != "PASS" {
			t.Error(utils.LogErrorMessage(fmt.Sprintf("  FAIL : %s : devfile was valid - Expected Error not found :  %s", yamlFile, testToRun.ExpectOutcome)))
		} else {
			passTests++
		}

	}

	failedTests := totalTests - passTests - skippedTests

	if failedTests > 0 {
		t.Errorf(utils.LogMessage(fmt.Sprintf("OVERALL FAIL :  %d tests passed. %d test skipped. %d tests failed.", passTests, skippedTests, failedTests)))
	} else {
		t.Log(utils.LogMessage(fmt.Sprintf("OVERALL PASS : %d tests passed. %d test skipped.", totalTests, skippedTests)))
	}
}
