package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// StartServer starts the server
func StartServer() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "443"
	}

	tlsDir := os.Getenv("TLS")
	if tlsDir == "" {
		tlsDir = "tls-local"
	}

	app := &App{}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.HandleRoot)
	mux.HandleFunc("/mutate", app.HandleMutate)

	pair, err := tls.LoadX509KeyPair(filepath.Join(tlsDir, "cert.pem"), filepath.Join(tlsDir, "key.pem"))
	if err != nil {
		log.Println(err)
		return err
	}

	s := &http.Server{
		Addr:           fmt.Sprintf(":%s", port),
		Handler:        mux,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1048576
		TLSConfig:      &tls.Config{Certificates: []tls.Certificate{pair}},
	}

	fmt.Printf("Listening on port %s\n", port)

	return s.ListenAndServeTLS("", "")
}
