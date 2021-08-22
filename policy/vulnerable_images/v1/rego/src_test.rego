package k8sazuredefendervulnerableimages

# Check that the annotations is empty there is no violation.
test_input_no_annotations {
    input := { "review": input_review_no_annotations, "parameters": input_parameters_empty}
    results := violation with input as input
    count(results) == 0
}

# Check that if the annotations are old i.e. are not updated in the last __ seconds, there is no violoation.
test_input_grace_annotations {
    input := { "review": input_review_violation_with_grace_annotations, "parameters": input_parameters_empty}
    results := violation with input as input
    count(results) == 0
}

# Checks that if there is unscanned msg and the image is appers in the excluded images regex list, then there is no violation.
test_input_image_appear_in_excluded_images {
    input := { "review": input_review_creation_time_ok_scan_status_unscanned, "parameters": input_parameters_tomerazurecr_image_excluded}
    results := violation with input as input
    count(results) == 0
}

# Checks that if there is unscanned image, then there is 1 violtation.
test_input_creation_time_ok_scan_status_unscanned {
    input := { "review": input_review_creation_time_ok_scan_status_unscanned, "parameters": input_parameters_empty}
    results := violation with input as input    
    count(results) == 1
}

# Checks that if there is one container that is complainet (high sevirity) and one that isn't then we get only 1 violation.
test_input_review_unhealthy_one_container_above_highSeverity_one_violotation {
    input := { "review": input_review_unhealthy_2_containers_with_diff_severities, "parameters": input_parameters_sevirityHighTreshold_2}
    results := violation with input as input
    count(results) == 1
}

# Checks that if there are 2 conatiners with higher highSevirity than the treshold then we get 2 viloations.
test_input_review_unhealthy_two_containers_above_highSeverity_two_violotation {
    input := { "review": input_review_unhealthy_2_containers_with_diff_severities, "parameters": input_parameters_sevirityHighTreshold_1}
    results := violation with input as input
    count(results) == 2
}


test_input_review_unhealthy_one_container_above_mediumSeverity_one_violotation {
    input := { "review": input_review_unhealthy_2_containers_with_diff_severities, "parameters": input_parameters_sevirityMediumTreshold_0}
    results := violation with input as input
    count(results) == 1
}

test_input_review_unhealthy_one_container_above_lowSeverity_one_violotation {
    input := { "review": input_review_unhealthy_2_containers_with_diff_severities, "parameters": input_parameters_sevirityLowTreshold_0}
    results := violation with input as input
    count(results) == 1
}

test_input_review_unhealthy_exclude_not_patchable_findings_zero_violoations {
    input := { "review": input_review_unhealthy_container_with_not_patchable_finding, "parameters": input_parameters_high_0_exclude_not_patchable_findings}
    results := violation with input as input
    count(results) == 0
}

test_input_review_unhealthy_include_not_patchable_findings_one_violoations {
    input := { "review": input_review_unhealthy_container_with_not_patchable_finding, "parameters": input_parameters_high_0_include_not_patchable_findings}
    results := violation with input as input
    count(results) == 1
}

test_input_review_unhealthy_cotainer_with_1_high_finding_that_is_appears_in_exlcudeFindingIDs_zero_violoations {
    input := { "review": input_review_unhealthy_container_with_not_patchable_finding, "parameters": input_parameters_high_0_excluded_finding_id_125}
    results := violation with input as input
    count(results) == 0
}

test_input_review_unhealthy_cotainer_with_1_high_finding_that_isnt_appears_in_exlcudeFindingIDs_one_violations {
    input := { "review": input_review_unhealthy_container_with_not_patchable_finding, "parameters": input_parameters_high_0_excluded_finding_id_126}
    results := violation with input as input
    count(results) == 1
}



input_review_no_annotations = {
    "object": {
        "metadata": {
            "annotations": ""
        },
        "creationTimestamp": "2021-05-04T23:53:20Z"
    }
}

input_review_creation_time_ok_scan_status_unscanned = {
    "object": {
        "metadata": {
            "annotations": {
                "azuredefender.io/containers.vulnerability.scan.info": "{\"generatedTimestamp\":\"2021-05-04T23:53:20Z\",\"containers\":[{\"name\":\"testContainer\",\"image\":{\"name\":\"tomer.azurecr.io/core\/app:4.6\",\"digest\":\"sha256:4a1c4b21597c1b4415bdbecb28a3296c6b5e23ca4f9feeb599860a1dac6a0108\"},\"scanStatus\":\"unscanned\",\"scanFindings\":[{\"patchable\":true,\"id\":\"123\",\"severity\":\"High\"}]}]}"
            },
        "creationTimestamp": "2021-05-04T23:53:20Z"
        }
    }
}


