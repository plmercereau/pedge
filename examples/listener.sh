#!/bin/bash
# SERVER=6.tcp.eu.ngrok.io:13468
SERVER=localhost
NAMESPACE=default
CLUSTER_NAME=my-cluster
TOPIC=sensors

# USERNAME=esp32cam
# SECRET=$USERNAME-user-credentials
USERNAME=device-listener
SECRET=$CLUSTER_NAME-default-user

USER=$(kubectl -n $NAMESPACE get secret $SECRET -o jsonpath="{.data.username}" | base64 -d)
PASSWORD=$(kubectl -n $NAMESPACE get secret $SECRET -o jsonpath="{.data.password}" | base64 -d)

echo "mqttui -b mqtt://$SERVER -u $USER --password $PASSWORD $TOPIC/+/+"
mqttui -b mqtt://$SERVER -u $USER --password $PASSWORD $TOPIC/+/+
