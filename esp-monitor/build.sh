#!/bin/sh
export GSM_PIN='\"$GSM_PIN\"'
export MQTT_BROKER='\"$MQTT_BROKER\"'
export MQTT_PORT='$MQTT_PORT'
export DEVICE_NAME='\"$DEVICE_NAME\"'
export MQTT_PASSWORD='\"$MQTT_PASSWORD\"'

platformio run --environment esp32dev