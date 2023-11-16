package main

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	echojwt "github.com/labstack/echo-jwt"
	"github.com/labstack/echo/v4"
	"github.com/wing8169/golang-htmx-chat-app/dto"
	"github.com/wing8169/golang-htmx-chat-app/services"
	"github.com/wing8169/golang-htmx-chat-app/templates"
	"golang.org/x/crypto/bcrypt"
)

var (
	upgrader = websocket.Upgrader{}
)

func main() {
	e := echo.New()
	manager := NewManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go manager.HandleClientListEventChannel(ctx)
	e.GET("/", func(c echo.Context) error {
		component := templates.Index()
		return component.Render(ctx, c.Response().Writer)
	})
	guardedRoutes := e.Group("/chat")
	guardedRoutes.Use(services.TokenRefresherMiddleware)
	guardedRoutes.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey:   []byte(services.JwtSecretKey),
		TokenLookup:  "cookie:access-token", // "<source>:<name>"
		ErrorHandler: services.JWTErrorChecker,
	}))
	guardedRoutes.GET("", func(c echo.Context) error {
		component := templates.Chat()
		return component.Render(ctx, c.Response().Writer)
	})
	e.GET("/ws/chat", func(c echo.Context) error {
		return manager.Handle(c, ctx)
	})

	e.POST("/login", func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")

		var loggedInUser *dto.UserDto

		users := services.GetUsers()

		for _, user := range users {
			if user.Username != username {
				continue
			}
			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err == nil {
				loggedInUser = user
				break
			}
		}

		if loggedInUser == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid login.")
		}

		// Assign JWT tokens
		err := services.GenerateTokensAndSetCookies(loggedInUser, c)

		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Token is incorrect")
		}

		return c.Redirect(http.StatusMovedPermanently, "/chat")
	})

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
