package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	echojwt "github.com/labstack/echo-jwt"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/wing8169/golang-htmx-chat-app/services"
	"github.com/wing8169/golang-htmx-chat-app/templates"
)

var (
	upgrader = websocket.Upgrader{}
)

func main() {
	db, err := sql.Open("sqlite3", "./db/chat.db")
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	sqlStmt := `
	create table if not exists user (id text not null primary key, username varchar(255), password varchar(255));
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	userService := &services.UserService{
		DB: db,
	}

	e := echo.New()
	manager := NewManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go manager.HandleClientListEventChannel(ctx)
	unguardedRoutes := e.Group("/")
	unguardedRoutes.Use(services.GuestMiddleware)
	unguardedRoutes.GET("", func(c echo.Context) error {
		component := templates.Index()
		return component.Render(ctx, c.Response().Writer)
	})
	unguardedRoutes.GET("register", func(c echo.Context) error {
		component := templates.Register()
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
		loggedInUser, err := userService.LoginUser(username, password)
		if err != nil {
			fmt.Println(err)
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid login.")
		}

		// Assign JWT tokens
		err = services.GenerateTokensAndSetCookies(loggedInUser, c)

		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Token failed to be generated.")
		}

		return c.Redirect(http.StatusMovedPermanently, "/chat")
	})

	e.POST("/register", func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")
		confirmPassword := c.FormValue("confirmPassword")

		// password validation
		if password != confirmPassword {
			return echo.NewHTTPError(http.StatusBadRequest, "Password is not the same as confirm password.")
		}

		// user validation
		users, err := userService.GetUsers(username)
		if err != nil || len(users) > 0 {
			fmt.Println(err)
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid credentials.")
		}

		// create a new user
		newUser, err := userService.CreateUser(username, password)
		if err != nil {
			fmt.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Invalid login.")
		}

		// Assign JWT tokens
		err = services.GenerateTokensAndSetCookies(newUser, c)

		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Token failed to be generated.")
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
