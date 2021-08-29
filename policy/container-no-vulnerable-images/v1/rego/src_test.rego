package k8sazuredefendervulnerableimages

# Check that if the annotations are empty then there is no violation.
test_input_no_annotations {
    input := { "review": input_review_no_annotations, "parameters": input_parameters_empty}
    results := violation with input as input
    count(results) == 0
}

# Check that if the annotations are old  (i.e. the uid request that appears in the annotations is different from the uid of the review) then there is no violoation.
test_input_stale_annotations_zero_violations {
    input := { "review": input_review_violation_with_stale_annotations, "parameters": input_parameters_high_0_severityThresholdForExcludingNotPatchableFindings_None}
    results := violation with input as input
    count(results) == 0
}

# Check that if the annotations are old  (i.e. the uid request that appears in the annotations is different from the uid of the review) then there is no violoation.
test_input_not_stale_annotations_1_violation {
    input := { "review": input_review_violation_with_not_stale_annotations, "parameters": input_parameters_high_0_severityThresholdForExcludingNotPatchableFindings_None}
    results := violation with input as input
    count(results) == 1
}

# Checks that if there is unscanned image that is appers in the excluded images regex list, then there is no violation.
test_input_unscanned_image_that_appears_in_excluded_images_0_violation {
    input := { "review": input_review_creation_time_ok_scan_status_unscanned, "parameters": input_parameters_tomerazurecr_image_excluded_severityHighTreshold_2}
    results := violation with input as input
    count(results) == 0
}

# Checks that if there is unscanned image that is appers in the excluded images regex list, then there no violation.
test_input_unscanned_image_that_isnt_appears_in_excluded_images_1_violation {
    input := { "review": input_review_creation_time_ok_scan_status_unscanned, "parameters": input_parameters_liorazurecr_image_excluded_severityHighTreshold_2}
    results := violation with input as input
    count(results) == 1
}

# Checks that although there is image that exceeds the severity treshold,if it appears in the excluded images regex list, then there is no violation.
test_input_unhealthy_image_that_appears_in_excluded_images_0_violation {
    input := { "review": input_review_unhealthy_2_containers_with_diff_severities, "parameters": input_parameters_tomerazurecr_image_excluded_severityHighTreshold_2}
    results := violation with input as input
    count(results) == 0
}

# Checks that although there is image that exceeds the severity treshold,if it is not appears in the excluded images regex list, then there is violation.
test_input_unhealthy_image_that_isnt_appears_in_excluded_images_1_violation {
    input := { "review": input_review_unhealthy_2_containers_with_diff_severities, "parameters": input_parameters_liorazurecr_image_excluded_severityHighTreshold_2}
    results := violation with input as input
    count(results) == 1
}

# Checks that if there is unscanned image, then there is 1 violtation.
test_input_creation_time_ok_scan_status_unscanned {
    input := { "review": input_review_creation_time_ok_scan_status_unscanned, "parameters": input_parameters_empty}
    results := violation with input as input    
    count(results) == 1
}

# Checks that if there is one container that is complainet (high severity) and one that isn't then we get only 1 violation.
test_input_review_unhealthy_one_container_above_highSeverity_1_violotation {
    input := { "review": input_review_unhealthy_2_containers_with_diff_severities, "parameters": input_parameters_severityHighTreshold_2}
    results := violation with input as input
    count(results) == 1
}

# Checks that if there are 2 conatiners with higher Highseverity than the treshold then we get 2 viloations.
test_input_review_unhealthy_two_containers_above_highSeverity_2_violotation {
    input := { "review": input_review_unhealthy_2_containers_with_diff_severities, "parameters": input_parameters_severityHighTreshold_1}
    results := violation with input as input
    count(results) == 2
}

# Checks that if there is one container with higher MediumSeverity than the threshold, then we get violation
test_input_review_unhealthy_one_container_above_mediumSeverity_one_violotation {
    input := { "review": input_review_unhealthy_2_containers_with_diff_severities, "parameters": input_parameters_severityMediumTreshold_0}
    results := violation with input as input
    count(results) == 1
}

# Checks that if there is one container with higher LowSeverity than the threshold, then we get violation
test_input_review_unhealthy_one_container_above_lowSeverity_one_violotation {
    input := { "review": input_review_unhealthy_2_containers_with_diff_severities, "parameters": input_parameters_severityLowTreshold_0}
    results := violation with input as input
    count(results) == 1
}

