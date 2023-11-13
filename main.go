package main

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/wing8169/golang-htmx-chat-app/templates"
)

var (
	upgrader = websocket.Upgrader{}
)

func main() {
	e := echo.New()
	manager := NewManager()
	go manager.HandleClientListEventChannel(context.Background())
	e.GET("/", func(c echo.Context) error {
		component := templates.Index()
		return component.Render(context.Background(), c.Response().Writer)
	})
	e.GET("/ws/chat", manager.Handle)

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
