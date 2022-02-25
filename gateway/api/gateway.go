package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"gateway/domain/intefaces"
	"gateway/domain/models"

	"github.com/gorilla/websocket"
)

type GatewayHandler struct {
	up         *websocket.Upgrader
	subscriber intefaces.SubscribeService
}

func NewGatewayHandler(upgrader *websocket.Upgrader, service intefaces.SubscribeService) *GatewayHandler {
	return &GatewayHandler{
		up:         upgrader,
		subscriber: service,
	}
}

const welcomeMsg = `Connection successful. Available commands : [
	"subscribe to all symbols" : {"action": "subscribe"},
	"subscribe to specific symbols: {"action" : "subscribe", "symbols": ["SHIBUSDT"]}",
	"unsubscribe : {"action": "unsubscribe"}"
]`

func (g *GatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := g.up.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer func() {
		err := c.Close()
		if err != nil {
			log.Printf("websocket connection close err: %s", err)
		}
	}()

	err = c.WriteMessage(1, []byte(welcomeMsg))
	if err != nil {
		log.Printf("write welcome messag err: %s", err)
		return
	}

	isSubscribed := false
	ctx, cancelFunc := context.WithCancel(r.Context())
	defer cancelFunc()

	responses := make(chan models.GatewayResponse, 10)
	dataCh := make(chan models.Data, 30)
	go g.writer(ctx, c, responses, dataCh)

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Printf("read message err: %s\n", err)
			break
		}

		request := models.GatewayRequest{}
		response := models.GatewayResponse{}
		err = json.Unmarshal(msg, &request)
		if err != nil {
			log.Printf("unmarshal request err: %s\n", err)
			response.Status = models.ErrorStatus
			response.SetError(unmarshalRequestErr)
			responses <- response
			continue
		}
		response.Request = request

		switch request.Action {
		case models.Subscribe:
			if isSubscribed {
				response.Status = models.ErrorStatus
				response.SetError(alreadySubscribe)
				responses <- response
				continue
			}
			symbols := request.Symbols
			if len(symbols) == 0 {
				symbols = []string{"all"}
			}
			err := g.subscriber.Subscribe(ctx, r.RemoteAddr, r.RemoteAddr, symbols, dataCh)
			if err != nil {
				log.Printf("subscribe request err: %s\n", err)
				response.Status = models.ErrorStatus
				response.SetError(subscribeRequestErr)
				responses <- response
				continue
			}
			isSubscribed = true
			response.Status = models.SuccessStatus
			responses <- response
		case models.Unsubscribe:
			err := g.subscriber.Unsubscribe(r.RemoteAddr)
			if err != nil {
				log.Printf("unsubscriber request err: %s\n", err)
				response.Status = models.ErrorStatus
				response.SetError(unsubscribeRequestErr)
				responses <- response
				continue
			}
			response.Status = models.SuccessStatus
			responses <- response
			isSubscribed = false
		}
	}
}

func (g *GatewayHandler) writer(ctx context.Context, conn *websocket.Conn, responses <-chan models.GatewayResponse, dataCh <-chan models.Data) {
	for {
		select {
		case <-ctx.Done():
			return
		case response := <-responses:
			err := conn.WriteJSON(response)
			if err != nil {
				log.Printf("write gatewayResponse err: %s\n", err)
			}
		case data := <-dataCh:
			err := conn.WriteJSON(data)
			if err != nil {
				log.Printf("write dataResponse err: %s\n", err)
			}
		}
	}
}