# Checks that if there altough there is scanFinding with high seveirty and its exceeed the threshold, if the ID is exist in exlcudeFindingIDs, then we won't get violation.
test_input_review_unhealthy_cotainer_with_1_high_finding_that_is_appears_in_exlcudeFindingIDs_zero_violoations {
    input := { "review": input_review_unhealthy_container_with_not_patchable_finding, "parameters": input_parameters_high_0_excluded_finding_id_125}
    results := violation with input as input
    count(results) == 0
}

# Checks that if there altough there is scanFinding with high seveirty and its exceeed the threshold, if the ID isn't exist in exlcudeFindingIDs, then we get violation.
test_input_review_unhealthy_cotainer_with_1_high_finding_that_isnt_appears_in_exlcudeFindingIDs_one_violations {
    input := { "review": input_review_unhealthy_container_with_patchable_finding, "parameters": input_parameters_high_0_excluded_finding_id_126}
    results := violation with input as input
    count(results) == 1
}

# Checks if patchableSeverityThreshold set to None and there is Low scanFinding, then there is violation
test_input_review_unhealthy_container_1_low_not_patchable_finding_patchableSeverityThreshold_none_1_violoations {
    input := { "review": input_review_unhealthy_container_with_not_patchable_severities, "parameters": input_parameters_low_0_severityThresholdForExcludingNotPatchableFindings_None}
    results := violation with input as input
    count(results) == 1
}

# Checks if patchableSeverityThreshold set to None and there is Medium scanFinding, then there is violation
test_input_review_unhealthy_container_1_medium_not_patchable_finding_patchableSeverityThreshold_none_1_violoations {
    input := { "review": input_review_unhealthy_container_with_not_patchable_severities, "parameters": input_parameters_medium_0_severityThresholdForExcludingNotPatchableFindings_None}
    results := violation with input as input
    count(results) == 1
}

# Checks if patchableSeverityThreshold set to None and there is High scanFinding, then there is violation
test_input_review_unhealthy_container_1_high_not_patchable_finding_patchableSeverityThreshold_none_1_violoations {
    input := { "review": input_review_unhealthy_container_with_not_patchable_severities, "parameters": input_parameters_high_0_severityThresholdForExcludingNotPatchableFindings_None}
    results := violation with input as input
    count(results) == 1
}

# Checks if patchableSeverityThreshold set to Low and there is Low scanFinding, then there is no violation
test_input_review_unhealthy_container_1_low_not_patchable_finding_patchableSeverityThreshold_low_0_violoations {
    input := { "review": input_review_unhealthy_container_with_not_patchable_severities, "parameters": input_parameters_low_0_severityThresholdForExcludingNotPatchableFindings_Low}
    results := violation with input as input
    count(results) == 0
}

# Checks if patchableSeverityThreshold set to Low and there is Medium scanFinding, then there is violation
test_input_review_unhealthy_container_1_medium_not_patchable_finding_patchableSeverityThreshold_low_1_violoations {
    input := { "review": input_review_unhealthy_container_with_not_patchable_severities, "parameters": input_parameters_medium_0_severityThresholdForExcludingNotPatchableFindings_Low}
    results := violation with input as input
    count(results) == 1
}

# Checks if patchableSeverityThreshold set to Low and there is High scanFinding, then there is violation
test_input_review_unhealthy_container_1_high_not_patchable_finding_patchableSeverityThreshold_low_1_violoations {
    input := { "review": input_review_unhealthy_container_with_not_patchable_severities, "parameters": input_parameters_high_0_severityThresholdForExcludingNotPatchableFindings_Low}
    results := violation with input as input
    count(results) == 1
}

# Checks if patchableSeverityThreshold set to Medium and there is Low scanFinding, then there is no violation
test_input_review_unhealthy_container_1_low_not_patchable_finding_patchableSeverityThreshold_Medium_0_violoations {
    input := { "review": input_review_unhealthy_container_with_not_patchable_severities, "parameters": input_parameters_low_0_severityThresholdForExcludingNotPatchableFindings_Medium}
    results := violation with input as input
    count(results) == 0
}

# Checks if patchableSeverityThreshold set to Medium and there is Medium scanFinding, then there is no violation
test_input_review_unhealthy_container_1_medium_not_patchable_finding_patchableSeverityThreshold_Medium_0_violoations {
    input := { "review": input_review_unhealthy_container_with_not_patchable_severities, "parameters": input_parameters_medium_0_severityThresholdForExcludingNotPatchableFindings_Medium}
    results := violation with input as input
    count(results) == 0
}

