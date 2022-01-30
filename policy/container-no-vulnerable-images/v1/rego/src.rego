package k8sazuredefendervulnerableimages

# This violation checks if there is a container with unscanned scanStatus.
violation[{"msg":msg}] {
    # Extract containers
    containers := getApplicableContainersScanInfo(input.review)
    container := containers[_]
    # Check if the scan status of the container is unscanned.
    container["scanStatus"] == "unscanned"
    # Construct violation msg:
    additionalData := getAdditionalData(container)
    msg := sprintf("Image <%v> under container <%v> must be scanned for vulnerabilities by Defender for Cloud before deployment, additionalData: <%v>", [container.image.name, container.name, additionalData])
}
# This violation checks if there is some container that it's sum of the severities of the scanFindings are exceed some thresholds.
violation[{"msg":msg}] {
    # Extract containers
    containers := getApplicableContainersScanInfo(input.review)
    container := containers[_]
    # Explicit filter all containers that don't have unhealthy scan status.
    container["scanStatus"] == "unhealthyScan"
    # Filter scanfindings
    scanFindings := filterScanFindings(container["scanFindings"])
    isSeverityAboveThreshold(scanFindings)
    # Construct violation msg
    additionalData := getAdditionalData(container)
    msg := sprintf("Image <%v> under container <%v> with digest <%v>, has vulnerabilities that must be remediated before deployment. Refer to Defender for Cloud to view vulnerability data for your image.", [container.image.name, container.name, container.image.digest])
}
# Extract the containers from the review object.
getApplicableContainersScanInfo(review) = containers{
  # Extract ContainerVulnerabilityScanInfoList
  containerVulnerabilityScanInfoList := getContainerVulnerabilityScanInfoList(review)
  # Filter containers from containerVulnerabilityScanInfoList
  containers := filterContainers(containerVulnerabilityScanInfoList["containers"])
}
# Gets review object and returns unnmarshelled scan resulsts (i.e. as array of scan results).
# See https://github.com/Azure/AzureDefender-K8S-InClusterDefense/blob/master/pkg/azdsecinfo/contracts/containers_vulnerability_scan_info.go
# for more information about the contract and the unmarshalled object.
getContainerVulnerabilityScanInfoList(review) = containerVulnerabilityScanInfoList{
  scanResults := review.object.metadata.annotations["azuredefender.io/containers.vulnerability.scan.info"]
  containerVulnerabilityScanInfoList := json.unmarshal(scanResults)
}
# Filter containers.
filterContainers(containers) = containers{
  containers = filterContainersWithHealthyScanStatus(containers)
  containers = filterContaintersWithExcludedImages(containers)
}
# Filter containers that are have healthy scanStatus.
filterContainersWithHealthyScanStatus(containers) = out{
  out = [containerVulnerabilityScanInfo | 	containerVulnerabilityScanInfo := containers[_]
  containerVulnerabilityScanInfo["scanStatus"] != "healthyScan"]
}
# Filter containers that are appear in the excludedImages parameter.
filterContaintersWithExcludedImages(containers) = out{
  out = [containerVulnerabilityScanInfo | 	containerVulnerabilityScanInfo := containers[_]
  not isImageMatchExcludedImagesPattern(containerVulnerabilityScanInfo["image"]["name"])]
}
# Checks if the registry appears in the excludedImages pattern
isImageMatchExcludedImagesPattern(image_name){
  image_pattern := input.parameters["excludedImages"][_]
  re_match(image_pattern, image_name)
}
# Filter ScanFindings
filterScanFindings(scanFindings) = out{
  filtered := filterScanFindingsExcludedFindings(scanFindings)
  out = filterScanFindingsNotPatchableBelowThreshold(filtered)
}
# Filter all scanfindings that appear in the excludeFindingIDsList.
filterScanFindingsExcludedFindings(scanFindings) = out{
  out = [scanFinding | 	scanFinding := scanFindings[_]
  not isScanFindingAppearsInexcludeFindingIDsList(scanFinding)]
}
# Checks if the scanFinding appers in the list of the excluded findings id:
isScanFindingAppearsInexcludeFindingIDsList(scanFinding){
  scanFindingID := scanFinding["id"]
  excludedScanFinding := input.parameters.excludeFindingIDs[_]
  scanFindingID == excludedScanFinding
}
# Filter all scanfindings that are not patchable and their severity is below severityThresholdForExcludingNotPatchableFindings.
filterScanFindingsNotPatchableBelowThreshold(scanFindings) = out{
  out = [scanFinding | scanFinding := scanFindings[_] ; isScanFindingPatchableOrAboveThresholdSeverity(scanFinding)]
}
# Check if scanFinding is patchable
isScanFindingPatchableOrAboveThresholdSeverity(scanFinding){
  scanFinding["patchable"]
}
# Check if scanFinding is not patchable and the severity is above the threshold (severityThresholdForExcludingNotPatchableFindings)
isScanFindingPatchableOrAboveThresholdSeverity(scanFinding){
  not scanFinding["patchable"]
  # Create map between severity to the integer level. None = 0, Low = 1, Medium = 2, High = 3
  severityToLevel := {"None":0, "Low":1, "Medium":2, "High": 3}
  # Check that the level of the scanFinding is above the threshold level.
  severityToLevel[scanFinding["severity"]] > severityToLevel[input.parameters.severityThresholdForExcludingNotPatchableFindings]
}
# Checks if the total of High severity is above the threshold
isSeverityAboveThreshold(scanFindings){
  isSeverityTypeAboveThreshold(scanFindings, "High")
}
# Checks if the total of  Medium severity is above the threshold
isSeverityAboveThreshold(scanFindings){
  isSeverityTypeAboveThreshold(scanFindings, "Medium")
}
# Checks if the total of  Low severity is above the threshold
isSeverityAboveThreshold(scanFindings){
  isSeverityTypeAboveThreshold(scanFindings, "Low")
}
# Check if the total of all findings with severity level of severtyType (patchable and not patchable) is exceeding the threshold
isSeverityTypeAboveThreshold(scanFindings, severityType){
  # Extract all scanFinding that have serverity level equal to severity type.
  c := count([scanFinding | scanFinding := scanFindings[_]
  scanFinding["severity"] == severityType])
  c > input.parameters.severity[severityType]
}
getAdditionalData(container) = additionalData{
 not container.additionalData
 additionalData = "None" # Default additionalData
}
getAdditionalData(container) = additionalData{
 additionalData = container.additionalData	# Extract additionalData from the object
}