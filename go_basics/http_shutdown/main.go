package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 1. HTTP-сервер с graceful shutdown
// Реализовать сервер с обработкой SIGTERM, корректным завершением текущих запросов.

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}

func main() {
	handler := http.NewServeMux()
	handler.HandleFunc("/hello", hello)

	server := &http.Server{
		Handler: handler,
		Addr:    ":8080",
	}

	go func() {
		defer func() {
			r := recover()
			if r != nil {
				log.Printf("PANIC - %s", r)
			}
		}()
		err := server.ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Println("server was closed")
				return
			}
			log.Printf("ERROR - %s\n", err.Error())
		}
	}()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGTERM, syscall.SIGINT)

	<-sigchan

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	go func() {
		defer func() {
			r := recover()
			if r != nil {
				log.Printf("PANIC - %s", r)
			}
		}()
		err := server.Shutdown(ctx)
		if err != nil {
			log.Printf("ERROR - %s\n", err.Error())
		}
	}()

	<-ctx.Done()
}
