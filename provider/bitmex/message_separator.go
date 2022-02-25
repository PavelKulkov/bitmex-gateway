package bitmex

import (
	"context"
	"encoding/json"
	"log"
)


type MessageSeparator struct {
	ctx          context.Context
	cancelFunc   context.CancelFunc
	src          <-chan []byte
	commandOut   chan<- CommandResponse
	subscribeOut chan<- SubscribeResponse
}

func NewMessageSeparator(
	ctx context.Context,
	source <-chan []byte,
	commandResponseCh chan<- CommandResponse,
	subscribeResponseCh chan<- SubscribeResponse) *MessageSeparator {
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	return &MessageSeparator{
		ctx:          cancelCtx,
		cancelFunc:   cancelFunc,
		src:          source,
		commandOut:   commandResponseCh,
		subscribeOut: subscribeResponseCh,
	}
}

func (m *MessageSeparator) Run() {
	log.Println("MessageSeparator run")
	for {
		select {
		case <-m.ctx.Done():
			return
		default:
			b := <-m.src

			response := SubscribeResponse{}
			err := json.Unmarshal(b, &response)
			if err != nil {
				log.Println(err)
				continue
			}
			if response.Action == "" && response.Table == "" && len(response.Data) == 0 {
				commandResponse := CommandResponse{}
				err := json.Unmarshal(b, &commandResponse)
				if err != nil {
					log.Println(err)
					continue
				}
				if commandResponse.Subscribe == "" {
					continue
				}
				m.commandOut <- commandResponse
				continue
			}
			m.subscribeOut <- response
		}
	}
}

func (m *MessageSeparator) Stop() {
	close(m.commandOut)
	close(m.subscribeOut)
	m.cancelFunc()
	log.Println("MessageSeparator stopped")
}
