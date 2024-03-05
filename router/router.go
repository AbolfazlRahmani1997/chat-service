package router

import (
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"io/ioutil"
	"net/http"
	"server/internal/ws"
	"strconv"
	"time"
)

var r *gin.Engine

func InitRouter(wsHandler *ws.Handler) {
	r = gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "*"
		},
		MaxAge: 12 * time.Hour,
	}))
	//r.Use(Auth())
	//todo:create from rabbitmq
	r.POST("/ws/createRoom", wsHandler.CreateRoom)
	r.GET("/ws/joinRoom/:roomId", wsHandler.JoinRoom)
	r.GET("/ws/seenMessage/:roomId", wsHandler.ReadMessage)
	r.GET("/ws/getRooms/:userId", wsHandler.GetRooms)
	r.GET("/ws/getClients/:roomId", wsHandler.GetClients)
}

func Start(addr string) error {
	return r.Run(addr)
}

func Auth() gin.HandlerFunc {
	type User struct {
		Id int `json:"id"`
	}

	return func(c *gin.Context) {
		var user User
		// Set example variable
		client := &http.Client{}
		request, err := http.NewRequest("GET", "http://gateway-backend/api/user", nil)
		request.Header.Set("Authorization", c.GetHeader("Authorization"))
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := client.Do(request)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("sendRequest")
		body, _ := ioutil.ReadAll(res.Body)

		derr := json.Unmarshal(body, &user)

		if derr != nil {
			fmt.Println(derr)
		}
		c.Set("userId", strconv.Itoa(user.Id))
		c.Next()

	}

}