# Checks if patchableSeverityThreshold set to Medium and there is High scanFinding, then there is violation
test_input_review_unhealthy_container_1_high_not_patchable_finding_patchableSeverityThreshold_Medium_1_violoations {
    input := { "review": input_review_unhealthy_container_with_not_patchable_severities, "parameters": input_parameters_high_0_severityThresholdForExcludingNotPatchableFindings_Medium}
    results := violation with input as input
    count(results) == 1
}

# Checks if patchableSeverityThreshold set to High and there is Low scanFinding, then there is no violation
test_input_review_unhealthy_container_1_low_not_patchable_finding_patchableSeverityThreshold_High_0_violoations {
    input := { "review": input_review_unhealthy_container_with_not_patchable_severities, "parameters": input_parameters_low_0_severityThresholdForExcludingNotPatchableFindings_High}
    results := violation with input as input
    count(results) == 0
}

# Checks if patchableSeverityThreshold set to High and there is Medium scanFinding, then there is no violation
test_input_review_unhealthy_container_1_medium_not_patchable_finding_patchableSeverityThreshold_High_0_violoations {
    input := { "review": input_review_unhealthy_container_with_not_patchable_severities, "parameters": input_parameters_medium_0_severityThresholdForExcludingNotPatchableFindings_High}
    results := violation with input as input
    count(results) == 0
}

# Checks if patchableSeverityThreshold set to High and there is High scanFinding, then there is no violation
test_input_review_unhealthy_container_1_high_not_patchable_finding_patchableSeverityThreshold_High_0_violoations {
    input := { "review": input_review_unhealthy_container_with_not_patchable_severities, "parameters": input_parameters_high_0_severityThresholdForExcludingNotPatchableFindings_High}
    results := violation with input as input
    count(results) == 0
}

# Checks if there are 2 containers, one with unscaneed result, and one with exceeding serverity, then there are 2 violations.
test_input_review_unhealthy_2_container_1_high_1_unscanned {
    input := { "review": input_review_unhealthy_2_containers_one_unscanned_and_one_high, "parameters": input_parameters_high_0_severityThresholdForExcludingNotPatchableFindings_None}
    results := violation with input as input
    count(results) == 2
}

input_review_no_annotations = {
    "object": {
        "metadata": {
            "annotations": ""
        }
    }
}

input_review_creation_time_ok_scan_status_unscanned = {
    "uid": "123",
    "object": {
        "metadata": {
            "annotations": {
                "azuredefender.io/containers.vulnerability.scan.info": "{\"generatedTimestamp\":\"2021-05-04T23:53:20Z\",\"uidRequest\":\"123\",\"containers\":[{\"name\":\"testContainer\",\"image\":{\"name\":\"tomer.azurecr.io/core/app:4.6\",\"digest\":\"sha256:4a1c4b21597c1b4415bdbecb28a3296c6b5e23ca4f9feeb599860a1dac6a0108\"},\"scanStatus\":\"unscanned\",\"scanFindings\":[{\"patchable\":true,\"id\":\"123\",\"severity\":\"High\"}]}]}"
            }
        }
    }
}

input_review_violation_with_stale_annotations = {
    "uid": "124",
    "object": {
        "metadata": {
            "annotations": {
                "azuredefender.io/containers.vulnerability.scan.info": "{\"generatedTimestamp\":\"2021-03-04T23:53:20Z\",\"uidRequest\":\"123\",\"containers\":[{\"name\":\"testContainer\",\"image\":{\"name\":\"tomer.azurecr.io/core/app:4.6\",\"digest\":\"sha256:4a1c4b21597c1b4415bdbecb28a3296c6b5e23ca4f9feeb599860a1dac6a0108\"},\"scanStatus\":\"unscanned\",\"scanFindings\":[{\"patchable\":true,\"id\":\"123\",\"severity\":\"High\"}]}]}"
            },
        "managedFields":[{"time": "2021-05-04T23:53:20Z"}]
        }
    }
}

input_review_violation_with_not_stale_annotations = {
    "uid": "124",
    "object": {
        "metadata": {
            "annotations": {
                "azuredefender.io/containers.vulnerability.scan.info": "{\"generatedTimestamp\":\"2021-03-04T23:53:20Z\",\"uidRequest\":\"123\",\"containers\":[{\"name\":\"testContainer\",\"image\":{\"name\":\"tomer.azurecr.io/core/app:4.6\",\"digest\":\"sha256:4a1c4b21597c1b4415bdbecb28a3296c6b5e23ca4f9feeb599860a1dac6a0108\"},\"scanStatus\":\"unscanned\",\"scanFindings\":[{\"patchable\":true,\"id\":\"123\",\"severity\":\"High\"}]}]}"
            },
        "managedFields":[{"time": "2021-03-04T23:53:20Z"}]
        }
    }
}

