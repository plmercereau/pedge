#include "mqtt_helper.h"

#include <Preferences.h>
#include <WiFi.h>

WiFiClient espClient;
PubSubClient mqttClient(espClient);

boolean mqttConnect() {
  Preferences preferences;
  preferences.begin(MQTT_NAMESPACE, false);
  String deviceName = preferences.getString(MQTT_USERNAME, "");
  String password = preferences.getString(MQTT_PASSWORD, "");
  preferences.end();
  // TODO check if all values are present

  // TODO make it non-esp32 specific
  const String clientId = String(deviceName) + "-" + getESP32ChipID();

  // Connect to MQTT Broker
  boolean status = mqttClient.connect(clientId.c_str(), deviceName.c_str(),
                                      password.c_str());

  Serial.print("connection ");
  if (status == false) {
    Serial.println("fail");
    return false;
  }
  Serial.println("success");
  return mqttClient.connected();
}

void setupMqtt(MQTT_CALLBACK_SIGNATURE) {
  Preferences preferences;
  preferences.begin(MQTT_NAMESPACE, false);
  String broker = preferences.getString(MQTT_BROKER, "");
  String strPort = preferences.getString(MQTT_PORT, "");
  preferences.end();
  // TODO check if all values are present and valid

  // convert port to int
  int port = strPort.toInt();

  IPAddress ip = IPAddress();
  if (ip.fromString(broker)) {
    mqttClient.setServer(ip, port);
    Serial.print("MQTT server set to IP address");
    Serial.print(ip);
    Serial.print(":");
    Serial.println(port);
  } else {
    // ! https://github.com/knolleary/pubsubclient/issues/375
    static char pHost[64] = {0};
    strcpy(pHost, broker.c_str());
    mqttClient.setServer(pHost, port);
    Serial.print("MQTT server set to ");
    Serial.print(pHost);
    Serial.print(":");
    Serial.println(port);
  }
  mqttClient.setCallback(callback);
}

void reconnectMqtt() {
  // Loop until we're reconnected
  while (!mqttClient.connected()) {
    Serial.print("Attempting MQTT connection... ");

    // Attempt to connect
    if (mqttConnect()) {
      Serial.println("connected");
    } else {
      Serial.print("failed, rc=");
      Serial.print(mqttClient.state());
      Serial.println(" try again in 5 seconds");
      // Wait 5 seconds before retrying
      delay(5000);
    }
  }
}

void connectMqtt() {
  if (!mqttClient.connected()) {
    reconnectMqtt();
  }
}

const char* getSensorTopic(String sensorType) {
  Preferences preferences;
  preferences.begin(MQTT_NAMESPACE, false);
  String deviceName = preferences.getString(MQTT_USERNAME, "");
  String sensorsTopic = preferences.getString(MQTT_SENSORS_TOPIC, "");
  preferences.end();
  return (sensorsTopic + "/" + deviceName + "/" + sensorType).c_str();
}
