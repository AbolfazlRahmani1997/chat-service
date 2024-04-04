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

	mongoUrl := fmt.Sprintf("mongodb://%s:%s/", os.Getenv("MONGO_DB_HOST"), os.Getenv("MONGO_DB_PORT"))
	credential := options.Credential{
		Username: os.Getenv("MONGO_DB_USERNAME"),
		Password: os.Getenv("MONGO_DB_PASSWORD"),
	}
	//mongoUrl := fmt.Sprintf("mongodb://%s:%s/", "127.0.0.1", "27017")

	//serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().ApplyURI(mongoUrl).SetAuth(credential)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		fmt.Println(err)
	}
	hub := ws.NewHub(client)

	wsHandler := ws.NewHandler(hub)
	go wsHandler.UpdateUserPool()
	go hub.Run()
	go hub.Manager()
	router.InitRouter(wsHandler)
	err = router.Start("0.0.0.0:8080")
	if err != nil {
		fmt.Println(err)
		return
	}

}
