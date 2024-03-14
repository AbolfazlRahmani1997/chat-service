package ws

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
	"strconv"
)

type UserRepository struct {
	database *mongo.Database
}

func NewUserRepository(client *mongo.Database) UserRepository {
	return UserRepository{database: client}
}

func (receiver UserRepository) getUser(userid string) Member {

	var model Member
	userId, err := strconv.Atoi(userid)
	filter := bson.M{"user_id": userId}
	one := receiver.database.Collection("users").FindOne(context.TODO(), filter)
	err = one.Decode(&model)
	if err != nil {
		fmt.Println(err)
		return model
	}
	fmt.Print(model)
	return model

}
