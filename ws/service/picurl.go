package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
)

func PicURL(rg *gin.RouterGroup, r *gin.Engine) {
	r.Static("/images", "./images")

	files, err := ioutil.ReadDir("./images")
	if err != nil {
		panic(err)
	}

	// 为每个图片文件创建路由
	for _, file := range files {
		fmt.Println(file.Name())
		r.GET("/image/"+file.Name(), func(c *gin.Context) {
			c.File("./images/" + file.Name())
		})
	}
}
