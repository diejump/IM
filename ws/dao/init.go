package dao

import (
	"fmt"
	"github.com/spf13/viper"
	"ws/service"
)

func InitDB() {
	viper.SetConfigName("config")
	viper.AddConfigPath("./conf")

	// 读取配置文件
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("无法读取配置文件: %s", err))
	}

	mysqlUsername := viper.GetString("database.mysql.username")
	mysqlHost := viper.GetString("database.mysql.host")
	mysqlPort := viper.GetString("database.mysql.port")
	mysqlPW := viper.GetString("database.mysql.password")

	mongodbURL := viper.GetString("database.mongodb.url")

	redisHost := viper.GetString("cache.redis.host")
	redisPort := viper.GetString("cache.redis.port")

	mqHost := viper.GetString("mq.rabbit.host")
	mqPort := viper.GetString("mq.rabbit.port")

	Initmysql(mysqlUsername, mysqlPW, mysqlHost, mysqlPort)
	InitRedis(redisHost, redisPort)
	InitMongoDB(mongodbURL)
	service.InitRabbitMQ(mqHost, mqPort)
}
