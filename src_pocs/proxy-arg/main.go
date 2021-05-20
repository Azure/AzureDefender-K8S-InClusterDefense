package main

import (
	"bytes"
	"context"
	"encoding/json"

	//"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var (
	debug  = pflag.Bool("debug", true, "sets log to debug level")
	server *Server
	ctx    context.Context
)

// LogHook is used to setup custom hooks
type LogHook struct {
	Writer    io.Writer
	Loglevels []log.Level
}

func main() {
	//pflag.Parse()

	var err error

	setupLogger()

	ctx = context.Background()
	server, err = NewServer()
	if err != nil {
		log.Fatalf("[error] : %v", err)
	}
	http.HandleFunc("/process", handle)
	http.ListenAndServe(":8090", nil)

	os.Exit(0)
}

func handle(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	image := req.URL.Query().Get("image") // e.g. : oss/kubernetes/aks/etcd-operator
	if image == "" {
		log.Info("Failed to provide image to query")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(nil)
	}

	// registry := "upstream.azurecr.io"
	// repo := "oss/kubernetes/ingress/nginx-ingress-controller"
	// tag := "0.16.2"
	registry := strings.Split(image, "/")[0]
	repo := strings.Replace(image, registry+"/", "", 1)
	tag := "latest"
	if strings.Contains(repo, ":") {
		tag = strings.Split(repo, ":")[1]
		repo = strings.Replace(repo, ":"+tag, "", 1)
	}
	digest := ""
	// In case that the image deploted with the digest:
	if strings.Contains(image, "@sha256") {
		digest = strings.Split(image, "@")[1]
		// else, extract digest first.
	} else {
		getImageShaBinary := "getimagesha.sh"
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		cmd := exec.Command(
			"sh",
			getImageShaBinary,
			registry,
			repo,
			tag,
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
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(nil)
		} else {
			digest = strings.TrimSuffix(output, "\n")
		}
	}
	if strings.EqualFold(digest, "") {
		log.Infof("Digest is empty.")
		return
	}
	log.Infof("digest: %s", digest)
	data, err := server.Process(ctx, digest)
	if err != nil {
		log.Infof("[error] : %s", err)
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(data)
	} else {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
	}

}

// setupLogger sets up hooks to redirect stdout and stderr
func setupLogger() {
	log.SetOutput(ioutil.Discard)

	// set log level
	log.SetLevel(log.InfoLevel)
	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	// add hook to send info, debug, warn level logs to stdout
	log.AddHook(&LogHook{
		Writer: os.Stdout,
		Loglevels: []log.Level{
			log.InfoLevel,
			log.DebugLevel,
			log.WarnLevel,
		},
	})

	// add hook to send panic, fatal, error logs to stderr
	log.AddHook(&LogHook{
		Writer: os.Stderr,
		Loglevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
		},
	})
}

// Fire is called when logging function with current hook is called
// write to appropriate writer based on log level
func (hook *LogHook) Fire(entry *log.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write([]byte(line))
	return err
}

// Levels defines log levels at which hook is triggered
func (hook *LogHook) Levels() []log.Level {
	return hook.Loglevels
}
