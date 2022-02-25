package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"gateway/api"
	"gateway/service/subscribe_service"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	rmqUrl := os.Getenv("RMQ_URL")
	if rmqUrl == "" {
		return errors.New("RMQ_URL is empty")
	}

	appCtx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancelFunc()

	dial, err := amqp.Dial(rmqUrl)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := dial.Close()
		if err != nil {
			log.Printf("amqp dial close err: %s\n", err)
		}
	}()

	channel, err := dial.Channel()
	if err != nil {
		return fmt.Errorf("create rabbitmq channel err: %w", err)
	}
	defer func() {
		err := channel.Close()
		if err != nil {
			log.Printf("rabbitmq channel create err: %s", err)
		}
	}()

	service := subscribe_service.New(channel, "symbols")

	handler := api.NewGatewayHandler(&websocket.Upgrader{}, service)

	mux := http.NewServeMux()
	mux.Handle("/gateway", handler)
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
		BaseContext: func(_ net.Listener) context.Context {
			return appCtx
		},
	}

	go func() {
		log.Printf("http server successfully started on port 8080")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen and serve err: %s", err)
		}
	}()

	<-appCtx.Done()
	timeoutCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()

	return server.Shutdown(timeoutCtx)
}
