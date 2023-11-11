package main

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/wing8169/golang-htmx-chat-app/templates"
	"github.com/wing8169/golang-htmx-chat-app/templates/components"
)

var (
	upgrader = websocket.Upgrader{}
)

func joinChat(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	for {
		// Write
		component := components.Message("Hello, Client!")
		buffer := &bytes.Buffer{}
		component.Render(context.Background(), buffer)

		err := ws.WriteMessage(websocket.TextMessage, buffer.Bytes())
		if err != nil {
			c.Logger().Error(err)
		}
		time.Sleep(time.Second * 10)

		// Read
		// _, msg, err := ws.ReadMessage()
		// if err != nil {
		// 	c.Logger().Error(err)
		// }
		// fmt.Printf("%s\n", msg)
	}
}

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		component := templates.Index()
		return component.Render(context.Background(), c.Response().Writer)
	})
	e.GET("/ws/chat", joinChat)

	e.GET("/components", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusNotImplemented, "Not supported yet.")
		// t := c.QueryParam("type")
		// switch t {
		// case "add-todo":
		// 	component := components.AddTodoInput()
		// 	return component.Render(context.Background(), c.Response().Writer)
		// case "add-todo-btn":
		// 	component := components.AddTodoButton()
		// 	return component.Render(context.Background(), c.Response().Writer)
		// }
		// return echo.NewHTTPError(http.StatusBadRequest, "Invalid element")
	})
	e.Static("/css", "css")
	e.Static("/static", "static")
	e.Static("/fonts", "fonts")
	e.Logger.Fatal(e.Start(":3000"))
}
