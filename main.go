package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"time"

	"github.com/coreos/go-systemd/v22/activation"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	listeners, err := activation.Listeners()
	if err != nil {
		return err
	}
	if len(listeners) != 1 {
		return fmt.Errorf("want one socket from systemd socket activation, but got %d socket(s)", len(listeners))
	}

	http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q\n", html.EscapeString(r.URL.Path))
	})

	s := &http.Server{
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	return s.Serve(listeners[0])
}
