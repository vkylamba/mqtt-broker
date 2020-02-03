import os
import logging
import paho.mqtt.client as mqtt
import dotenv

dotenv.load_dotenv()

MQTT_HOST = os.getenv("MQTT_HOST")
MQTT_PORT = int(os.getenv("MQTT_PORT"))
TEST_CLIENT_NAME = os.getenv("TEST_CLIENT_NAME")
TEST_CLIENT_IP = os.getenv("TEST_CLIENT_IP")
CLIENT_RESPONSE_TOPIC = os.getenv("CLIENT_RESPONSE_TOPIC")
CLIENT_COMMAND_TOPIC = os.getenv("CLIENT_COMMAND_TOPIC")
CLIENT_CONNECTED_TOPIC = os.getenv("CLIENT_CONNECTED_TOPIC")

logger = logging.getLogger(__name__)

# The callback for when the client receives a CONNACK response from the server.
def on_connect(client, userdata, flags, rc):
    logger.info("Connected with result code %s" % str(rc))
    # Subscribing in on_connect() means that if we lose the connection and
    # reconnect then subscriptions will be renewed.
    # client.subscribe("$SYS/#")
    client.subscribe(CLIENT_COMMAND_TOPIC)
    message = "{}-{}".format(TEST_CLIENT_NAME, TEST_CLIENT_IP)
    client.publish(CLIENT_CONNECTED_TOPIC, message)

# The callback for when a PUBLISH message is received from the server.
def on_message(client, userdata, msg):
    logger.info(msg.topic+" "+str(msg.payload))
    message = "{}-{}".format(CLIENT_NAME, CLIENT_IP)
    client.publish(CLIENT_RESPONSE_TOPIC, message)

client = mqtt.Client()
client.on_connect = on_connect
client.on_message = on_message

client.connect(MQTT_HOST, MQTT_PORT, 60)

# Blocking call that processes network traffic, dispatches callbacks and
# handles reconnecting.
# Other loop*() functions are available that give a threaded interface and a
# manual interface.
client.loop_forever()
