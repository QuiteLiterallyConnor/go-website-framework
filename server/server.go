package server

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
	"go-website-framework/auth"
)

type Server struct {
	Router     *gin.Engine
	Auth       *auth.Auth
	wsCallback func(sessionUUID string, msg string) error
}

func NewServer() *Server {
	router := gin.Default()

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	router.LoadHTMLGlob("templates/*")

	s := &Server{Router: router}

	router.GET("/", func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")
		loggedIn := user != nil
		c.HTML(http.StatusOK, "index.html", gin.H{
			"IsLoggedIn":  loggedIn,
			"CurrentPage": "index",
			"Username":    user,
		})
	})

	router.GET("/ws", s.handleWebsocket)

	return s
}

func (s *Server) AuthRoutes(port string) {
	s.Router.GET("/login", s.Auth.LoginPage)
	s.Router.POST("/login", s.Auth.Login)
	s.Router.GET("/register", s.Auth.RegisterPage)
	s.Router.POST("/register", s.Auth.Register)
	s.Router.GET("/logout", s.Auth.Logout)
	s.Router.GET("/profile", s.Auth.ProfilePage)
}

func (s *Server) OnReceiveWebsocket(callback func(sessionUUID string, msg string) error) {
	s.wsCallback = callback
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (s *Server) handleWebsocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Failed to upgrade connection:", err)
		return
	}
	sessionUUID := c.Query("session_uuid")
	if sessionUUID == "" {
		sessionUUID = uuid.New().String()
	}
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}
		if s.wsCallback != nil {
			if err := s.wsCallback(sessionUUID, string(msg)); err != nil {
				fmt.Println("WebSocket callback error:", err)
			}
		}
	}
	conn.Close()
}

func (s *Server) Serve(port string) {
	s.Router.Run(port)
}
