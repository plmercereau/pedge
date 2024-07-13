#ifndef MQTT_HELPER_H
#define MQTT_HELPER_H

#include <PubSubClient.h>
#include <esp32.h>

#include "Arduino.h"
// TODO make this work without wifi
#include "WiFi.h"
#include "utilities.h"

boolean mqttConnect();
void setupMqtt(MQTT_CALLBACK_SIGNATURE);
void connectMqtt();
const char* getSensorTopic(String sensorType);

extern WiFiClient espClient;
extern PubSubClient mqttClient;

#endif  // MQTT_HELPER_H