import asyncio
import logging
import os
import time

import dotenv
import paho.mqtt.client as mqtt
from asyncio_paho import AsyncioPahoClient
from asyncio_paho.client import AsyncioMqttAuthError
from retry import retry

dotenv.load_dotenv()

logger = logging.getLogger(__name__)
logger.info = print

MQTT_HOST = os.getenv("MQTT_HOST")
MQTT_PORT = int(os.getenv("MQTT_PORT", "9024"))
TEST_CLIENT_NAME = os.getenv("TEST_CLIENT_NAME", "test-client")
TEST_CLIENT_IP = os.getenv("TEST_CLIENT_IP", "test-client-ip")
CLIENT_RESPONSE_TOPIC = os.getenv("CLIENT_RESPONSE_TOPIC",  "test-client-resp-topic")
CLIENT_COMMAND_TOPIC = os.getenv("CLIENT_COMMAND_TOPIC", "/iot-gw-v3-user/device_commands")
CLIENT_CONNECTED_TOPIC = os.getenv("CLIENT_CONNECTED_TOPIC", "test-client-connected-topic")

CLIENT_HEARTBEAT_TOPIC = os.getenv("CLIENT_HEARTBEAT_TOPIC", "/iotaapsys/services/heartbeat")
CLIENT_SYNC_TOPIC = os.getenv("CLIENT_HEARTBEAT_TOPIC", "/iot-gw-v3-user/sync")


# The callback for when the client receives a CONNACK response from the server.
async def on_connect(client, userdata, flags_dict, result):
    logger.info("Connected with result code  %s" % str(result))

    # Subscribing in on_connect() means that if we lose the connection and
    # reconnect then subscriptions will be renewed.
    # client.subscribe("$SYS/#")

    await client.asyncio_subscribe(CLIENT_HEARTBEAT_TOPIC)
    await client.asyncio_subscribe(CLIENT_CONNECTED_TOPIC)
    await client.asyncio_subscribe(CLIENT_RESPONSE_TOPIC)
    await client.asyncio_subscribe(CLIENT_COMMAND_TOPIC)
    await client.asyncio_subscribe(CLIENT_SYNC_TOPIC)


# The callback for when a PUBLISH message is received from the server.
async def on_message(client, userdata, msg):
    logger.info("Message received on topic {}: {}".format(msg.topic, str(msg.payload)))


async def publish_heartbeats(client):
    while True:
        epoch_now = int(time.time() * 1000)
        client.publish(CLIENT_HEARTBEAT_TOPIC, f"epoch_ms {epoch_now}")
        await asyncio.sleep(60)


async def publish_commands(client):
    while True:
        client.publish(CLIENT_COMMAND_TOPIC, "test command")
        await asyncio.sleep(120)


@retry(delay=2)
async def main():
    async with AsyncioPahoClient() as client:
        client.asyncio_listeners.add_on_connect(on_connect)
        client.asyncio_listeners.add_on_message(on_message)
        await client.asyncio_connect(MQTT_HOST, MQTT_PORT, 60)
        await asyncio.gather(publish_heartbeats(client), publish_commands(client))


if __name__ == '__main__':
    asyncio.run(main())
