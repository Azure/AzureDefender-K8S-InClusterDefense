package queries

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/stretchr/testify/assert"
	"testing"
)

// _expectedContainerVulnerabilityScanResultsQuery is the expected query to be genrated
const _expectedContainerVulnerabilityScanResultsQuery = `
securityresources
 | where type == 'microsoft.security/assessments/subassessments'
 // The 2 lines below describe why we used in the third line thw two numbers 130 and 78.
 // 130 = strlen("providers/Microsoft.Security/assessments/dbd0cb49-b563-45e7-9724-889e799fa648/subassessments/b894c178-8c91-448d-9a77-9de8bb4508dc");
 // 78  = strlen("providers/Microsoft.Security/assessments/dbd0cb49-b563-45e7-9724-889e799fa648");
 | where indexof(id,'/providers/Microsoft.Security/assessments/dbd0cb49-b563-45e7-9724-889e799fa648', -130, 78) != -1 
 | extend digest = tostring(properties.additionalData.imageDigest)
 | extend repository = tostring(properties.additionalData.repositoryName)
 | extend registry = tostring(properties.additionalData.registryHost)
 | where   registry =~ "tomer.azurecr.io" and repository =~ "test-image" and digest == "sha256:763bdd5314d126766d54cec7585f361c8c1429a2c51c818f0e7d0cab21a1481e"
 | extend scanFindingSeverity = tostring(properties.status.severity), scanStatus = tostring(properties.status.code)
 | extend findingsIds = tostring(properties.id), patchable = tostring(properties.additionalData.patchable)
 | project id, registry, repository, digest, scanStatus, scanFindingSeverity, findingsIds, patchable
`

// Tests query temaplte it self and it's generation
func Test_QueryGenerator_GenerateImageVulnerabilityScanQuery(t *testing.T) {
	generator, err := CreateARGQueryGenerator(instrumentation.NewNoOpInstrumentationProvider())
	parameters := &ContainerVulnerabilityScanResultsQueryParameters{
		Registry:   "tomer.azurecr.io",
		Repository: "test-image",
		Digest:     "sha256:763bdd5314d126766d54cec7585f361c8c1429a2c51c818f0e7d0cab21a1481e",
	}
	query, err := generator.GenerateImageVulnerabilityScanQuery(parameters)
	assert.Nil(t, err)
	assert.Equal(t, _expectedContainerVulnerabilityScanResultsQuery, query)
}
