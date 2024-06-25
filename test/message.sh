#!/bin/bash

NAMESPACE=default
CLUSTER_NAME=devices-cluster
TOPIC=sensors

SECRET=$CLUSTER_NAME-default-user
# SECRET=device-sample-user-credentials
USER=$(kubectl -n $NAMESPACE get secret $SECRET -o jsonpath="{.data.username}" | base64 -d)
PASSWORD=$(kubectl -n $NAMESPACE get secret $SECRET -o jsonpath="{.data.password}" | base64 -d)
SLEEP=0.5

while true;
do
RANDOM_NUMBER=$(( ( RANDOM % 15 ) + 1 ))

mqttui -b mqtt://localhost -u $USER \
    --password "$PASSWORD" \
    publish $TOPIC/brussels/coordinates \
    '{ "latitude": 50.850346, "longitude": 4.851721 }'
sleep $SLEEP

mqttui -b mqtt://localhost -u $USER \
    --password "$PASSWORD" \
    publish $TOPIC/test/coordinates \
    '{ "latitude": 50.0, "longitude": 3.0 }'
sleep $SLEEP

mqttui -b mqtt://localhost -u $USER \
    --password "$PASSWORD" \
    publish $TOPIC/paris/coordinates \
    '{ "latitude": 48.864716, "longitude": 2.349014 }'
sleep $SLEEP

mqttui -b mqtt://localhost -u $USER \
    --password "$PASSWORD" \
    publish $TOPIC/paris/coordinates \
    '{ "latitude": 48.964716, "longitude": 2.349014 }'
sleep $SLEEP

echo "loop"
done
