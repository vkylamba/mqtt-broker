# Create CA file form env
echo -e $ROOT_CA_CERT > root_ca.crt
echo -e $SERVER_CERT > server.crt
echo -e $SERVER_KEY >  server.key

mosquitto -c config/mosquitto.conf
./main
