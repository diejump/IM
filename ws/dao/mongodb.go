package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"ws/model"
)

var friendCollection, MessageCollection *mongo.Collection

func InitMongoDB(url string) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(url))

	if err != nil {
		log.Println("MongoDB连接失败")
		panic(err)
		return
	}

	log.Println("MongoDB连接成功")
	friendCollection = client.Database("im").Collection("friend")
	MessageCollection = client.Database("im").Collection("sentmessage")
}

func RegisterFriendQueue(account string) {
	friendCollection.InsertOne(context.TODO(), bson.M{
		"account": account,
		"friend":  bson.M{},
	})
}

func SaveMessage(ChatMessage model.DBMessage) {
	MessageCollection.InsertOne(context.TODO(), ChatMessage)
}

func SendOfflineMessage(RecipientAccount string) [][]byte {
	filter := bson.M{
		"$and": []bson.M{
			{"message.recipientaccount": RecipientAccount},
			{"status": 0},
		},
	}

	update := bson.M{
		"$set": bson.M{
			"status": 1, //将status更新为1，已读
		},
	}

	projection := bson.M{
		"_id":     0, //不返回id
		"message": 1, //需要返回message字段
	}

	cursor, err1 := MessageCollection.Find(context.Background(), filter, options.Find().SetProjection(projection)) //查询
	if err1 != nil {
		log.Fatal(err1)
	}

	var results []bson.M
	if err := cursor.All(context.Background(), &results); err != nil {
		log.Fatal(err)
	}

	msg := make([][]byte, 0)
	for _, result := range results {
		jsondata, err4 := json.Marshal(result["message"])
		if err4 != nil {
			fmt.Println("转json失败", err4)
		}
		msg = append(msg, jsondata)
	}

	_, err2 := MessageCollection.UpdateMany(context.Background(), filter, update) //更新数据库中的数据
	if err2 != nil {
		log.Fatal(err2)
	}

	defer cursor.Close(context.Background())
	return msg //返回数据
}
