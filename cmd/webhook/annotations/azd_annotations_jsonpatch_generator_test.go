package annotations

import (
	"encoding/json"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook/admisionrequest"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

const (
	_expectedTestAddPatchOperation   = "add"
	_expectedTestAnnotationPatchPath = "/metadata/annotations"
	_annotationTestKeyOne            = "cluster-autoscaler.kubernetes.io/safe-to-evict"
	_annotationTestValueOne          = "true"
	_annotationTestKeyTwo            = "container.seccomp.security.alpha.kubernetes.io/manager"
	_annotationTestValueTwo          = "runtime/default"
)

type TestSuite struct {
	suite.Suite
	containersScanInfo              []*contracts.ContainerVulnerabilityScanInfo
	workloadResourceNoAnnotations   *admisionrequest.WorkloadResource
	workloadResourceWithAnnotations *admisionrequest.WorkloadResource
}

func (suite *TestSuite) SetupSuite() {
	suite.containersScanInfo = []*contracts.ContainerVulnerabilityScanInfo{
		{
			Name: "container1",
			Image: &contracts.Image{
				Name:   "imageTest1",
				Digest: "imageDigest1",
			},
			ScanStatus: contracts.UnhealthyScan,
			ScanFindings: []*contracts.ScanFinding{
				{
					Id:        "11",
					Severity:  "High",
					Patchable: true,
				},
			},
		},
		{
			Name: "container2",
			Image: &contracts.Image{
				Name:   "imageTest2",
				Digest: "imageDigest2",
			},
			ScanStatus: contracts.UnhealthyScan,
			ScanFindings: []*contracts.ScanFinding{
				{
					Id:        "22",
					Severity:  "Medium",
					Patchable: true,
				},
			},
		},
	}
	suite.workloadResourceNoAnnotations = &admisionrequest.WorkloadResource{Spec: &admisionrequest.PodSpec{}, Metadata: &admisionrequest.ObjectMetadata{}}
	suite.workloadResourceNoAnnotations = &admisionrequest.WorkloadResource{Spec: &admisionrequest.PodSpec{}, Metadata: &admisionrequest.ObjectMetadata{}}
	suite.workloadResourceWithAnnotations = createWorkloadResourceWithAnnotationsForTest()
}

func (suite *TestSuite) Test_CreateContainersVulnerabilityScanAnnotationPatchAdd_TwoContainersScanInfo_AnnotationsGeneratedAsExpected() {
	suite.checkContainersVulnerabilityScanAnnotation(1, suite.workloadResourceNoAnnotations)
}

func (suite *TestSuite) Test_CreateContainersVulnerabilityScanAnnotationPatchAdd_PodWithAnnotations_AnnotationsGeneratedAsExpected() {

	// check containers vulnerability scan annotations
	mapAnnotations := suite.checkContainersVulnerabilityScanAnnotation(3, suite.workloadResourceWithAnnotations)

	// check no override of existing annotations
	suite.checkNoOverrideOfExistingAnnotations(mapAnnotations, _annotationTestKeyOne, _annotationTestValueOne)
	suite.checkNoOverrideOfExistingAnnotations(mapAnnotations, _annotationTestKeyTwo, _annotationTestValueTwo)
}

func (suite *TestSuite) Test_DeleteContainersVulnerabilityScanAnnotationPatch_PodWithAzdAnnotations_AnnotationsGeneratedAsExpected() {
	result, err := CreateAnnotationPatchToDeleteContainersVulnerabilityScanAnnotationIfNeeded(createWorkloadResourceWithAzdAnnotationsForTest())
	suite.Nil(err)
	mapAnnotations, ok := result.Value.(map[string]string)
	suite.True(ok)
	suite.Equal(2, len(mapAnnotations))

	_, ok = mapAnnotations[contracts.ContainersVulnerabilityScanInfoAnnotationName]
	suite.False(ok)
	// check no override of existing annotations
	suite.checkNoOverrideOfExistingAnnotations(mapAnnotations, _annotationTestKeyOne, _annotationTestValueOne)
	suite.checkNoOverrideOfExistingAnnotations(mapAnnotations, _annotationTestKeyTwo, _annotationTestValueTwo)
}

func (suite *TestSuite) Test_DeleteContainersVulnerabilityScanAnnotationPatch_PodWithoutAnnotations_AnnotationsGeneratedAsExpected() {

	result, err := CreateAnnotationPatchToDeleteContainersVulnerabilityScanAnnotationIfNeeded(createWorkloadResourceWithoutAnnotationsForTest())
	suite.Nil(err)
	suite.Nil(result)
}

func (suite *TestSuite) Test_DeleteContainersVulnerabilityScanAnnotationPatch_PodWithAnnotations_AnnotationsGeneratedAsExpected() {

	result, err := CreateAnnotationPatchToDeleteContainersVulnerabilityScanAnnotationIfNeeded(createWorkloadResourceWithAnnotationsForTest())
	suite.Nil(err)
	suite.Nil(result)
}

func (suite *TestSuite) checkContainersVulnerabilityScanAnnotation(patchLen int, pod *admisionrequest.WorkloadResource) map[string]string {
	result, err := CreateContainersVulnerabilityScanAnnotationPatchAdd(suite.containersScanInfo, pod)
	suite.Nil(err)
	suite.Equal(_expectedTestAddPatchOperation, result.Operation)
	suite.Equal(_expectedTestAnnotationPatchPath, result.Path)

	mapAnnotations, ok := result.Value.(map[string]string)
	suite.True(ok)

	suite.Equal(patchLen, len(mapAnnotations))
	strContainersVulnerabilityScanValue, ok := mapAnnotations[contracts.ContainersVulnerabilityScanInfoAnnotationName]
	suite.True(ok)

	// Unmarshal
	scanInfoList := new(contracts.ContainerVulnerabilityScanInfoList)
	err = json.Unmarshal([]byte(strContainersVulnerabilityScanValue), scanInfoList)
	suite.Nil(err)

	// Verify timestamp
	diff := time.Now().UTC().Sub(scanInfoList.GeneratedTimestamp)
	suite.True((diff >= 0 && diff < time.Second))
	suite.Equal(time.UTC, scanInfoList.GeneratedTimestamp.Location())
	return mapAnnotations
}

func (suite *TestSuite) checkNoOverrideOfExistingAnnotations(mapAnnotations map[string]string, expectedKey string, expectedVal string) {
	strAnnotationField1, ok := mapAnnotations[expectedKey]
	suite.True(ok)
	suite.Equal(expectedVal, strAnnotationField1)
}

func TestCreateContainersVulnerabilityScanAnnotationPatchAdd(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func createWorkloadResourceWithAnnotationsForTest() *admisionrequest.WorkloadResource {
	metadata := &admisionrequest.ObjectMetadata{
		Name: "podTest",
		Annotations: map[string]string{
			_annotationTestKeyOne: _annotationTestValueOne,
			_annotationTestKeyTwo: _annotationTestValueTwo,
		},
	}
	workloadResource := admisionrequest.WorkloadResource{Metadata: metadata}
	return &workloadResource
}

func createWorkloadResourceWithAzdAnnotationsForTest() *admisionrequest.WorkloadResource {
	metadata := &admisionrequest.ObjectMetadata{
		Name: "podTest",
		Annotations: map[string]string{
			_annotationTestKeyOne: _annotationTestValueOne,
			_annotationTestKeyTwo: _annotationTestValueTwo,
			contracts.ContainersVulnerabilityScanInfoAnnotationName: "some value",
		},
	}
	workloadResource := admisionrequest.WorkloadResource{Metadata: metadata}
	return &workloadResource
}

func createWorkloadResourceWithoutAnnotationsForTest() *admisionrequest.WorkloadResource {
	metadata := &admisionrequest.ObjectMetadata{
		Name: "podTest",
	}
	workloadResource := admisionrequest.WorkloadResource{Metadata: metadata}
	return &workloadResource
}
