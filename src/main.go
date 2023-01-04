package main

import (
	"fmt"
	"mqttserver/mqtt"
	"time"
)


func main() {
	finished := make(chan bool)

	// Wait for the server to start properly
    fmt.Println("Waiting for the MQTT broker to start...")
    time.Sleep(10 * time.Second)
	go mqtt.StartMqtt(finished)
	<- finished
}
