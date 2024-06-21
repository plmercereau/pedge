#!/bin/bash

NAMESPACE=default
CLUSTER_NAME=devices-cluster
QUEUE_NAME=esp-queue
SECRET=device-listener-user-credentials

USER=$(kubectl -n $NAMESPACE get secret $SECRET -o jsonpath="{.data.username}" | base64 -d)
PASSWORD=$(kubectl -n $NAMESPACE get secret $SECRET -o jsonpath="{.data.password}" | base64 -d)

echo "mqttui -b mqtt://localhost -u $USER --password $PASSWORD $QUEUE_NAME/+/coordinates"
mqttui -b mqtt://localhost -u $USER --password $PASSWORD $QUEUE_NAME/+/coordinates
