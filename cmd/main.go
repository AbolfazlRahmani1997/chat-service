package main

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
	"server/internal/ws"
	"server/router"
)

func main() {
	clientOptions := options.Client().ApplyURI("mongodb://127.0.0.1:27017")
	// Connect to MongoDB
	client, _ := mongo.Connect(context.TODO(), clientOptions)
	hub := ws.NewHub(client)
	wsHandler := ws.NewHandler(hub)
	go hub.Run()
	router.InitRouter(wsHandler)
	router.Start("0.0.0.0:8080")

}
