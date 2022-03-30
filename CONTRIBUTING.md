# Contributing to Devfile Library

This document is a new contributor guide, which outlines the requirements for contributing to this repository.

To get an overview of the project, read the [README](README.md). For more information on devfiles, check the official [devfile docs](https://devfile.io/docs/devfile/2.1.0/user-guide/).

## Prerequisites

The following are required to work on devfile library:

- Git
- Go 1.15 or later

## Code of Conduct
Before contributing to this repository, see [contributor code of conduct](https://github.com/devfile/api/blob/main/CODE_OF_CONDUCT.md#contributor-covenant-code-of-conduct)

## How to Contribute

### Issues

If you spot a problem with devfile library, [search if an issue already exists](https://github.com/devfile/api/issues). If a related issue doesn't exist, you can open a new issue using a relevant [issue form](https://github.com/devfile/api/issues/new/choose).

### Writing Code

For writing the code just follow [Go guide](https://go.dev/doc/effective_go), and also test with [tesing](https://pkg.go.dev/testing). Remember to add new unit tests if new features have been introducted, or changes have been made to existing code. If there is something unclear of the style, just look at existing code which might help you to understand it better.

### Testing Changes
To run unit tests and api tests. Visit [library tests](tests/README.md) to find out more information on tests
```
make test
```

### Submitting Pull Request

When you think the code is ready for review, create a pull request and link the issue associated with it. 
Owners of the repository will watch out for and review new PR‘s. 
By default for each change in the PR, Travis CI runs all the tests against it. If tests are failing make sure to address the failures. 
If comments have been given in a review, they have to get integrated. 
After addressing review comments, don’t forget to add a comment in the PR afterward, so everyone gets notified by Github.


## Managing the Repository

### Updating Devfile Schema Files in Library

Executing `./scripts/updateApi.sh` fetches the latest `github.com/devfile/api` go mod and updates the schema saved under `pkg/devfile/parser/data`

The script also accepts a version number as an argument to update the devfile schema for a specific devfile version.
For example, running the following command will update the devfile schema for 2.0.0
```
./updateApi.sh 2.0.0
```
Running the script with no arguments will default to update the latest devfile version.


### Releases

Currently devfile library does not have schedule for new releases. A new version is being generated and released on demand. 
A new branch is expected to be created for a new release.
To generate a changelog for a new release, execute `./scripts/changelog-script.sh v1.x.y` for all the changes since the release v1.x.y
