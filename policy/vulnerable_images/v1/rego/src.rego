package k8sazuredefendervulnerableimages

# This violation checks if there is a container with unscanned scanStatus.
violation[{"msg": msg, "details": details}] {
    # Extract containers
    containers := getContainers(input.review)
    container := containers[_]

    # Check if the scan status of the container is unscanned.
    container["scanStatus"] == "unscanned"

    # Construct violation msg:
    msg := sprintf("Unscanned image found in the container %v", [container.name])
    details := container
}

# This violation checks if there is some container that it's sum of the severities of the scanFindings are exceed some thresholds.
violation[{"msg": msg, "details": details}] {
    # Extract containers
    containers := getContainers(input.review)
    container := containers[_]
    # Explicit filter all containers that don't have unhealthy scan status.
    container["scanStatus"] == "unhealthyScan"
    # Filter scanfindings
	scanFindings := filterScanFindings(container["scanFindings"])
    isSevirityAboveThreshold(scanFindings)

    # Construct violation msg
    msg := sprintf("Found vulnerable container: %v", [container.name])
    details := {"Container": container, "ScanFindings": scanFindings}
}

# Extract the containers from the review object.
getContainers(review) = containers{
    # Extract ContainerVulnerabilityScanInfoList
    containerVulnerabilityScanInfoList := getContainerVulnerabilityScanInfoList(review)
    # Verify that the uid request that appears in containerVulnerabilityScanInfoList is match to the uid request of the request.
    isTheUIDRequestMatch(containerVulnerabilityScanInfoList)
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

# Verify that the uid request that appears in containerVulnerabilityScanInfoList is match to the uid request of the request.
isTheUIDRequestMatch(containerVulnerabilityScanInfoList){
    # Extract the uid requst from containerVulnerabilityScanInfoList
    uidRequestInContainerVulnerabilityScanInfoList := containerVulnerabilityScanInfoList["uidRequest"]
    # Extract the uid requst from the admission review
    uidRequestOfAdmissionReview := input.review["uid"]

    uidRequestInContainerVulnerabilityScanInfoList == uidRequestOfAdmissionReview
}

# Filter containers.
filterContainers(containers) = containers{
	containers = filterContainersWithHealthyScanStatus(containers)
    containers = filterContaintersWithExcludedImages(containers)
}

# Filter containers that are have healthy scanStatus.
filterContainersWithHealthyScanStatus(containers) = containers{
	containers = [containerVulnerabilityScanInfo | 	containerVulnerabilityScanInfo := containers[_]
    												containerVulnerabilityScanInfo["scanStatus"] != "healthyScan"]
}

# Filter containers that are appear in the excluded_images_pattern parameter.
filterContaintersWithExcludedImages(containers) = containers{
	containers = [containerVulnerabilityScanInfo | 	containerVulnerabilityScanInfo := containers[_]
    												not isImageMatchExcludedImagesPattern(containerVulnerabilityScanInfo["image"]["name"])]
}

# Checks if the registry appers in the exclduded_registreis pattern
isImageMatchExcludedImagesPattern(image_name){
    image_pattern := input.parameters["excluded_images_pattern"][_]
    re_match(image_pattern, image_name)
}

# Filter ScanFindings
filterScanFindings(scanFindings) = scanFindings{
	scanFindings = filterScanFindingsExcludedFindings(scanFindings)
	scanFindings = filterScanFindingsIfExcludeNotPatchableFindingsParamTrue(scanFindings)
}

# Filter all scanfindings that appear in the excludeFindingIDsList.
filterScanFindingsExcludedFindings(scanFindings) = scanFindings{
	scanFindings = [scanFinding | 	scanFinding := scanFindings[_]
    								not isScanFindingAppearsInExlcudeFindingIDsList(scanFinding)]
}

# Checks if the scanFinding appers in the list of the excluded findings id:
isScanFindingAppearsInExlcudeFindingIDsList(scanFinding){
    scanFindingID := scanFinding["id"]
    excludedScanFinding := input.parameters.exlcudeFindingIDs[_]
    scanFindingID == excludedScanFinding
}

# Filter scanFindings that are not patchable if the ExcludeNotPatchableFindings param is set to true
filterScanFindingsIfExcludeNotPatchableFindingsParamTrue(scanFindings) = scanFindings{
	input.parameters.excludeNotPatchableFindings
	scanFindings = [scanFinding | 	scanFinding := scanFindings[_]
    								scanFinding["patchable"]]
}
# Keep all scanFindings (patchable and not patchable) if the ExcludeNotPatchableFindings param is set to false
filterScanFindingsIfExcludeNotPatchableFindingsParamTrue(scanFindings) = scanFindings{
	not input.parameters.excludeNotPatchableFindings
}

# Checks if the total of High sevirity is above the threshold
isSevirityAboveThreshold(scanFindings){
    isSevirityTypeAboveThreshold(scanFindings, "High")
}
# Checks if the total of  Medium sevirity is above the threshold
isSevirityAboveThreshold(scanFindings){
    isSevirityTypeAboveThreshold(scanFindings, "Medium")
}
# Checks if the total of  Low sevirity is above the threshold
isSevirityAboveThreshold(scanFindings){
    isSevirityTypeAboveThreshold(scanFindings, "Low")
}

# Check if the total of all findings with sevirity level of severtyType (patchable and not patchable) is exceeding the threshold
isSevirityTypeAboveThreshold(scanFindings, sevirityType){
    c := count([scanFinding | 	scanFinding := scanFindings[_]
                                scanFinding["severity"] == sevirityType])

    c > input.parameters.sevirity[sevirityType]
}