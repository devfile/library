# Devfile Parser Library

## About

The Devfile Parser library is a Golang module that:
1. parses the devfile.yaml as specified by the [api](https://devfile.github.io/devfile/api-reference.html) & [schema](https://github.com/devfile/api/tree/master/schemas/latest).
2. writes to the devfile.yaml with the updated data.
3. generates Kubernetes objects for the various devfile resources.
4. defines util functions for the devfile.


## Usage

In the future, the following projects will be consuming this library as a Golang dependency

* [Workspace Operator](https://github.com/devfile/devworkspace-operator)
* [odo](https://github.com/openshift/odo)
* [OpenShift Console](https://github.com/openshift/console)

## Issues

Issues are tracked in the [devfile/api](https://github.com/devfile/api) repo with the label [area/library](https://github.com/devfile/api/issues?q=is%3Aopen+is%3Aissue+label%3Aarea%2Flibrary) 