input_review_violation_with_grace_annotations = {
    "object": {
        "metadata": {
            "annotations": {
                "azuredefender.io/containers.vulnerability.scan.info": "{\"generatedTimestamp\":\"2021-03-04T23:53:20Z\",\"containers\":[{\"name\":\"testContainer\",\"image\":{\"name\":\"tomer.azurecr.io/core\/app:4.6\",\"digest\":\"sha256:4a1c4b21597c1b4415bdbecb28a3296c6b5e23ca4f9feeb599860a1dac6a0108\"},\"scanStatus\":\"unscanned\",\"scanFindings\":[{\"patchable\":true,\"id\":\"123\",\"severity\":\"High\"}]}]}"
            },
        "creationTimestamp": "2021-05-04T23:53:20Z"
        }
    }
}


input_review_unhealthy_2_containers_with_diff_severities = {
    "object": {
        "metadata": {
            "annotations": {
                "azuredefender.io/containers.vulnerability.scan.info": "{\"generatedTimestamp\":\"2021-05-04T23:53:20Z\",\"containers\":[{\"name\":\"testContainer\",\"image\":{\"name\":\"tomer.azurecr.io/core\/app:4.6\",\"digest\":\"sha256:4a\"},\"scanStatus\":\"unhealthyScan\",\"scanFindings\":[{\"patchable\":true,\"id\":\"123\",\"severity\":\"High\"},{\"patchable\":true,\"id\":\"124\",\"severity\":\"High\"},{\"patchable\":true,\"id\":\"125\",\"severity\":\"High\"}]},{\"name\":\"testContainer2\",\"image\":{\"name\":\"tomer.azurecr.io/core\/app:4.6\",\"digest\":\"sha256:4a\"},\"scanStatus\":\"unhealthyScan\",\"scanFindings\":[{\"patchable\":true,\"id\":\"124\",\"severity\":\"Low\"},{\"patchable\":true,\"id\":\"124\",\"severity\":\"Medium\"},{\"patchable\":true,\"id\":\"124\",\"severity\":\"High\"},{\"patchable\":true,\"id\":\"125\",\"severity\":\"High\"}]}]}"
            },
        "creationTimestamp": "2021-05-04T23:53:20Z"
        }
    }
}

input_review_unhealthy_container_with_not_patchable_finding = {
    "object": {
        "metadata": {
            "annotations": {
                "azuredefender.io/containers.vulnerability.scan.info": "{\"generatedTimestamp\":\"2021-05-04T23:53:20Z\",\"containers\":[{\"name\":\"testContainer2\",\"image\":{\"name\":\"tomer.azurecr.io/core\/app:4.6\",\"digest\":\"sha256:4a\"},\"scanStatus\":\"unhealthyScan\",\"scanFindings\":[{\"patchable\":false,\"id\":\"125\",\"severity\":\"High\"}]}]}"
            },
        "creationTimestamp": "2021-05-04T23:53:20Z"
        }
    }
}

input_parameters_empty = {}

input_parameters_tomerazurecr_image_excluded = {
    "excluded_images_pattern": ["(tomer\\.azurecr\\.io).*"]
}

input_parameters_sevirityHighTreshold_2 = {
    "sevirity" : {
        "High": 2,
    }
}

input_parameters_sevirityHighTreshold_1 = {
    "sevirity" : {
        "High": 1,
    }
}

input_parameters_sevirityMediumTreshold_0 = {
    "sevirity" : {
        "Medium" : 0,
    }
}

input_parameters_sevirityLowTreshold_0 = {
    "sevirity" : {
        "Low": 0
    }
}

input_parameters_high_0_exclude_not_patchable_findings = {
    "excludeNotPatchableFindings": true,
    "sevirity" : {
        "High": 0
    }
}

input_parameters_high_0_include_not_patchable_findings = {
    "excludeNotPatchableFindings": false,
    "sevirity" : {
        "High": 0
    }
}

input_parameters_high_0_excluded_finding_id_125 = {
    "exlcudeFindingIDs" : ["125"],
    "sevirity" : {
        "High": 0
    }
}

input_parameters_high_0_excluded_finding_id_126 = {
    "exlcudeFindingIDs" : ["126"],
    "sevirity" : {
        "High": 0
    }
}