input_review_unhealthy_2_containers_with_diff_severities = {
    "uid": "123",
    "object": {
        "metadata": {
            "annotations": {
                "azuredefender.io/containers.vulnerability.scan.info": "{\"generatedTimestamp\":\"2021-05-04T23:53:20Z\",\"uidRequest\":\"123\",\"containers\":[{\"name\":\"testContainer\",\"image\":{\"name\":\"tomer.azurecr.io/core/app:4.6\",\"digest\":\"sha256:4a\"},\"scanStatus\":\"unhealthyScan\",\"scanFindings\":[{\"patchable\":true,\"id\":\"123\",\"severity\":\"High\"},{\"patchable\":true,\"id\":\"124\",\"severity\":\"High\"},{\"patchable\":true,\"id\":\"125\",\"severity\":\"High\"}]},{\"name\":\"testContainer2\",\"image\":{\"name\":\"tomer.azurecr.io/core/app:4.6\",\"digest\":\"sha256:4a\"},\"scanStatus\":\"unhealthyScan\",\"scanFindings\":[{\"patchable\":true,\"id\":\"124\",\"severity\":\"Low\"},{\"patchable\":true,\"id\":\"124\",\"severity\":\"Medium\"},{\"patchable\":true,\"id\":\"124\",\"severity\":\"High\"},{\"patchable\":true,\"id\":\"125\",\"severity\":\"High\"}]}]}"
            }
        }
    }
}


input_review_unhealthy_container_with_patchable_finding = {
    "uid": "123",
    "object": {
        "metadata": {
            "annotations": {
                "azuredefender.io/containers.vulnerability.scan.info": "{\"generatedTimestamp\":\"2021-05-04T23:53:20Z\",\"uidRequest\":\"123\",\"containers\":[{\"name\":\"testContainer2\",\"image\":{\"name\":\"tomer.azurecr.io/core/app:4.6\",\"digest\":\"sha256:4a\"},\"scanStatus\":\"unhealthyScan\",\"scanFindings\":[{\"patchable\":true,\"id\":\"125\",\"severity\":\"High\"}]}]}"
            }
        }
    }
}


input_review_unhealthy_container_with_not_patchable_finding = {
    "uid": "123",
    "object": {
        "metadata": {
            "annotations": {
                "azuredefender.io/containers.vulnerability.scan.info": "{\"generatedTimestamp\":\"2021-05-04T23:53:20Z\",\"uidRequest\":\"123\",\"containers\":[{\"name\":\"testContainer2\",\"image\":{\"name\":\"tomer.azurecr.io/core/app:4.6\",\"digest\":\"sha256:4a\"},\"scanStatus\":\"unhealthyScan\",\"scanFindings\":[{\"patchable\":false,\"id\":\"125\",\"severity\":\"High\"}]}]}"
            }
        }
    }
}

input_review_unhealthy_container_with_not_patchable_severities = {
    "uid": "123",
    "object": {
        "metadata": {
            "annotations": {
                "azuredefender.io/containers.vulnerability.scan.info": "{\"generatedTimestamp\":\"2021-05-04T23:53:20Z\",\"uidRequest\":\"123\",\"containers\":[{\"name\":\"testContainer\",\"image\":{\"name\":\"tomer.azurecr.io/core/app:4.6\",\"digest\":\"sha256:4a\"},\"scanStatus\":\"unhealthyScan\",\"scanFindings\":[{\"patchable\":false,\"id\":\"125\",\"severity\":\"Low\"},{\"patchable\":false,\"id\":\"126\",\"severity\":\"Medium\"},{\"patchable\":false,\"id\":\"127\",\"severity\":\"High\"}]}]}"
            }
        }
    }
}

input_review_unhealthy_container_with_one_finding_patchable_low_severity = {
    "uid": "123",
    "object": {
        "metadata": {
            "annotations": {
                "azuredefender.io/containers.vulnerability.scan.info": "{\"generatedTimestamp\":\"2021-05-04T23:53:20Z\",\"uidRequest\":\"123\",\"containers\":[{\"name\":\"testContainer2\",\"image\":{\"name\":\"tomer.azurecr.io/core/app:4.6\",\"digest\":\"sha256:4a\"},\"scanStatus\":\"unhealthyScan\",\"scanFindings\":[{\"patchable\":true,\"id\":\"125\",\"severity\":\"Low\"}]}]}"
            }
        }
    }
}

