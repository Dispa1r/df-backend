package main

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"log"
)

// todo list:
// redis 持久化
// emoj
var userTwitterMap map[string]string
var planetEmojiMap map[string]planetMessage
var userTwitterCache map[string]string
var emailList []string

func init() {
	userTwitterMap = make(map[string]string)
	planetEmojiMap = make(map[string]planetMessage)
	userTwitterCache = make(map[string]string)
	// init config
	viper.SetConfigFile("./config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Println("read config failed ", err)
	}
}

func main() {

	r := gin.Default()
	r.Use(Cors())

	r.GET("/twitter/all-twitters", func(c *gin.Context) {
		c.JSON(200, userTwitterMap)
	})
	r.POST("/email/interested", emailHandler)
	r.POST("/add-message", emojiHandler)
	r.POST("/messages", messageHandler)
	r.POST("/twitter/verify-twitter", twitterVerifyHandler)
	r.POST("/twitter/disconnect", disconnectHandler)

	r.Run(":3000") // 监听并在 0.0.0.0:8080 上启动服务
	return
}
