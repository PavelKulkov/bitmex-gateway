package main

import (
	"context"
	"log"

	"provider/domain/intefaces"
)

type DataManager struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewDataManager(ctx context.Context) *DataManager {
	cancel, cancelFunc := context.WithCancel(ctx)
	return &DataManager{
		ctx:        cancel,
		cancelFunc: cancelFunc,
	}
}

func (d *DataManager) Run(src intefaces.DataSource, producer intefaces.Publisher) {
	dataCh := src.Run()
	log.Println("Data manager run")
	for {
		select {
		case <-d.ctx.Done():
			return
		default:
			data := <-dataCh
			err := producer.Publish(data)
			if err != nil {
				log.Printf("data publish err: %s\n", err)
				continue
			}
		}
	}
}

func (d *DataManager) Stop() {
	d.cancelFunc()
	log.Println("Data manager stopped")
}
