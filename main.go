package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/coreos/go-systemd/v22/activation"
	"github.com/coreos/go-systemd/v22/daemon"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	shutdownTimeout := flag.Duration("shutdown-timeout", 5*time.Second, "shutdown timeout")
	flag.Parse()

	listeners, err := activation.Listeners()
	if err != nil {
		return err
	}
	if len(listeners) != 1 {
		return fmt.Errorf("want one socket from systemd socket activation, but got %d socket(s)", len(listeners))
	}

	http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		if wait := r.FormValue("wait"); wait != "" {
			if d, err := time.ParseDuration(wait); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "invalid duration at \"wait\" query paramter.\n")
				return
			} else {
				fmt.Fprintf(w, "Sleeping %s...\n", d)
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				time.Sleep(d)
			}
		}
		fmt.Fprintf(w, "Hello, %q\n", html.EscapeString(r.URL.Path))
	})

	s := &http.Server{
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	idleConnsClosed := make(chan struct{})
	var errShutdown error
	go func() {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		<-ctx.Done()

		sent, err := daemon.SdNotify(false, daemon.SdNotifyStopping)
		log.Printf("SdNotify stopping, sent=%v, err=%v", sent, err)

		ctx, cancel := context.WithTimeout(context.Background(), *shutdownTimeout)
		defer cancel()
		errShutdown = s.Shutdown(ctx)
		close(idleConnsClosed)
	}()

	sent, err := daemon.SdNotify(false, daemon.SdNotifyReady)
	log.Printf("SdNotify ready, sent=%v, err=%v", sent, err)

	err = s.Serve(listeners[0])
	if err == http.ErrServerClosed {
		err = nil
	}
	<-idleConnsClosed
	if err != nil {
		return errors.Join(err, errShutdown)
	}
	return nil
}
