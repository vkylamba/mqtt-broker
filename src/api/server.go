package api

import (
	"fmt"
	"mqttserver/mqtt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var API_HOST string
var API_PORT string
var BASIC_AUTH_USER string
var BASIC_AUTH_PASSWORD string

type MessageData struct {
	TopicName string `json:"topic"`
	Message   string   `json:"message"`
}

func readEnvs() {
	API_HOST = os.Getenv("API_HOST")
	API_PORT = os.Getenv("API_PORT")
	BASIC_AUTH_USER = os.Getenv("BASIC_AUTH_USER")
	BASIC_AUTH_PASSWORD = os.Getenv("BASIC_AUTH_PASSWORD")

	if API_HOST == "" {
		API_HOST = "0.0.0.0"
	}

	if API_PORT == "" {
		API_PORT = "8000"
	}

	if BASIC_AUTH_USER == "" {
		BASIC_AUTH_USER = "admin"
	}

	if BASIC_AUTH_PASSWORD == "" {
		BASIC_AUTH_PASSWORD = "admin"
	}
}

// getClients responds with the list of all clients as JSON.
func getClients(c *gin.Context) {
    c.IndentedJSON(http.StatusOK, mqtt.SystemInfoData)
}

// publishMqtt publish to the provided mqtt topic with given data
func publishMqtt(publishQueue chan map[string]string) func(c *gin.Context) {
	return func (c *gin.Context) {
		var messageData MessageData
    	c.BindJSON(&messageData)
		mqttData := make(map[string]string)
		mqttData["topicName"] = messageData.TopicName
		mqttData["message"] = messageData.Message
		publishQueue <- mqttData
		c.String(http.StatusOK, messageData.TopicName)
	}
}

func StartServer(finished chan bool, publishQueue chan map[string]string) {
	readEnvs()
    router := gin.Default()

	authorized := router.Group("/", gin.BasicAuth(gin.Accounts{
        BASIC_AUTH_USER: BASIC_AUTH_PASSWORD,
    }))

    authorized.GET("/clients", getClients)
	authorized.POST("/publish/", publishMqtt(publishQueue))

	authorized.StaticFile("/dashboard-config", "./static/dashboard-config/config.json")
	authorized.StaticFS("/lib", http.Dir("./static/lib"))
	authorized.StaticFS("/plugins", http.Dir("./static/plugins"))
	authorized.StaticFS("/img", http.Dir("./static/img"))
	authorized.StaticFile("/", "./static/index.html")

    router.Run(fmt.Sprintf("%v:%v", API_HOST, API_PORT))
	finished <- true
}
