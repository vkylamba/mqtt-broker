package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var MQTT_HOST string
var MQTT_PORT string
var MQTT_USER string
var MQTT_PASSWORD string

var CLIENT_HEARTBEAT_TOPIC string
var CLIENT_COMMAND_TOPIC string


var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
    fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
    fmt.Printf("Connected to %v:%v\n", MQTT_HOST, MQTT_PORT)
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
    fmt.Printf("Connect lost: %v\n", err)
}

func readEnvs() {
    MQTT_HOST = os.Getenv("MQTT_HOST")
    MQTT_PORT = os.Getenv("MQTT_PORT")
    MQTT_USER = os.Getenv("MQTT_USER")
    MQTT_PASSWORD = os.Getenv("MQTT_PASSWORD")

    CLIENT_HEARTBEAT_TOPIC = os.Getenv("CLIENT_HEARTBEAT_TOPIC")
    CLIENT_COMMAND_TOPIC = os.Getenv("CLIENT_COMMAND_TOPIC")

    if len(MQTT_HOST) < 5 {
        MQTT_HOST = "0.0.0.0"   
    }

    if len(MQTT_PORT) < 2 {
        MQTT_PORT = "9024"   
    }

    if len(CLIENT_HEARTBEAT_TOPIC) < 10 {
        CLIENT_HEARTBEAT_TOPIC = "/iotaapsys/services/heartbeat"   
    }

    if len(CLIENT_COMMAND_TOPIC) < 10 {
        CLIENT_COMMAND_TOPIC = "/iotaapsys/services/command"   
    }

}

func main() {
    readEnvs()

    // Wait for the server to start properly
    fmt.Println("Waiting for the MQTT broker to start...")
    time.Sleep(10 * time.Second)

    finished := make(chan bool)

    port, err := strconv.Atoi(MQTT_PORT)
    if err != nil {
        port = 9024
    }
    opts := mqtt.NewClientOptions()
    opts.AddBroker(fmt.Sprintf("tcp://%s:%d", MQTT_HOST, port))
    opts.SetClientID("go_mqtt_client")

    if MQTT_USER != "" {
        opts.SetUsername(MQTT_USER)
        opts.SetPassword(MQTT_PASSWORD)
    }

    opts.SetDefaultPublishHandler(messagePubHandler)
    opts.OnConnect = connectHandler
    opts.OnConnectionLost = connectLostHandler
    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }

    subscribeAllTopics(client)
    go publishHeartbeat(client, finished)
    go checkAndPublishCommands(client, finished)

    // Wait for the go routines to finish
    <- finished
    client.Disconnect(250)
}

func publishHeartbeat(client mqtt.Client, finished chan bool) {
    defer func() {
        finished <- true
    }()

    for {
        timeNow := time.Now()
        text := fmt.Sprintf("epoch_ms %v", timeNow.UnixMilli())
        token := client.Publish(CLIENT_HEARTBEAT_TOPIC, 0, false, text)
        token.Wait()
        time.Sleep(60 * time.Second)
    }
}

func checkAndPublishCommands(client mqtt.Client, finished chan bool) {
    defer func() {
        finished <- true
    }()

    for {
        timeNow := time.Now()
        text := fmt.Sprintf("command %v", timeNow.UnixMilli())
        token := client.Publish(CLIENT_COMMAND_TOPIC, 0, false, text)
        token.Wait()
        time.Sleep(10 * time.Second)
    }
}

func subscribeAllTopics(client mqtt.Client) {
    topic := "#"
    token := client.Subscribe(topic, 1, nil)
    token.Wait()
    fmt.Printf("Subscribed to topic: %s\n", topic)
}
