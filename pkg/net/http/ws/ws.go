package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/peterlearn/kratos/v1/pkg/log"
	"net/http"
	"sync"
	"time"
)

const (
	_Ping = 4
	_Pong = 5
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type _HeartBeat struct {
	NowTime int `json:"nowTime"`
}

type _Message struct {
	Cmd  int         `json:"cmd"`
	Data interface{} `json:"data"`
}

type WsConnContext struct {
	context.Context
	ws           *websocket.Conn
	wg           *sync.WaitGroup
	lock         *sync.Mutex
	lastPingTime time.Time
}

func New(c *gin.Context) (*WsConnContext, error) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return nil, err
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	return &WsConnContext{
		ws:           ws,
		wg:           wg,
		lock:         &sync.Mutex{},
		lastPingTime: time.Now(),
	}, nil
}

func (w *WsConnContext) Wait() {
	for {
		_, message, err := w.ws.ReadMessage()
		if err != nil {
			w.Close()
			break
		}

		msg := &_Message{}
		err = json.Unmarshal(message, msg)
		if err != nil {
			continue
		}

		if msg.Cmd == _Ping {
			w.Pong()
		} else if msg.Cmd == _Pong {
			w.lastPingTime = time.Now()
		} else {
			log.Info("not support!!!")
		}
	}

	w.wg.Wait()
}

func (w *WsConnContext) Pong() {
	heart := &_HeartBeat{
		NowTime: int(time.Now().Unix()),
	}
	buf, err := json.Marshal(heart)
	if err != nil {
		return
	}
	if err := w.Send(_Pong, buf); err != nil {
		log.Error("Send Pong err ", err)
	}
}

func (w *WsConnContext) Send(cmd int, data []byte) error {
	msg := &_Message{
		Cmd:  cmd,
		Data: data,
	}
	buf, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return w.Write(buf)
}

func (w *WsConnContext) Write(buf []byte) error {
	if w.ws == nil {
		return fmt.Errorf("send ws is nil")
	}
	w.lock.Lock()
	err := w.ws.WriteMessage(websocket.TextMessage, buf)
	w.lock.Unlock()
	return err
}

func (w *WsConnContext) Close() {
	w.ws.Close()
	w.wg.Done()
}

func (w *WsConnContext) Deadline() (deadline time.Time, ok bool) {
	return
}

func (w *WsConnContext) Done() <-chan struct{} {
	return nil
}

func (w *WsConnContext) Err() error {
	return nil
}

func (w *WsConnContext) Value(key interface{}) interface{} {
	return nil
}
