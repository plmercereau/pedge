#include <Arduino.h>
#include <ArduinoJson.h>
#include <PubSubClient.h>
#include <esp32.h>

#include "mqtt_helper.h"
#include "wifi_helper.h"

#define CAMERA_MODEL_AI_THINKER

void mqttCallback(char *topic, byte *payload, unsigned int len) {
  Serial.print("Message arrived [");
  Serial.print(topic);
  Serial.print("]: ");
  Serial.write(payload, len);
  Serial.println();
}

void setup() {
  delay(1000);
  // Set console baud rate
  Serial.begin(UART_BAUD);
  delay(10);
  Serial.println("SETUP");

  // Set LED OFF
  pinMode(LED_PIN, OUTPUT);
  digitalWrite(LED_PIN, HIGH);

  setupWifi();
  setupMqtt(mqttCallback);

  Serial.println(" success");
}

static long lastMsg = 0;

void loop() {
  connectMqtt();

  // Subscribe
  // TODO shall I put it in the loop or in main?
  // use sprintf to create the topic

  // mqttClient.subscribe((SENSOR_TOPIC + "/temperature", topic).c_str());

  mqttClient.loop();

  long now = millis();
  if (now - lastMsg > 5000) {
    lastMsg = now;
    // TODO capture error
    // uptime
    mqttClient.publish(getSensorTopic("uptime"), String(now / 1000).c_str());

    // TODO https://arduinojson.org/v6/how-to/use-arduinojson-with-pubsubclient/

    // rssi
    char buffer[1];
    buffer[0] = WiFi.RSSI();
    mqttClient.publish(getSensorTopic("rssi"), buffer, 1);
  }
}