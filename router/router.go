package router

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"server/internal/ws"
	"strconv"
)

var r *gin.Engine

func InitRouter(wsHandler *ws.Handler) {
	gin.SetMode(gin.ReleaseMode)
	r = gin.Default()

	//r.Use(cors.New(cors.Config{
	//	AllowOrigins:     []string{"*"},
	//	AllowMethods:     []string{"GET", "POST"},
	//	AllowHeaders:     []string{"Content-Type"},
	//	ExposeHeaders:    []string{"Content-Length"},
	//	AllowCredentials: true,
	//	AllowOriginFunc: func(origin string) bool {
	//		return origin == "*"
	//	},
	//	MaxAge: 12 * time.Hour,
	//}))
	//r.Use(Auth(*wsHandler))
	//todo:create from rabbitmq
	r.POST("/chat/ws/createRoom", wsHandler.CreateRoom)
	r.GET("/chat/ws/joinRoom/:roomId", wsHandler.JoinRoom)
	r.GET("/chat/ws/seenMessage/:roomId", wsHandler.ReadMessage)
	r.GET("/chat/ws/getRooms/", wsHandler.GetRooms)
	r.GET("/chat/ws/syncRooms/", wsHandler.SyncRoom)
	r.GET("/chat/ws/getClients/:roomId", wsHandler.GetClients)
	r.POST("/chat/updateRoom/:roomId", wsHandler.UpdateRoom)
}

func Start(addr string) error {
	return r.Run(addr)
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
		var user User
		// Set example variable
		client := &http.Client{}
		getwayUrl := fmt.Sprintf("%s/api/user", "http://dev.oteacher.org/")
		request, err := http.NewRequest("GET", getwayUrl, nil)
		request.Header.Set("Authorization", c.GetHeader("Authorization"))
		if err != nil {
			return
		}
		res, err := client.Do(request)
		if err != nil {
			fmt.Println(err)
		}

		body, _ := ioutil.ReadAll(res.Body)
		derr := json.Unmarshal(body, &user)

		if derr != nil {
			fmt.Println(derr)
		}

		c.Set("userId", strconv.Itoa(user.Id))
		c.Set("Avatar", user.Avatar)
		c.Set("FirstName", user.FirstName)
		c.Set("LastName", user.LastName)
		c.Set("username", user.UserName)
		handler.UpdateUser(ws.UserDto{UserId: strconv.Itoa(user.Id), UserName: user.UserName, FirstName: user.FirstName, LastName: user.LastName, AvatarUrl: user.Avatar})
		c.Next()

	}

}
