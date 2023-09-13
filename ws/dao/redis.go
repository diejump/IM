package dao

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
)

var client *redis.Client

func InitRedis(host, port string) {
	client = redis.NewClient(&redis.Options{
		Addr: host + ":" + port, // Redis服务器地址
		DB:   0,                 // 使用默认数据库
	})
}

func RedisAddFriend(myAccount any, friendAccount string) {
	err := client.SAdd(context.Background(), fmt.Sprintf("friends:%s", myAccount), friendAccount).Err()
	err2 := client.SAdd(context.Background(), fmt.Sprintf("friends:%s", friendAccount), myAccount).Err()
	if err != nil || err2 != nil {
		log.Fatal(err)
	}

	err3 := client.Save(context.Background()).Err()
	if err3 != nil {
		log.Fatal(err3)
	}

	log.Println("添加好友成功")

}

func IsFriend(myAccount string, friendAccount string) bool { //判定用户和另一个用户是否为好友

	isFriend, err := client.SIsMember(context.Background(), fmt.Sprintf("friends:%s", myAccount), friendAccount).Result()
	if err != nil {
		log.Fatal(err)
		return false
	}

	if isFriend {
		fmt.Printf("%s 和 %s 是好友.\n", myAccount, friendAccount)
		return true
	} else {
		fmt.Printf("%s 和 %s 不是好友.\n", myAccount, friendAccount)
		return false
	}
}

func RedisDeleteFriend(friendAccount string, myAccount any) {
	sremCmd := client.SRem(context.Background(), fmt.Sprintf("friends:%s", myAccount), friendAccount)
	client.SRem(context.Background(), fmt.Sprintf("friends:%s", friendAccount), myAccount)
	if sremCmd.Err() != nil {
		log.Fatal(sremCmd.Err())
	}
	log.Println("删除好友成功")
}

func RedisClose() {
	client.Close()
}
