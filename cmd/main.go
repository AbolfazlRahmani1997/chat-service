package main

import (
	"fmt"
	"github.com/Valgard/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
	"os"
	"server/internal/ws"
	"server/router"
)

func main() {
	dotenv := godotenv.New()
	if err := dotenv.Load("../.env"); err != nil {

	}
	mongoUrl := fmt.Sprintf("mongodb://%s:%s/", os.Getenv("MONGO_DB_HOST"), os.Getenv("MONGO_DB_PORT"))
	credential := options.Credential{
		Username: os.Getenv("MONGO_DB_USERNAME"),
		Password: os.Getenv("MONGO_DB_PASSWORD"),
	}
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
	chatUrl := fmt.Sprintf("0.0.0.0:%s", os.Getenv("CHAT_WS_PORT"))
	err = router.Start(chatUrl)
	if err != nil {
		fmt.Println(err)
		return
	}

}
