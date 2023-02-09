# Create CA file form env
echo -e $ROOT_CA_CERT > /app/root_ca.crt
echo -e $SERVER_CERT > /app/server.crt
echo -e $SERVER_KEY >  /app/server.key

mosquitto -c config/mosquitto.conf
./main
