package k8sazuredefendervulnerableimages
# Is Unscanned violation
violation[{"msg": msg, "details": details}] {
    # Extract ContainerVulnerabilityScanInfoList
    containerVulnerabilityScanInfoList := getContainerVulnerabilityScanInfoList(input.review)
    # Check that the annotations are exist and updated.
    are_the_annotations_updated(containerVulnerabilityScanInfoList)
    # iterate on the containers of containerVulnerabilityScanInfoList.
    ContainerVulnerabilityScanInfo := containerVulnerabilityScanInfoList["containers"][_]
    is_unscanned(ContainerVulnerabilityScanInfo)
    not is_image_match_excluded_images_pattern(ContainerVulnerabilityScanInfo.image.name)
    # In case that there is a viloation, create informative msg:
    msg := sprintf("The image %v is unscanned", [ContainerVulnerabilityScanInfo["image"].name])
    details := ContainerVulnerabilityScanInfo
}

# Servirity violation
violation[{"msg": msg, "details": details}] {
    # Extract ContainerVulnerabilityScanInfoList
    containerVulnerabilityScanInfoList := getContainerVulnerabilityScanInfoList(input.review)
    # Check that the annotations are exist and updated.
    are_the_annotations_updated(containerVulnerabilityScanInfoList)
    # iterate on the containers of containerVulnerabilityScanInfoList.
    ContainerVulnerabilityScanInfo := containerVulnerabilityScanInfoList["containers"][_]
    # Check that scanStatus is unhealthy scan.
    ContainerVulnerabilityScanInfo["scanStatus"] == "unhealthyScan"
    # Check if the container is exceed the threshold.
    is_sevirity_above_threshold(ContainerVulnerabilityScanInfo)
    msg := sprintf("The image %v is unscanned", [ContainerVulnerabilityScanInfo])
    details := ContainerVulnerabilityScanInfo
}

# Checks if the total of High sevirity is above the threshold
is_sevirity_above_threshold(ContainerVulnerabilityScanInfo){
    is_sevirity_above_threshold_in_context_of_patchable_findings(ContainerVulnerabilityScanInfo, "High")
}
# ********************* OR *************************************
# Checks if the total of  Medium sevirity is above the threshold
is_sevirity_above_threshold(ContainerVulnerabilityScanInfo){
    is_sevirity_above_threshold_in_context_of_patchable_findings(ContainerVulnerabilityScanInfo, "Medium")
}
# ********************* OR *************************************
# Checks if the total of  Low sevirity is above the threshold
is_sevirity_above_threshold(ContainerVulnerabilityScanInfo){
    is_sevirity_above_threshold_in_context_of_patchable_findings(ContainerVulnerabilityScanInfo, "Low")
}


# In case that input.parameters.excludeNotPatchableFindings is True, summarize only patchable findings
is_sevirity_above_threshold_in_context_of_patchable_findings(ContainerVulnerabilityScanInfo, sevirityType){
    input.parameters.excludeNotPatchableFindings
    is_sevirity_type_above_threshold_while_excluding_not_patchable_findings(ContainerVulnerabilityScanInfo, sevirityType)
}
# ********************* OR *************************************
# In case that input.parameters.excludeNotPatchableFindings is False, summarize all (patchable and not patchable) findings
is_sevirity_above_threshold_in_context_of_patchable_findings(ContainerVulnerabilityScanInfo, sevirityType){
    not input.parameters.excludeNotPatchableFindings
    is_sevirity_type_above_threshold_while_including_not_patchable_findings(ContainerVulnerabilityScanInfo, sevirityType)
}


# Check if the total of all findings with sevirity level of severtyType and are patchable is exceeding the threshold
is_sevirity_type_above_threshold_while_excluding_not_patchable_findings(ContainerVulnerabilityScanInfo, sevirityType){
    c := count([scanFinding | 	scanFinding := ContainerVulnerabilityScanInfo["scanFindings"][_]
                                scanFinding["patchable"]
                                scanFinding["severity"] == sevirityType
                                not is_scanFinding_appears_in_exlcudeFindingIDsList(scanFinding)])

    c > input.parameters.sevirity[sevirityType]
}

# Check if the total of all findings with sevirity level of severtyType (patchable and not patchable) is exceeding the threshold
is_sevirity_type_above_threshold_while_including_not_patchable_findings(ContainerVulnerabilityScanInfo, sevirityType){
    c := count([scanFinding | 	scanFinding := ContainerVulnerabilityScanInfo["scanFindings"][_]
                                scanFinding["severity"] == sevirityType
                                not is_scanFinding_appears_in_exlcudeFindingIDsList(scanFinding)])
    c > input.parameters.sevirity[sevirityType]
}


# Checks if the scanFinding appers in the list of the excluded findings id:
is_scanFinding_appears_in_exlcudeFindingIDsList(scanFinding){
    scanFindingID := scanFinding["id"]
    excludedScanFinding := input.parameters.exlcudeFindingIDs[_]
    scanFindingID == excludedScanFinding
}

# Checks if the registry appers in the exclduded_registreis pattern
is_image_match_excluded_images_pattern(image_name){
    image_pattern := input.parameters["excluded_images_pattern"][_]
    re_match(image_pattern, image_name)
}

# Checks if scan info is unscanned
is_unscanned(scan_info){
    scan_info["scanStatus"] == "unscanned"
}

# Gets review object and returns unnmarshelled scan resulsts (i.e. as array of scan results)
getContainerVulnerabilityScanInfoList(review) = containerVulnerabilityScanInfoList{
    scanResults := review.object.metadata.annotations["azuredefender.io/containers.vulnerability.scan.info"]
    containerVulnerabilityScanInfoList := json.unmarshal(scanResults)
}

# Verify that the abs(enrichment time stemp - creation time stamp) <= dur 
are_the_annotations_updated(containerVulnerabilityScanInfoList){
    # Extract enrichment timestamp.
    timestamp := containerVulnerabilityScanInfoList["generatedTimestamp"]
    enrichmentTimestamp := time.parse_rfc3339_ns(timestamp)
    # Extract creation timestamp
    creationsTimestamp := time.parse_rfc3339_ns(input.review.object.metadata["creationTimestamp"])
    # Convert duration param to time object
    # TODO Should we define diff time than 20 seconds?
    dur := time.parse_duration_ns("20s")
    abs(enrichmentTimestamp - creationsTimestamp) <= dur
}