package admisionrequest

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// StringInSlice return true if list contain str,false otherwise.
func StringInSlice(str string, list []string) bool {
	for _, listValue := range list {
		if listValue == str {
			return true
		}
	}
	return false
}

// goToDestNode returns the *Rnode of the given path.
func goToDestNode(yamlFile *yaml.RNode, path ...string) (destNode *yaml.RNode, err error) {
	// Return filters of the given path strings.
	pathFilters := yaml.Lookup(path...)
	// gets the rNode of the given pathFilters.
	DestNode, err := yamlFile.Pipe(pathFilters)
	if err != nil {
		return nil, errors.Wrap(err, _errMsgInvalidPath)
	}
	return DestNode, nil
}