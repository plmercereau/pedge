#!/bin/bash

NAMESPACE=pedge-devices-system
CLUSTER_NAME=devices-cluster
TOPIC=sensors

USERNAME=device-sample
USERNAME=device-listener
SECRET=$CLUSTER_NAME-default-user
# SECRET=$USERNAME-user-credentials

USER=$(kubectl -n $NAMESPACE get secret $SECRET -o jsonpath="{.data.username}" | base64 -d)
PASSWORD=$(kubectl -n $NAMESPACE get secret $SECRET -o jsonpath="{.data.password}" | base64 -d)

echo "mqttui -b mqtt://localhost -u $USER --password $PASSWORD $TOPIC/+/+"
mqttui -b mqtt://localhost -u $USER --password $PASSWORD $TOPIC/+/+
