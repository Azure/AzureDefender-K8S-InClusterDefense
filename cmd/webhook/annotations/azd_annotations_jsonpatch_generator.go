package annotations

import (
	"encoding/json"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/pkg/errors"
	"gomodules.xyz/jsonpatch/v2"
)

const (
	_addPatchOperation                             = "add"
	_annotationPatchPathPrefix                     = "/metadata/annotations"
	_containersVulnerabilityScanAnnotationFullPath = _annotationPatchPathPrefix + "/" + contracts.ContainersVulnerabilityScanInfoAnnotationName
)

func CreateContainersVulnerabilityScanAnnotationPatch(scanInfoList contracts.ContainerVulnerabilityScanInfoList) (*jsonpatch.JsonPatchOperation, error) {
	serVulnerabilitySecInfo, err := serializeAnnotationInnerObject(scanInfoList)
	if err != nil {
		return nil, errors.Wrap(err, "AzdAnnotationsPatchGenerator failed marshaling scanInfoList during CreateContainersVulnerabilityScanAnnotationPatch")
	}

	patch := jsonpatch.NewOperation(_addPatchOperation, _containersVulnerabilityScanAnnotationFullPath, serVulnerabilitySecInfo)
	return &patch, nil
}

func CreateInitAnnotations() (*jsonpatch.JsonPatchOperation, error) {
	patch := jsonpatch.NewOperation(_addPatchOperation, _annotationPatchPathPrefix, make(map[string]string))
	return &patch, nil
}

func serializeAnnotationInnerObject(object interface{}) (*string, error) {
	marshaledVulnerabilitySecInfo, err := json.Marshal(object)
	if err != nil {
		return nil, err
	}

	ser := string(marshaledVulnerabilitySecInfo)
	return &ser, nil
}
