package mqtt

import (
	"fmt"
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
const CLIENT_UPDATE_RESP_TOPIC_TYPE string = "update-response"

const CLIENT_COUNT_TOPIC = "$SYS/broker/clients/connected"


type DeviceInfoType struct {
	DeviceName string `json:"name"`
    GroupName string `json:"group"`
    LastSyncTime time.Time `json:"lastSyncTime"`
    Topics []string `json:"topics"`
    LatestDataByTopic map[string]string `json:"latestDataByTopic"`
}

type SystemInfoType struct {
	NumberOfClients int `json:"clients"`
    Devices map[string]DeviceInfoType `json:"devices"`
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
        groupName := string(topicDataList[topicDataLength-3])

        switch topicType {
        case CLIENT_SYSTEM_STATUS_TOPIC_TYPE,
             CLIENT_METERS_DATA_TOPIC_TYPE,
             CLIENT_MODBUS_DATA_TOPIC_TYPE,
             CLIENT_UPDATE_RESP_TOPIC_TYPE:
            device := findDevice(deviceName, groupName)
            if !contains(device.Topics, topicType) {
                if len(device.Topics) == 0 {
                    device.Topics = make([]string, 0)
                    device.LatestDataByTopic = make(map[string]string, 0)
                }
                device.Topics = append(device.Topics, topicType)
                device.LatestDataByTopic[topicType] = string(messagePayload)
            }
            device.LastSyncTime = time.Now()
        case "heartbeat", "command":
            
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

func findDevice(groupName string, deviceName string) DeviceInfoType {
    dev := SystemInfoData.Devices[groupName + "/" + deviceName]
    dev.DeviceName = deviceName
    dev.GroupName = groupName
    SystemInfoData.Devices[groupName + "/" + deviceName] = dev
    return dev
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
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

func StartMqtt(finished chan bool) {

    defer func() {
        finished <- true
    }()

    readEnvs()

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

    SystemInfoData.Devices = make(map[string]DeviceInfoType);

    subscribeAllTopics(client)
    subscribeActiveClientsTopic(client)
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

func subscribeActiveClientsTopic(client mqtt.Client) {
    topic := CLIENT_COUNT_TOPIC
    token := client.Subscribe(topic, 1, nil)
    token.Wait()
    fmt.Printf("Subscribed to topic: %s\n", topic)
}
