package main

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/wing8169/golang-htmx-chat-app/templates/components"
)

type Client struct {
	Conn           *websocket.Conn
	ID             string
	Chatroom       string
	Manager        *Manager
	MessageChannel chan string
}

var (
	pongWaitTime = time.Second * 10
	pingInterval = time.Second * 9
)

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		Conn:           conn,
		ID:             uuid.New().String(),
		Chatroom:       "general",
		Manager:        manager,
		MessageChannel: make(chan string),
	}
}

func (c *Client) ReadMessages(ctx echo.Context) {
	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWaitTime)); err != nil {
		ctx.Logger().Error(err)
		return
	}
	c.Conn.SetPongHandler(func(appData string) error {
		if err := c.Conn.SetReadDeadline(time.Now().Add(pongWaitTime)); err != nil {
			ctx.Logger().Error(err)
			return err
		}
		fmt.Println("pong")
		return nil
	})
	defer func() {
		c.Conn.Close()
		c.Manager.ClientListEventChannel <- &ClientListEvent{
			Client:    c,
			EventType: "REMOVE",
		}
	}()
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			ctx.Logger().Error(err)
			return
		}
		fmt.Printf("%s\n", msg)
		if err := c.Manager.WriteMessage(string(msg), "general"); err != nil {
			ctx.Logger().Error(err)
			return
		}
	}
}

func (c *Client) WriteMessage(echoContext echo.Context, ctx context.Context) {
	defer func() {
		c.Conn.Close()
		c.Manager.ClientListEventChannel <- &ClientListEvent{
			Client:    c,
			EventType: "REMOVE",
		}
	}()
	ticker := time.NewTicker(pingInterval)
	for {
		select {
		case text, ok := <-c.MessageChannel:
			if !ok {
				return
			}
			// Write
			component := components.Message(text)
			buffer := &bytes.Buffer{}
			component.Render(context.Background(), buffer)
			err := c.Conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				echoContext.Logger().Error(err)
				return
			}
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.Conn.WriteMessage(websocket.PingMessage, []byte("")); err != nil {
				echoContext.Logger().Error(err)
				return
			}
			fmt.Println("Ping")
		}
	}
}
