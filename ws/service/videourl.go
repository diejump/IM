package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
)

func VideoURL(r *gin.Engine) {
	r.Static("/video", "./video")

	files, err := ioutil.ReadDir("./video")
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		fmt.Println(file.Name())
		filename := file.Name()
		r.GET("/videos/"+filename, func(c *gin.Context) {
			c.File("./video/" + filename)
		})
	}
}
