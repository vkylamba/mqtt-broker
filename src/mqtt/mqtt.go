package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var MQTT_HOST string
var MQTT_PORT string
var MQTT_USER string
var MQTT_PASSWORD string

var CLIENT_HEARTBEAT_TOPIC string
var CLIENT_COMMAND_TOPIC string

const CLIENT_SYSTEM_STATUS_TOPIC_TYPE string = "status"
const CLIENT_METERS_DATA_TOPIC_TYPE string = "meters-data"
const CLIENT_MODBUS_DATA_TOPIC_TYPE string = "modbus-data"
const CLIENT_CMD_RESPONSE_TOPIC_TYPE string = "cmd-resp"
const CLIENT_CMD_REQUEST_TOPIC_TYPE string = "cmd-req"
const CLIENT_UPDATE_REQ_TOPIC_TYPE string = "update-trigger"
const CLIENT_UPDATE_RESP_TOPIC_TYPE string = "update-response"
const CLIENT_HEARTBEAT_RESP_TOPIC_TYPE string = "heartbeat"
const CLIENT_COMMAND_RESP_TOPIC_TYPE string = "command"

const MEROSS_DEVICE_DATA_TOPIC_TYPE string = "publish"

const CLIENT_COUNT_TOPIC = "$SYS/broker/clients/connected"

type JsonData map[string]interface{}
type DataTopicInfoType struct {
	TopicName    string    `json:"name"`
	LastSyncTime time.Time `json:"lastSyncTime"`
}
type DeviceInfoType struct {
	DeviceName        string              `json:"name"`
	GroupName         string              `json:"group"`
	LastSyncTime      time.Time           `json:"lastSyncTime"`
	DataTopics        []DataTopicInfoType `json:"dataTopics"`
	LatestDataByTopic map[string]JsonData `json:"latestDataByTopic"`
}

type SystemInfoType struct {
	NumberOfClients int                       `json:"clients"`
	Devices         map[string]DeviceInfoType `json:"devices"`
}

var SystemInfoData SystemInfoType

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	messageTopic := msg.Topic()
	messagePayload := msg.Payload()

	fmt.Printf("Received message: %s from topic: %s\n", messagePayload, messageTopic)

	if messageTopic == CLIENT_COUNT_TOPIC {
		SystemInfoData.NumberOfClients, _ = strconv.Atoi(string(messagePayload))
	}

	// Topic is in the following format for IoT devices:
	// /Devtest/devices/Dev-test/meters-data
	topicDataList := strings.Split(messageTopic, "/")
	topicDataLength := len(topicDataList)
	if topicDataLength >= 4 {
		topicType := string(topicDataList[topicDataLength-1])
		deviceName := string(topicDataList[topicDataLength-2])
		groupName := string(topicDataList[topicDataLength-4])

		switch topicType {
		case CLIENT_SYSTEM_STATUS_TOPIC_TYPE,
			CLIENT_METERS_DATA_TOPIC_TYPE,
			CLIENT_MODBUS_DATA_TOPIC_TYPE,
			CLIENT_CMD_RESPONSE_TOPIC_TYPE,
			CLIENT_CMD_REQUEST_TOPIC_TYPE,
			CLIENT_UPDATE_REQ_TOPIC_TYPE,
			CLIENT_UPDATE_RESP_TOPIC_TYPE,
			MEROSS_DEVICE_DATA_TOPIC_TYPE,
			CLIENT_HEARTBEAT_RESP_TOPIC_TYPE,
			CLIENT_COMMAND_RESP_TOPIC_TYPE:
			device := findDevice(groupName, deviceName, topicType)
			messageData := string(messagePayload)
			var objMap = make(JsonData)
			if err := json.Unmarshal(messagePayload, &objMap); err != nil {
				if topicType != "heartbeat" && topicType != "command" {
					fmt.Printf("Invalid json data: %s.\n", messageData)
				}
				objMap["time"] = messageData
				objMap["rawMessage"] = messageData
			}
			device.LatestDataByTopic[topicType] = objMap

		default:
			fmt.Printf("Unknown topic %s.\n", topicType)
		}
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Printf("Connected to %v:%v\n", MQTT_HOST, MQTT_PORT)
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v\n", err)
}

func findDevice(groupName string, deviceName string, topicType string) *DeviceInfoType {
	dev := SystemInfoData.Devices[groupName+"/"+deviceName]
	dev.DeviceName = deviceName
	dev.GroupName = groupName
	dev.LastSyncTime = time.Now().UTC()
	updateDataTopics(&dev.DataTopics, topicType)
	dev.LatestDataByTopic = make(map[string]JsonData)
	SystemInfoData.Devices[groupName+"/"+deviceName] = dev
	return &dev
}

func updateDataTopics(dataTopics *[]DataTopicInfoType, topicType string) {
	topicFound := false
	if dataTopics == nil {
		temp := make([]DataTopicInfoType, 0)
		dataTopics = &temp
	}
	for _, v := range *dataTopics {
		if v.TopicName == topicType {
			v.LastSyncTime = time.Now().UTC()
			topicFound = true
		}
	}
	if !topicFound {
		dataTopic := DataTopicInfoType{
			TopicName: topicType,
			LastSyncTime: time.Now().UTC(),
		}
		*dataTopics = append(*dataTopics, dataTopic)
	}
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

func StartMqtt(finished chan bool, publishQueue chan map[string]string) {

	defer func() {
		finished <- true
	}()

	readEnvs()

	port, err := strconv.Atoi(MQTT_PORT)
	if err != nil {
		port = 9024
	}
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("ssl://%s:%d", MQTT_HOST, port))
	opts.SetClientID("go_mqtt_client")

	if MQTT_USER != "" {
		opts.SetUsername(MQTT_USER)
		opts.SetPassword(MQTT_PASSWORD)
	}

	var tlsConfig tls.Config
	certpool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("root_ca.crt")
	if err != nil {
		log.Fatalln(err.Error())
	}
	certpool.AppendCertsFromPEM(ca)
	tlsConfig.RootCAs = certpool
	tlsConfig.InsecureSkipVerify = true
	// tlsConfig.VerifyConnection = nil

	opts.SetTLSConfig(&tlsConfig)

	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	opts.TLSConfig.InsecureSkipVerify = true

	SystemInfoData.Devices = make(map[string]DeviceInfoType)

	subscribeAllTopics(client)
	subscribeActiveClientsTopic(client)
	go publishHeartbeat(client, finished)
	go checkAndPublishCommands(client, finished)
	go publish(client, publishQueue, finished)

	// Wait for the go routines to finish
	<-finished
	client.Disconnect(250)
}

func publishHeartbeat(client mqtt.Client, finished chan bool) {
	defer func() {
		finished <- true
	}()

	for {
		timeNow := time.Now()
		text := fmt.Sprintf("epoch_ms %v ", timeNow.UnixMilli())
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

func subscribeActiveClientsTopic(client mqtt.Client) {
	topic := CLIENT_COUNT_TOPIC
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s\n", topic)
}

func publish(client mqtt.Client, publishQueue chan map[string]string, finished chan bool) {

	defer func() {
		finished <- true
	}()

	for {
		jsonData := <-publishQueue
		fmt.Printf("Publishing to mqtt: %s, %v\n", jsonData["topicName"], jsonData["message"])
		if len(jsonData["topicName"]) > 0 {
			token := client.Publish(jsonData["topicName"], 0, false, jsonData["message"])
			token.Wait()
		}
		time.Sleep(1 * time.Second)
	}
}
