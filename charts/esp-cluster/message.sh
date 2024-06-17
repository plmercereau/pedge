#!/bin/bash
# USER=test
# PASSWORD=$(kubectl -n esp-cluster get secret test-user-credentials -o jsonpath="{.data.password}" | base64 -d)
USER=$(kubectl -n esp-cluster get secret esp-cluster-default-user -o jsonpath="{.data.username}" | base64 -d)
PASSWORD=$(kubectl -n esp-cluster get secret esp-cluster-default-user -o jsonpath="{.data.password}" | base64 -d)
SLEEP=0.5
while true;
do
RANDOM_NUMBER=$(( ( RANDOM % 15 ) + 1 ))

mqttui -b mqtt://localhost -u $USER \
    --password "$PASSWORD" \
    publish esp-queue/brussels/coordinates \
    '{ "latitude": 50.850346, "longitude": 4.851721 }'
sleep $SLEEP

mqttui -b mqtt://localhost -u $USER \
    --password "$PASSWORD" \
    publish esp-queue/test/coordinates \
    '{ "latitude": 50.0, "longitude": 3.0 }'
sleep $SLEEP

mqttui -b mqtt://localhost -u $USER \
    --password "$PASSWORD" \
    publish esp-queue/paris/coordinates \
    '{ "latitude": 48.864716, "longitude": 2.349014 }'
sleep $SLEEP

mqttui -b mqtt://localhost -u $USER \
    --password "$PASSWORD" \
    publish esp-queue/paris/coordinates \
    '{ "latitude": 48.964716, "longitude": 2.349014 }'
sleep $SLEEP

echo "loop"
done