input_review_unhealthy_2_containers_one_unscanned_and_one_high = {
    "uid": "123",
    "object": {
        "metadata": {
            "annotations": {
                "azuredefender.io/containers.vulnerability.scan.info": "{\"generatedTimestamp\":\"2021-05-04T23:53:20Z\",\"uidRequest\":\"123\",\"containers\":[{\"name\":\"testContainer\",\"image\":{\"name\":\"tomer.azurecr.io/core/app:4.6\",\"digest\":\"sha256:4a\"},\"scanStatus\":\"unhealthyScan\",\"scanFindings\":[{\"patchable\":false,\"id\":\"125\",\"severity\":\"High\"}]},{\"name\":\"testContainer2\",\"image\":{\"name\":\"tomer.azurecr.io/core/app:4.6\",\"digest\":\"sha256:4a\"},\"scanStatus\":\"unscanned\",\"scanFindings\":[]}]}"
            }
        }
    }
}

input_parameters_empty = {}

input_parameters_tomerazurecr_image_excluded_severityHighTreshold_2 = {
    "excluded_images_pattern": ["(tomer.azurecr.io).*"],
    "severity" : {
        "High": 2,
    }
}

input_parameters_liorazurecr_image_excluded_severityHighTreshold_2 = {
    "excluded_images_pattern": ["(lior.azurecr.io).*"],
    "severity" : {
        "High": 2,
    }
}

input_parameters_severityHighTreshold_2 = {
    "severity" : {
        "High": 2,
    }
}

input_parameters_severityHighTreshold_1 = {
    "severity" : {
        "High": 1,
    }
}

input_parameters_severityMediumTreshold_0 = {
    "severity" : {
        "Medium" : 0,
    }
}

input_parameters_high_0_excluded_finding_id_125 = {
    "exlcudeFindingIDs" : ["125"],
    "severity" : {
        "High": 0
    }
}

input_parameters_high_0_excluded_finding_id_126 = {
    "exlcudeFindingIDs" : ["126"],
    "severity" : {
        "High": 0
    }
}

input_parameters_severityLowTreshold_0 = {
    "severity" : {
        "Low": 0
    }
}

input_parameters_low_0_severityThresholdForExcludingNotPatchableFindings_None = {
    "severityThresholdForExcludingNotPatchableFindings": "None",
    "severity" : {
        "Low": 0
    }
}

input_parameters_medium_0_severityThresholdForExcludingNotPatchableFindings_None = {
    "severityThresholdForExcludingNotPatchableFindings": "None",
    "severity" : {
        "Medium": 0
    }
}

input_parameters_high_0_severityThresholdForExcludingNotPatchableFindings_None = {
    "severityThresholdForExcludingNotPatchableFindings": "None",
    "severity" : {
        "High": 0
    }
}

input_parameters_low_0_severityThresholdForExcludingNotPatchableFindings_Low = {
    "severityThresholdForExcludingNotPatchableFindings": "Low",
    "severity" : {
        "Low": 0
    }
}

input_parameters_medium_0_severityThresholdForExcludingNotPatchableFindings_Low = {
    "severityThresholdForExcludingNotPatchableFindings": "Low",
    "severity" : {
        "Medium": 0
    }
}

input_parameters_high_0_severityThresholdForExcludingNotPatchableFindings_Low = {
    "severityThresholdForExcludingNotPatchableFindings": "Low",
    "severity" : {
        "High": 0
    }
}

input_parameters_low_0_severityThresholdForExcludingNotPatchableFindings_Medium = {
    "severityThresholdForExcludingNotPatchableFindings": "Medium",
    "severity" : {
        "Low": 0
    }
}

input_parameters_medium_0_severityThresholdForExcludingNotPatchableFindings_Medium = {
    "severityThresholdForExcludingNotPatchableFindings": "Medium",
    "severity" : {
        "Medium": 0
    }
}

input_parameters_high_0_severityThresholdForExcludingNotPatchableFindings_Medium = {
    "severityThresholdForExcludingNotPatchableFindings": "Medium",
    "severity" : {
        "High": 0
    }
}

input_parameters_low_0_severityThresholdForExcludingNotPatchableFindings_High = {
    "severityThresholdForExcludingNotPatchableFindings": "High",
    "severity" : {
        "Low": 0
    }
}

input_parameters_medium_0_severityThresholdForExcludingNotPatchableFindings_High = {
    "severityThresholdForExcludingNotPatchableFindings": "High",
    "severity" : {
        "Medium": 0
    }
}

input_parameters_high_0_severityThresholdForExcludingNotPatchableFindings_High = {
    "severityThresholdForExcludingNotPatchableFindings": "High",
    "severity" : {
        "High": 0
    }
}