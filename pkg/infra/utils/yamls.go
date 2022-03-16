package utils

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	yaml1 "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"
	"strings"
)

var (
	_errInvalidPath =errors.New("admisionrequest.extractor: failed to access the given path")
)

/*
CheckIfTwoYamlsHaveTheSameKeys Checks if 2 yaml's files have the same keys.
returns ...
*/
func CheckIfTwoYamlsHaveTheSameKeys(path1 string, path2 string) (bool, error) {
	ok, err := CheckIfAllKeysOfFirstAreInSecond(path1, path2)
	if err != nil {
		return false, err
	} else if ok == false {
		return false, nil
	}
	ok2, err := CheckIfAllKeysOfFirstAreInSecond(path2, path1)
	if err != nil {
		return false, err
	} else if ok2 == false {
		return false, nil
	}
	return true, nil
}

// CheckIfAllKeysOfFirstAreInSecond checks if all keys of the first yaml file are exists in second yaml
func CheckIfAllKeysOfFirstAreInSecond(path1 string, path2 string) (bool, error) {
	values1, err := CreateMapFromPathOfYaml(path1)
	if err != nil {
		return false, err
	}
	values2, err := CreateMapFromPathOfYaml(path2)
	if err != nil {
		return false, err
	}
	return AreAllKeysOfFirstMapExistsInSecondMap(values1, values2), nil
}

// CreateMapFromPathOfYaml Loads map from path of yaml file.
func CreateMapFromPathOfYaml(path string) (map[string]interface{}, error) {
	// check that file extension is yaml
	if !strings.HasSuffix(path, ".yaml") {
		errMsg := fmt.Sprintf("Suffix of %s is not yaml.", path)
		return nil, errors.New(errMsg)
	}

	// Read yaml file.
	values, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Convert yaml file into map.
	valuesMap := make(map[string]interface{})
	err = yaml.Unmarshal(values, &valuesMap)
	if err != nil {
		return nil, err
	}
	return valuesMap, err
}

// GoToDestNode returns the *Rnode of the given path.
func GoToDestNode(yamlFile *yaml1.RNode, path ...string) (destNode *yaml1.RNode, err error) {
	// Return filters of the given path strings.
	pathFilters := yaml1.Lookup(path...)
	// gets the rNode of the given pathFilters.
	DestNode, err := yamlFile.Pipe(pathFilters)
	if err != nil {
		return nil,  err
	}
	if DestNode == nil{
		return nil, _errInvalidPath
	}
	return DestNode, nil
}