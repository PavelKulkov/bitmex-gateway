package bitmex

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebsocketClient struct {
	url           string
	conn          *websocket.Conn
	readCh        chan []byte
	writeCh       chan []byte
	ctx           context.Context
	cancelFuncCtx context.CancelFunc
}

func NewWebsocketClient(ctx context.Context, url string) *WebsocketClient {
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	return &WebsocketClient{
		url:           url,
		ctx:           cancelCtx,
		cancelFuncCtx: cancelFunc,
		writeCh:       make(chan []byte),
		readCh:        make(chan []byte),
	}
}

func (c *WebsocketClient) Connect(requestHeader http.Header) error {
	dial, _, err := websocket.DefaultDialer.DialContext(c.ctx, c.url, requestHeader)
	if err != nil {
		return fmt.Errorf("dial err: %w", err)
	}
	response := ConnectResponse{}
	err = dial.ReadJSON(&response)
	if err != nil {
		return fmt.Errorf("read connect response err: %w", err)
	}
	log.Println(response)
	c.conn = dial
	go c.writer()
	go c.reader()
	return nil
}

func (c *WebsocketClient) Close() error {
	c.cancelFuncCtx()
	return c.conn.Close()
}

func (c *WebsocketClient) Subscribe(topics []string) error {
	subscribeCommand := Command{
		Op:   subscribe,
		Args: topics,
	}
	err := c.execCommand(subscribeCommand)
	if err != nil {
		return err
	}
	return nil
}

func (c *WebsocketClient) Unsubscribe() error {
	unsubscribeCommand := Command{
		Op: unsubscribe,
	}
	err := c.execCommand(unsubscribeCommand)
	if err != nil {
		return err
	}
	return nil
}

func (c *WebsocketClient) ReadChannel() <-chan []byte {
	return c.readCh
}

func (c *WebsocketClient) reader() {
	for {
		select {
		case <-c.ctx.Done():
			close(c.readCh)
			return
		default:
			_, p, err := c.conn.ReadMessage()
			if err != nil {
				log.Println(err)
				continue
			}
			c.readCh <- p
		}
	}
}

func (c *WebsocketClient) writer() {
	for {
		select {
		case <-c.ctx.Done():
			close(c.writeCh)
			return
		default:
			bytes := <-c.writeCh
			err := c.conn.WriteMessage(1, bytes)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (c *WebsocketClient) execCommand(command Command) error {
	marshal, err := json.Marshal(command)
	if err != nil {
		return err
	}
	c.writeCh <- marshal
	return nil
}
