#include "mqtt_helper.h"

#include <WiFi.h>

WiFiClient espClient;
PubSubClient mqttClient(espClient);

boolean mqttConnect() {
  Preferences preferences;
  preferences.begin(MQTT_NAMESPACE, false);
  String username = preferences.getString(MQTT_USERNAME, "");
  String password = preferences.getString(MQTT_PASSWORD, "");
  preferences.end();
  // TODO check if all values are present

  // TODO make it non-esp32 specific
  const String clientId = String(username) + "-" + getESP32ChipID();

  // Connect to MQTT Broker
  boolean status =
      mqttClient.connect(clientId.c_str(), username.c_str(), password.c_str());

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
  } else {
    mqttClient.setServer(broker.c_str(), port);
  }
  mqttClient.setCallback(callback);
}

void reconnectMqtt() {
  // Loop until we're reconnected
  while (!mqttClient.connected()) {
    Serial.print("Attempting MQTT connection to ");

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
