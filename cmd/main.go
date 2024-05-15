package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
	"log"
	"os"
	"server/adapters"
	"server/internal/admin"
	"server/internal/ws"
	"server/router"
	"server/services"
)

func main() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Print("No .env file found")
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
	roomRepository := adapters.NewRoomRepository(client)
	adminService := services.NewRoomService(roomRepository)
	wsHandler := ws.NewHandler(hub)
	wsAdminHandler := admin.NewHandler(hub, adminService)
	go wsHandler.UpdateUserPool()
	go hub.Run()
	go hub.Manager()
	router.InitRouter(wsHandler, wsAdminHandler)
	chatUrl := fmt.Sprintf("0.0.0.0:%s", os.Getenv("CHAT_WS_PORT"))
	err = router.Start(chatUrl)
	if err != nil {
		fmt.Println(err)
		return
	}

}
