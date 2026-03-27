package net

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/forgoer/openssl"
	"github.com/gorilla/websocket"

	clog "github.com/actorgo-game/actorgo/logger"
	"github.com/llr104/slgserver/internal/util"
)

type ServerConn struct {
	wsSocket    *websocket.Conn
	outChan     chan *WsMsgRsp
	isClosed    bool
	needSecret  bool
	Seq         int64
	router      *Router
	beforeClose func(conn WSConn)
	onClose     func(conn WSConn)
	property    map[string]interface{}
	propertyLock sync.RWMutex
}

func NewServerConn(wsSocket *websocket.Conn, needSecret bool) *ServerConn {
	conn := &ServerConn{
		wsSocket:   wsSocket,
		outChan:    make(chan *WsMsgRsp, 1000),
		isClosed:   false,
		property:   make(map[string]interface{}),
		needSecret: needSecret,
		Seq:        0,
	}
	return conn
}

func (c *ServerConn) Start() {
	go c.wsReadLoop()
	go c.wsWriteLoop()
}

func (c *ServerConn) Addr() string {
	return c.wsSocket.RemoteAddr().String()
}

func (c *ServerConn) Push(name string, data interface{}) {
	rsp := &WsMsgRsp{Body: &RspBody{Name: name, Msg: data, Seq: 0}}
	c.outChan <- rsp
}

func (c *ServerConn) wsReadLoop() {
	defer func() {
		if err := recover(); err != nil {
			clog.Error("wsReadLoop error: %v", fmt.Sprintf("%v", err))
			c.Close()
		}
	}()

	for {
		_, data, err := c.wsSocket.ReadMessage()
		if err != nil {
			break
		}

		data, err = util.UnZip(data)
		if err != nil {
			clog.Error("wsReadLoop UnZip error: %v", err)
			continue
		}

		body := &ReqBody{}
		if c.needSecret {
			if secretKey, err := c.GetProperty("secretKey"); err == nil {
				key := secretKey.(string)
				d, err := util.AesCBCDecrypt(data, []byte(key), []byte(key), openssl.ZEROS_PADDING)
				if err != nil {
					clog.Error("AesDecrypt error: %v", err)
					c.Handshake()
				} else {
					data = d
				}
			} else {
				clog.Info("secretKey not found, client need handshake")
				c.Handshake()
				return
			}
		}

		if err := json.Unmarshal(data, body); err == nil {
			req := &WsMsgReq{Conn: c, Body: body}
			rsp := &WsMsgRsp{Body: &RspBody{Name: body.Name, Seq: req.Body.Seq}}

			if req.Body.Name == HeartbeatMsg {
				h := &Heartbeat{}
				raw, _ := json.Marshal(body.Msg)
				json.Unmarshal(raw, h)
				h.STime = time.Now().UnixNano() / 1e6
				rsp.Body.Msg = h
			} else {
				if c.router != nil {
					c.router.Run(req, rsp)
				}
			}
			c.outChan <- rsp
		} else {
			clog.Error("wsReadLoop Unmarshal error: %v", err)
			c.Handshake()
		}
	}

	c.Close()
}

func (c *ServerConn) wsWriteLoop() {
	defer func() {
		if err := recover(); err != nil {
			clog.Error("wsWriteLoop error")
			c.Close()
		}
	}()

	for msg := range c.outChan {
		c.write(msg.Body)
	}
}

func (c *ServerConn) write(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		clog.Error("wsWriteLoop Marshal body error: %v", err)
		return err
	}

	if c.needSecret {
		if secretKey, err := c.GetProperty("secretKey"); err == nil {
			key := secretKey.(string)
			data, _ = util.AesCBCEncrypt(data, []byte(key), []byte(key), openssl.ZEROS_PADDING)
		}
	}

	if zipped, err := util.Zip(data); err == nil {
		if err := c.wsSocket.WriteMessage(websocket.BinaryMessage, zipped); err != nil {
			c.Close()
			return err
		}
	} else {
		return err
	}
	return nil
}

func (c *ServerConn) Close() {
	c.wsSocket.Close()
	if !c.isClosed {
		c.isClosed = true
		if c.beforeClose != nil {
			c.beforeClose(c)
		}
		if c.onClose != nil {
			c.onClose(c)
		}
	}
}

func (c *ServerConn) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	c.property[key] = value
}

func (c *ServerConn) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()
	if value, ok := c.property[key]; ok {
		return value, nil
	}
	return nil, errors.New("no property found")
}

func (c *ServerConn) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	delete(c.property, key)
}

func (c *ServerConn) SetRouter(router *Router) {
	c.router = router
}

func (c *ServerConn) SetOnClose(hookFunc func(WSConn)) {
	c.onClose = hookFunc
}

func (c *ServerConn) SetOnBeforeClose(hookFunc func(WSConn)) {
	c.beforeClose = hookFunc
}

func (c *ServerConn) Handshake() {
	secretKey := ""
	if c.needSecret {
		key, err := c.GetProperty("secretKey")
		if err == nil {
			secretKey = key.(string)
		} else {
			secretKey = util.RandSeq(16)
		}
	}

	handshake := &Handshake{Key: secretKey}
	body := &RspBody{Name: HandshakeMsg, Msg: handshake}
	data, err := json.Marshal(body)
	if err != nil {
		clog.Error("handshake Marshal body error: %v", err)
		return
	}

	if secretKey != "" {
		c.SetProperty("secretKey", secretKey)
	} else {
		c.RemoveProperty("secretKey")
	}

	if zipped, err := util.Zip(data); err == nil {
		c.wsSocket.WriteMessage(websocket.BinaryMessage, zipped)
	}
}
