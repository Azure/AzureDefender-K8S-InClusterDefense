package admisionrequest

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// stringInSlice return true if list contain str,false otherwise.
func stringInSlice(str string, list []string) bool {
	for _, listValue := range list {
		if listValue == str {
			return true
		}
	}
	return false
}

// goToDestNode returns the *Rnode of the given path.
func goToDestNode(yamlFile *yaml.RNode, path ...string) (destNode *yaml.RNode, err error) {
	DestNode, err := yamlFile.Pipe(yaml.Lookup(path...))
	if err != nil {
		return nil, errors.Wrap(err, _errMsgInvalidPath)
	}
	return DestNode, err
}