import paho.mqtt.client as mqtt

CLIENT_NAME = 'test client'
CLIENT_IP = 'x.z.r.d'
CLIENT_RESPONSE_TOPIC = 'CLIENT_RESPONSE'
CLIENT_COMMAND_TOPIC = 'CLIENT_COMMAND'
CLIENT_CONNECTED_TOPIC = 'CLIENT_CONNECTED'


# The callback for when the client receives a CONNACK response from the server.
def on_connect(client, userdata, flags, rc):
    print("Connected with result code "+str(rc))

    # Subscribing in on_connect() means that if we lose the connection and
    # reconnect then subscriptions will be renewed.
    # client.subscribe("$SYS/#")
    client.subscribe(CLIENT_COMMAND_TOPIC)
    client.publish(CLIENT_CONNECTED_TOPIC, f"{CLIENT_NAME}-{CLIENT_IP}")

# The callback for when a PUBLISH message is received from the server.
def on_message(client, userdata, msg):
    print(msg.topic+" "+str(msg.payload))
    client.publish(CLIENT_RESPONSE_TOPIC, f"{CLIENT_NAME}-{CLIENT_IP}")

client = mqtt.Client()
client.on_connect = on_connect
client.on_message = on_message

client.connect("localhost", 9024, 60)

# Blocking call that processes network traffic, dispatches callbacks and
# handles reconnecting.
# Other loop*() functions are available that give a threaded interface and a
# manual interface.
client.loop_forever()
