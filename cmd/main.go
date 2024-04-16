package main

import (
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
	"server/adapters"
	"server/internal/admin"
	"server/internal/ws"
	"server/router"
	"server/services"
)

func main() {

	//mongoUrl := fmt.Sprintf("mongodb://%s:%s/", os.Getenv("MONGO_DB_HOST"), os.Getenv("MONGO_DB_PORT"))

	mongoUrl := fmt.Sprintf("mongodb://%s:%s/", "127.0.0.1", "27017")
	credential := options.Credential{
		Username: "root",
		Password: "root",
	}
	//serverAPI := options.ServerAPI(options.ServerAPIVersion1)
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
	err = router.Start("0.0.0.0:8088")
	if err != nil {
		fmt.Println(err)
		return
	}

}
