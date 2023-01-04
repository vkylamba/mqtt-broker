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

func StartServer(finished chan bool) {
	readEnvs()
    router := gin.Default()

	authorized := router.Group("/", gin.BasicAuth(gin.Accounts{
        BASIC_AUTH_USER: BASIC_AUTH_PASSWORD,
    }))

    authorized.GET("/clients", getClients)

    router.Run(fmt.Sprintf("%v:%v", API_HOST, API_PORT))
	finished <- true
}
