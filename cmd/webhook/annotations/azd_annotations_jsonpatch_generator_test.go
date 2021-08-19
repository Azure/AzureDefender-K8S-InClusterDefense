package annotations

import (
	"encoding/json"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/stretchr/testify/suite"
	"reflect"
	"testing"
	"time"
)

const (
	_expectedTestAddPatchOperation   = "add"
	_expectedTestAnnotationPatchPath = "/metadata/annotations"
)

type TestSuite struct {
	suite.Suite
	containersScanInfo []*contracts.ContainerVulnerabilityScanInfo
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
}

func (suite *TestSuite) Test_CreateContainersVulnerabilityScanAnnotationPatchAdd_TwoContainersScanInfo_AnnotationsGeneratedAsExpected() {
	result, err := CreateContainersVulnerabilityScanAnnotationPatchAdd(suite.containersScanInfo)
	suite.Nil(err)
	suite.Equal(_expectedTestAddPatchOperation, result.Operation)
	suite.Equal(_expectedTestAnnotationPatchPath, result.Path)

	// Get data string
	mapAnnotations, ok := result.Value.(map[string]string)
	suite.True(ok)

	suite.Equal(1, len(mapAnnotations))
	strValue, ok := mapAnnotations[contracts.ContainersVulnerabilityScanInfoAnnotationName]
	suite.True(ok)

	// Unmarshal
	scanInfoList := new(contracts.ContainerVulnerabilityScanInfoList)
	err = json.Unmarshal([]byte(strValue), scanInfoList)
	suite.Nil(err)

	// Verify timestamp
	diff := time.Now().UTC().Sub(scanInfoList.GeneratedTimestamp)
	suite.True((diff >= 0 && diff < time.Second))
	suite.Equal(time.UTC, scanInfoList.GeneratedTimestamp.Location())

	// Verify containers slice deep equal verification
	suite.True(reflect.DeepEqual(suite.containersScanInfo, scanInfoList.Containers))
}

func TestCreateContainersVulnerabilityScanAnnotationPatchAdd(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
