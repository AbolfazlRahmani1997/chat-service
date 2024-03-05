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

	mongoUrl := fmt.Sprintf("mongodb://127.0.0.1:27017")
	//credential := options.Credential{
	//	Username: "",
	//	Password: "",
	//}
	option := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().ApplyURI(mongoUrl).SetServerAPIOptions(option)
	client, _ := mongo.Connect(context.Background(), clientOptions)
	hub := ws.NewHub(client)
	wsHandler := ws.NewHandler(hub)
	go hub.Run()
	router.InitRouter(wsHandler)
	err := router.Start("0.0.0.0:8080")
	if err != nil {
		fmt.Println(err)
		return
	}

}
