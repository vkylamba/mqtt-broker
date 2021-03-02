import logging
import os

import dotenv
import paho.mqtt.client as mqtt
from retry import retry
from ipgetter2 import ipgetter1 as ipgetter

dotenv.load_dotenv()

MQTT_HOST = os.getenv("MQTT_HOST", "168.119.153.236")
MQTT_PORT = int(os.getenv("MQTT_PORT", "9024"))
TEST_CLIENT_NAME = os.getenv("TEST_CLIENT_NAME", "test-client")
TEST_CLIENT_IP = os.getenv("TEST_CLIENT_IP", "test-client-ip")
CLIENT_RESPONSE_TOPIC = os.getenv("CLIENT_RESPONSE_TOPIC",  "test-client-resp-topic")
CLIENT_COMMAND_TOPIC = os.getenv("CLIENT_COMMAND_TOPIC", "test-client-command-topic")
CLIENT_CONNECTED_TOPIC = os.getenv("CLIENT_CONNECTED_TOPIC", "test-client-connected-topic")

logger = logging.getLogger(__name__)
logger.info = print

# The callback for when the client receives a CONNACK response from the server.


def on_connect(client, userdata, flags, rc):
    logger.info("Connected with result code %s" % str(rc))
    # Subscribing in on_connect() means that if we lose the connection and
    # reconnect then subscriptions will be renewed.
    # client.subscribe("$SYS/#")
    client.subscribe(CLIENT_COMMAND_TOPIC)
    ip_address = ipgetter.myip()
    message = "{}, ip_address: {}".format(TEST_CLIENT_NAME, ip_address)
    client.publish(CLIENT_CONNECTED_TOPIC, message)

# The callback for when a PUBLISH message is received from the server.


def on_message(client, userdata, msg):
    logger.info(msg.topic + " " + str(msg.payload))
    message = "{}".format(TEST_CLIENT_NAME)
    client.publish(CLIENT_RESPONSE_TOPIC, message)


@retry(delay=10)
def run():
    client = mqtt.Client()
    client.on_connect = on_connect
    client.on_message = on_message

    client.connect(MQTT_HOST, MQTT_PORT, 60)

    # Blocking call that processes network traffic, dispatches callbacks and
    # handles reconnecting.
    # Other loop*() functions are available that give a threaded interface and a
    # manual interface.
    client.loop_forever()


if __name__ == '__main__':
    run()
