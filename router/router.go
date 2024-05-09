package router

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"server/internal/ws"
	"strings"
	"time"
)

var r *gin.Engine

func InitRouter(wsHandler *ws.Handler) {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	r = gin.Default()
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST"},
		AllowHeaders: []string{"*"},

		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "*"
		},
		MaxAge: 12 * time.Hour,
	}))

	//todo:create from rabbitmq
	r.POST("/chat/ws/createRoom", wsHandler.CreateRoom)
	r.GET("/chat/ws/joinRoom/:roomId", wsHandler.JoinRoom)
	r.GET("/chat/ws/seenMessage/:roomId", wsHandler.ReadMessage)
	r.GET("/chat/ws/getRooms/", PGPToken(), Auth(*wsHandler), wsHandler.GetRooms).Use(PGPToken())
	r.GET("/chat/ws/syncRooms/", wsHandler.SyncRoom)
	r.GET("/chat/ws/getClients/:roomId", wsHandler.GetClients)
	r.GET("/chat/room/pin/:roomId", wsHandler.UpdatePin)
	r.GET("/chat/room/notification/:roomId", wsHandler.UpdateNotification)
}

func Start(addr string) error {
	return r.Run(addr)
}

func PGPToken() gin.HandlerFunc {

	return func(c *gin.Context) {
		token := c.Query("token")
		if token != "" {
			data := strings.Split(token, ".")
			if len(data) != 3 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token is required"})
				return

			}
			c.Next()
		}

		c.Next()
	}
}

func Auth(handler ws.Handler) gin.HandlerFunc {
	type User struct {
		Id        int    `json:"id"`
		Avatar    string `json:"avatar"`
		FirstName string `json:"firstname"`
		LastName  string `json:"lastname"`
		UserName  string `json:"username"`
	}

	return func(c *gin.Context) {
		token := c.Query("token")
		user, ok := handler.UserHandler[token]
		fmt.Print(ok)
		if ok {
			if user.LastStatusCode == 500 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token is required"})
			}
		}
		c.Next()

	}

}
