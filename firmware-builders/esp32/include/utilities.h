
#define uS_TO_S_FACTOR \
  1000000ULL             /* Conversion factor for micro seconds to seconds */
#define TIME_TO_SLEEP 30 /* Time ESP32 will go to sleep (in seconds) */

#define UART_BAUD 115200

// ! These are the LilyGO T-SIM7600X pins
#define MODEM_TX 27
#define MODEM_RX 26
#define MODEM_PWRKEY 4
#define MODEM_DTR 32
#define MODEM_RI 33
#define MODEM_FLIGHT 25
#define MODEM_STATUS 34

#define SD_MISO 2
#define SD_MOSI 15
#define SD_SCLK 14
#define SD_CS 13

// #define LED_PIN             12
#define LED_PIN 33
// ESP32-CAM: 4=while, bright led; 33=red led. See
// https://randomnerdtutorials.com/esp32-cam-ai-thinker-pinout/
// LILYGO T-SIM7600X: 12
// https://github.com/Xinyuan-LilyGO/T-SIM7600X/blob/ac7195842b571b9d176aea8a1549f870a6d1e3ad/examples/MqttClient/MqttClient.ino#L123
// ESP32 standard: 2
// TODO https://chatgpt.com/share/3adb8057-00e3-4b0f-96b6-bf1bf2d0b881

// Settings keys in the NVS should not be longer than 15 characters
#define MQTT_NAMESPACE "mqtt"
#define MQTT_BROKER "BROKER"
#define MQTT_PORT "PORT"
#define MQTT_USERNAME "USERNAME"
#define MQTT_PASSWORD "PASSWORD"
#define MQTT_SENSORS_TOPIC "SENSORS_TOPIC"

#define CUSTOM_NAMESPACE "settings"
#define WIFI_SSID_KEY "WIFI_SSID"
#define WIFI_PASSWORD_KEY "WIFI_PASSWORD"
