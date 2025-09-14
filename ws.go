package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
)

type subscriber struct {
	msgs chan []byte
}

// this is run as a go routine
// without context for now...
func (v *View) servicePause() {
	// half second timer loop for now

	for {
		v.ws_mtx.Lock()
		wait := len(v.waiters) > 0
		v.ws_mtx.Unlock()
		if wait {
			v.sendMsgAll([]byte(`{"type":"wait"}`))
		} else {
			v.sendMsgAll([]byte(`{"type":"proceed"}`))
		}
		time.Sleep(time.Millisecond * 500)
	}
}

// Handler for new websocket connection requests
func (v *View) Subscribe(w http.ResponseWriter, r *http.Request) {
	var c *websocket.Conn
	fmt.Println("*** We have a subscriber.")
	s := &subscriber{msgs: make(chan []byte, 16)}
	v.addSub(s)
	defer v.delSub(s)

	fmt.Println("    Gonna try and accept it.")
	c, err := websocket.Accept(w, r, nil)
	fmt.Println("    And now check for error.")
	if err != nil {
		fmt.Println("Subscribe error: ", err.Error())
		return
	}

	fmt.Println("*** connection accepted")
	defer c.CloseNow()
	ctx := c.CloseRead(context.Background())
	// generate a random id
	id, err := rand.Int(rand.Reader, big.NewInt(0x7FFFFFFFF))
	if err != nil {
		fmt.Println("Subscribe id gen err: ", err.Error())
		return
	}
	idkey := fmt.Sprintf("%08X", id)
	initmsg := fmt.Sprintf(`{"type":"sub","id":"%s"}`, idkey)
	fmt.Println("*** writing: ", initmsg)

	err = v.writeTimeout(ctx, time.Second*1, c, []byte(initmsg))
	if err != nil {
		fmt.Println("ws.Write err: ", err.Error())
		return
	}

	for {
		select {
		case msg := <-s.msgs:
			err = v.writeTimeout(ctx, time.Second*5, c, msg)
			if err != nil {
				fmt.Println("ws.Write err: ", err.Error())
				return
			}
		case <-ctx.Done():
			// TODO we know that the context was canceled
			if ctx.Err() != nil {
				fmt.Println("Context Err: ", ctx.Err().Error())
			}
			return
		}
	}
}

// Handler for publishing messages (should be POST method)
func (v *View) UpdateGigPause(w http.ResponseWriter, r *http.Request) {
	body := http.MaxBytesReader(w, r.Body, 512)
	raw, err := io.ReadAll(body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		return
	}
	// v.sendMsgAll(msg)
	msg := string(raw)
	fmt.Printf("%v\n", msg)
	parts := strings.Split(msg, "=")
	if len(parts) != 2 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	switch parts[1] {
	case "wait":
		// TODO rethink this (may have to use "session" that is somehow correlated to subscriber...)
		v.ws_mtx.Lock()
		v.waiters[parts[0]] = struct{}{}
		v.ws_mtx.Unlock()
	case "ready":
		v.ws_mtx.Lock()
		delete(v.waiters, parts[0])
		v.ws_mtx.Unlock()
	default:
	}
	w.WriteHeader(http.StatusAccepted)
}

func (v *View) sendMsgAll(msg []byte) {
	v.ws_mtx.Lock()
	defer v.ws_mtx.Unlock()

	for s := range v.subs {
		s.msgs <- msg
	}
}

func (v *View) addSub(s *subscriber) {
	v.ws_mtx.Lock()
	v.subs[s] = struct{}{}
	v.ws_mtx.Unlock()
}

func (v *View) delSub(s *subscriber) {
	v.ws_mtx.Lock()
	delete(v.subs, s)
	v.ws_mtx.Unlock()
}

func (v *View) writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.Write(ctx, websocket.MessageText, msg)
}
