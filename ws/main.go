package main

import (
	"ws/dao"
	"ws/router"
)

func main() {
	dao.InitDB()
	router.InitRouter()
}
