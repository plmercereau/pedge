#include "wifi_helper.h"
#define LED_PIN 33
// ESP32-CAM: 4=while, bright led; 33=red led. See
// https://randomnerdtutorials.com/esp32-cam-ai-thinker-pinout/
// LILYGO T-SIM7600X: 12
// https://github.com/Xinyuan-LilyGO/T-SIM7600X/blob/ac7195842b571b9d176aea8a1549f870a6d1e3ad/examples/MqttClient/MqttClient.ino#L123
// ESP32 standard: 2
// TODO https://chatgpt.com/share/3adb8057-00e3-4b0f-96b6-bf1bf2d0b881

void setup() {
  Serial.begin(115200);

  delay(100);

  // Initialize the LED pin as an output
  pinMode(LED_PIN, OUTPUT);

  connectToWiFi();
}

void loop() {
  // Your code here
  Serial.println("Hello World");

  digitalWrite(LED_PIN, HIGH);
  // Wait for a second (1000 milliseconds)
  delay(1000);
  // Turn the LED off
  digitalWrite(LED_PIN, LOW);
  // Wait for a second (1000 milliseconds)
  delay(1000);
}