package main

import (
	"fmt"
	"mqttserver/api"
	"mqttserver/mqtt"
	"time"
)

func main() {
	finished := make(chan bool)
	publishQueue := make(chan map[string]string)

	// Wait for the server to start properly
    fmt.Println("Waiting for the MQTT broker to start...")
    time.Sleep(10 * time.Second)
	go mqtt.StartMqtt(finished, publishQueue)
	go api.StartServer(finished, publishQueue)
	<- finished
	<- finished
}
