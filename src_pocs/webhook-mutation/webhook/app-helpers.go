package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

// http helpers

// HandleError depending on error type
func (app *App) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	jsonError(w, err.Error(), http.StatusBadRequest)
}

// readJSON from request body
func readJSON(r *http.Request, v interface{}) error {
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		return fmt.Errorf("invalid JSON input")
	}

	return nil
}

// jsonOk renders json with 200 ok
func jsonOk(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, v)
}

// writeJSON to response body
func writeJSON(w http.ResponseWriter, v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		http.Error(w, fmt.Sprintf("json encoding error: %v", err), http.StatusInternalServerError)
		return
	}

	writeBytes(w, b)
}

// writeBytes to response body
func writeBytes(w http.ResponseWriter, b []byte) {
	_, err := w.Write(b)
	if err != nil {
		http.Error(w, fmt.Sprintf("write error: %v", err), http.StatusInternalServerError)
		return
	}
}

// jsonError renders json with error
func jsonError(w http.ResponseWriter, errStr string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	writeJSON(w, &jsonErr{Err: errStr})
}

// jsonErr err
type jsonErr struct {
	Err string `json:"err"`
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
