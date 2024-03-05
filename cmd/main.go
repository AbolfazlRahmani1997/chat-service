package main

import (
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
	"os"
	"server/internal/ws"
	"server/router"
)

func main() {

	mongoUrl := fmt.Sprintf("mongodb://%s:27017", os.Getenv("MONGO_DB_HOST"))
	credential := options.Credential{
		Username: os.Getenv("MONGO_DB_USERNAME"),
		Password: os.Getenv("MONGO_DB_PASSWORD"),
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
	err = router.Start("0.0.0.0:8080")
	if err != nil {
		fmt.Println(err)
		return
	}

}
