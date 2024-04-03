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

type UserDto struct {
	UserId    string `json:"Id,omitempty"`
	UserName  string `json:"Username"`
	FirstName string `json:"Firstname"`
	LastName  string `json:"LastName"`
	AvatarUrl string `json:"AvatarUrl"`
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

	return model

}

func (receiver UserRepository) UpdateUser(userid string, user UserDto) {

	userId, _ := strconv.Atoi(userid)
	update := bson.D{{"$set", bson.D{{"firstname", user.FirstName}, {"lastname", user.LastName}, {"username", user.UserName}, {"avatar_url", user.AvatarUrl}}}}
	fmt.Println(update)
	filter := bson.M{"user_id": userId}

	one, err := receiver.database.Collection("users").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		fmt.Println("userUpdate hase eroo")
		fmt.Println(err)
	}
	fmt.Println("updated")
	fmt.Println(one.ModifiedCount, one.UpsertedCount)
}
