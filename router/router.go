package router

import (
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"server/internal/admin"
	"server/internal/ws"
	"strconv"
	"time"
)

var r *gin.Engine

func InitRouter(wsHandler *ws.Handler, adminHandler admin.Handler) {
	gin.SetMode(gin.ReleaseMode)
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
	//r.Use(Auth(*wsHandler))
	//todo:create from rabbitmq
	r.POST("/chat/ws/createRoom", wsHandler.CreateRoom)
	r.GET("/chat/ws/joinRoom/:roomId", wsHandler.JoinRoom)
	r.GET("/chat/ws/seenMessage/:roomId", wsHandler.ReadMessage)
	r.GET("/chat/ws/getRooms/", wsHandler.GetRooms)
	r.GET("/chat/ws/syncRooms/", wsHandler.SyncRoom)
	r.GET("/chat/ws/getClients/:roomId", wsHandler.GetClients)
	adminRoute := r.Group("api/admin")
	adminRoute.GET("/:roomId", adminHandler.FetchRooms)
	adminRoute.GET("/chat/user", adminHandler.FetchRooms)
	adminRoute.PUT("/chat/user/:roomId", adminHandler.EditRoom)
	//r.GET("api/admin/chat/user", func(context *gin.Context) {
	//	data := context.GetHeader("Authorization")
	//	s := strings.Split(data, ".")
	//	type user struct {
	//		UserId int `json:"user_id"`
	//	}
	//	var userTest user
	//	text := fmt.Sprintf(s[1])
	//	decodeString, err := base64.URLEncoding.DecodeString(text)
	//	if err != nil {
	//	}
	//	dat := decodeString[:len(decodeString)-1]
	//	t := string(dat) + "}"
	//
	//	err = json.Unmarshal([]byte(t), &userTest)
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//	context.JSON(200, userTest)
	//})
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
			fmt.Println(err)
			return
		}
		res, err := client.Do(request)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("sendRequest")
		fmt.Println(res.Body)
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
