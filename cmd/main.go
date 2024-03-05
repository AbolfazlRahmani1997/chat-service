package main

import (
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
	"server/internal/ws"
	"server/router"
)

func main() {

	mongoUrl := fmt.Sprintf("mongodb://10.0.0.15:27017/")
	credential := options.Credential{
		Username: "amir",
		Password: "d55t1kq6tg4p1ca2",
	}
	fmt.Println(mongoUrl)
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().ApplyURI(mongoUrl).SetAuth(credential).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		fmt.Println(err)
	}
	hub := ws.NewHub(client)
	wsHandler := ws.NewHandler(hub)
	go hub.Run()
	router.InitRouter(wsHandler)
	err = router.Start("0.0.0.0:8080/chat")
	if err != nil {
		fmt.Println(err)
		return
	}

}
