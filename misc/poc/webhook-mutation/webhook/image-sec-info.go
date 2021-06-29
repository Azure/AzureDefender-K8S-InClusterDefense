package main

import (
	"io"
	"strings"

	imageParser "github.com/novln/docker-parser"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var (
	debug = pflag.Bool("debug", true, "sets log to debug level")
)

// Image
type Image struct {
	Registry string `json:"registry,omitempty"`
	Repo     string `json:"repo,omitempty"`
	Tag      string `json:"tag,omitempty"`
	Digest   string `json:"digest,omitempty"`
}

// Image Security Info
type ImageSecInfo struct {
	//ScanStatus
	Image Image `json:"image,omitempty"`
	//ScanStatus
	ScanStatus *string `json:"scanStatus,omitempty"`
	// SeveritySummary
	SeveritySummary map[string]int `json:"severitySummary,omitempty"`
}

// LogHook is used to setup custom hooks
type LogHook struct {
	Writer    io.Writer
	Loglevels []log.Level
}

// Get the security
func GetImageSecInfo(imageAsString string) (imageSecInfo *ImageSecInfo, err error) {
	setupLogger()
	argProxy, err := NewARGProxy()
	if err != nil {
		log.Fatalf("[error] : %v", err)
	}

	if imageAsString == "" {
		log.Info("Failed to provide image to query")
		return nil, err
	}
	image := NewImage(imageAsString)
	if strings.EqualFold(image.Digest, "") {
		log.Errorf("Digest is empty.")
		return nil, err
	}

	imageSecInfo, err = argProxy.GetImageSecInfo(image)
	if err != nil {
		log.Infof("[error] : %s", err)
		return nil, err
	}
	imageSecInfo.Image = image // Assign image to image security info
	return imageSecInfo, nil
}

func NewImage(imageStr string) (image Image) {
	parsedImage, err := imageParser.Parse(imageStr)
	if err != nil {
		log.Error("Invalid imagestring")
		return Image{}
	}

	log.Debugf("Image after parsing (ref): %s", parsedImage)
	image = Image{
		Registry: parsedImage.Registry(),  // registry := "upstream.azurecr.io"
		Repo:     parsedImage.ShortName(), // repo := "oss/kubernetes/ingress/nginx-ingress-controller"
		Tag:      parsedImage.Tag(),       // tag := "0.16.2"
	}
	// In case that the image was deployed with the digest:
	if strings.EqualFold(image.Tag, "") {
		image.Tag = "latest" // Default tag.
	}
	if strings.Contains(image.Tag, "sha256:") {
		image.Digest = parsedImage.Tag()
		image.Tag = "" // Assign empty tag in case the image contains digest
	} else { // else, extract digest first.
		image.Digest = tag2Digest(image)
	}
	log.Debugf("Image: %s", image)
	return image
}
