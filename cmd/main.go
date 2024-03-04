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

	mongoUrl := fmt.Sprintf("mongodb://%s:27017", os.Getenv("MongoDbUrl"))
	clientOptions := options.Client().ApplyURI(mongoUrl)
	// Connect to MongoDB
	client, _ := mongo.Connect(context.TODO(), clientOptions)
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
