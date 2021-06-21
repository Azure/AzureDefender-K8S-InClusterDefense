package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	ShellScriptFileName string = "tag_2_digest.sh"
)

// Convert tag to digest - using shell script (getimagesha.sh)
func tag2Digest(image Image) (digest string) {
	if image.Registry == "" || image.Repo == "" || image.Tag == "" {
		log.Errorf("Invalid image - registry or repo or tag are empty : registry = %s, repo = %s, tag = %s", image.Registry, image.Repo, image.Tag)
		return ""
	}
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	cmd := exec.Command(
		"sh",
		ShellScriptFileName,
		image.Registry,
		image.Repo,
		image.Tag,
	)

	log.Infof("cmd: %v", cmd)
	cmd.Dir = dir
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stderr, cmd.Stdout = stderr, stdout

	err = cmd.Run()
	output := stdout.String()
	log.Infof("output: %s", output)
	if err != nil {
		log.Errorf("error invoking cmd, err: %v, output: %v", err, stderr.String())
	}

	if output == "null\n" {
		log.Infof("[error] : could not find valid digest %s", output)
	} else {
		digest = strings.TrimSuffix(output, "\n")
	}
	//TODO Check that the digest is valid.
	image.Digest = digest
	log.Infof("Digest successfully extracted: %s", digest)
	return digest
}
