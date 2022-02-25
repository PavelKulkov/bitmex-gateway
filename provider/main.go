package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"

	"provider/bitmex"
	"provider/rabbitmq"

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

	websocketClient := bitmex.NewWebsocketClient(appCtx, "wss://ws.testnet.bitmex.com/realtime")
	err := websocketClient.Connect(nil)
	if err != nil {
		return fmt.Errorf("websocket client connnect err: %w", err)
	}

	err = websocketClient.Subscribe([]string{"instrument"})
	if err != nil {
		return fmt.Errorf("websocket client subscribe instrument err: %w", err)
	}

	commandResponseCh := make(chan bitmex.CommandResponse, 10)
	subscriberResponseCh := make(chan bitmex.SubscribeResponse, 10)

	go func() {
		for {
			select {
			case <-appCtx.Done():
				return
			default:
				responseCh := <-commandResponseCh
				log.Printf("commandResponse: %v\n", responseCh)
			}
		}
	}()

	separator := bitmex.NewMessageSeparator(appCtx, websocketClient.ReadChannel(), commandResponseCh, subscriberResponseCh)
	go separator.Run()

	mapper := bitmex.NewSubscriberResponseMapper(appCtx, subscriberResponseCh)

	dial, err := amqp.Dial(rmqUrl)
	if err != nil {
		panic(err)
	}

	channel, err := dial.Channel()
	if err != nil {
		return fmt.Errorf("create rabbitmq channel err: %w", err)
	}

	err = channel.ExchangeDeclare("symbols", amqp.ExchangeDirect, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("exchanger declare err: %w", err)
	}

	producer := rabbitmq.NewProducer("symbols", channel)
	manager := NewDataManager(appCtx)

	go func() { manager.Run(mapper, producer) }()

	<-appCtx.Done()

	err = websocketClient.Close()
	if err != nil {
		log.Printf("websocketClient close err: %s\n", err)
	}

	separator.Stop()
	mapper.Stop()
	manager.Stop()

	err = channel.Close()
	if err != nil {
		log.Printf("rabbitmq channel close err: %s\n", err)
	}

	err = dial.Close()
	if err != nil {
		log.Printf("amqp dial close err: %s\n", err)
	}

	return nil
}
