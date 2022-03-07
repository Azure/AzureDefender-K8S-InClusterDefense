package utils

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"strings"
)

const (
	_errMsgInvalidPath = "admisionrequest.extractor: failed to access the given path"
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
func GoToDestNode(yamlFile *yaml.RNode, path ...string) (destNode *yaml.RNode, err error) {
	DestNode, err := yamlFile.Pipe(yaml.Lookup(path...))
	if err != nil {
		return nil, errors.Wrap(err, _errMsgInvalidPath)
	}
	return DestNode, err
}

// GetValue returns a string value that the given path contains, can be empty.
func GetValue(yamlFile *yaml.RNode, path ...string) (value string, err error) {
	DestNode, pathErr := GoToDestNode(yamlFile, path...)
	if err != nil {
		return "", pathErr
	}
	val := yaml.GetValue(DestNode)
	return val, nil
}

