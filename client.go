package main

import (
	"bytes"
	"context"
	"fmt"

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
		c.MessageChannel <- string(msg)
	}
}

func (c *Client) WriteMessage(ctx echo.Context) {
	defer func() {
		c.Conn.Close()
		c.Manager.ClientListEventChannel <- &ClientListEvent{
			Client:    c,
			EventType: "REMOVE",
		}
	}()
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

			for _, client := range c.Manager.ClientList {
				err := client.Conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
				if err != nil {
					ctx.Logger().Error(err)
				}
			}
		case <-context.Background().Done():
			return
		}
	}
}
