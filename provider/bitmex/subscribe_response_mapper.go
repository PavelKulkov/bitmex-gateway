package bitmex

import (
	"context"
	"log"

	"provider/domain/models"
)

type SubscribeResponseMapper struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	in         <-chan SubscribeResponse
	out        chan models.Data
}

func NewSubscriberResponseMapper(ctx context.Context, in <-chan SubscribeResponse) *SubscribeResponseMapper {
	cancel, cancelFunc := context.WithCancel(ctx)
	return &SubscribeResponseMapper{
		ctx:        cancel,
		cancelFunc: cancelFunc,
		in:         in,
		out:        make(chan models.Data, 10),
	}
}

func (s *SubscribeResponseMapper) Run() <-chan models.Data {
	go func() {
		log.Println("Response mapper run")
		for {
			select {
			case <-s.ctx.Done():
				close(s.out)
				return
			default:
				resp := <-s.in
				for _, info := range resp.Data {
					price := info.LastPrice
					switch {
					case info.AskPrice != 0:
						price = info.AskPrice
					case info.MarkPrice != 0:
						price = info.MarkPrice
					}

					if price == 0 {
						continue
					}
					s.out <- models.Data{
						Symbol:    info.Symbol,
						Price:     price,
						Timestamp: info.Timestamp,
					}
				}
			}
		}
	}()
	return s.out
}

func (s *SubscribeResponseMapper) Stop() {
	s.cancelFunc()
	log.Println("Response mapper stopped")
}
