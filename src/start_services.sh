# Create CA file form env
mkdir /app/certs
echo -e $ROOT_CA_CERT > /app/certs/root_ca.crt
echo -e $SERVER_CERT > /app/certs/server.crt
echo -e $SERVER_KEY >  /app/certs/server.key

mosquitto -c config/mosquitto.conf
./main
