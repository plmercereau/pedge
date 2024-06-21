#include <Arduino.h>
#include "utilities.h"
#include <esp32.h>
#include <ArduinoJson.h>

// TODO custom topic
#define TOPIC_SENSORS "esp-queue"

const String clientId = String(DEVICE_NAME) + "-" + getESP32ChipID();

const String topicCoordinates = (String(TOPIC_SENSORS) + "/" + DEVICE_NAME + "/coordinates");

#define TINY_GSM_MODEM_SIM7600

// Set serial for debug console (to the Serial Monitor, default speed 115200)
#define SerialMon Serial

// Set serial for AT commands (to the module)
// Use Hardware Serial on Mega, Leonardo, Micro
#define SerialAT Serial1

// See all AT commands, if wanted
// #define DUMP_AT_COMMANDS

// Define the serial console for debug prints, if needed
#define TINY_GSM_DEBUG SerialMon

// Add a reception delay, if needed.
// This may be needed for a fast processor at a slow baud rate.
// #define TINY_GSM_YIELD() { delay(2); }

// Your GPRS credentials, if any
// const char apn[]      = "YourAPN";
const char apn[] = "cmnet";
const char gprsUser[] = "";
const char gprsPass[] = "";

#include <TinyGsmClient.h>
#include <PubSubClient.h>

// TODO unused, remove
#include <Ticker.h>

#ifdef DUMP_AT_COMMANDS
#include <StreamDebugger.h>
StreamDebugger debugger(SerialAT, SerialMon);
TinyGsm modem(debugger);
#else
TinyGsm modem(SerialAT);
#endif
TinyGsmClient client(modem);

PubSubClient mqtt(client);

uint32_t lastReconnectAttempt = 0;

void mqttCallback(char *topic, byte *payload, unsigned int len) {
    SerialMon.print("Message arrived [");
    SerialMon.print(topic);
    SerialMon.print("]: ");
    SerialMon.write(payload, len);
    SerialMon.println();
}

boolean mqttConnect() {
    SerialMon.print("Connecting to MQTT Broker: ");
    SerialMon.println(MQTT_BROKER);

    // Connect to MQTT Broker
    // TODO unclear what the first parameter is ("id") - see https://pubsubclient.knolleary.net/api#connect
    boolean status = mqtt.connect(clientId.c_str(), DEVICE_NAME, MQTT_PASSWORD);

    SerialMon.print("connection ");
    if (status == false)
    {
        SerialMon.println("fail");
        return false;
    }
    SerialMon.println("success");
    // mqtt.publish(topicInit, "GsmClientTest started");
    // mqtt.subscribe(topicLed);
    return mqtt.connected();
}

void sendCoordinates() {
    JsonDocument message;

    modem.enableGPS();
    delay(15000);
    float lat      = 0;
    float lon      = 0;
    float speed    = 0;
    float alt      = 0;
    int   vsat     = 0;
    int   usat     = 0;
    float accuracy = 0;
    int   year     = 0;
    int   month    = 0;
    int   day      = 0;
    int   hour     = 0;
    int   min      = 0;
    int   sec      = 0;
    for (int8_t i = 15; i; i--) {
        SerialMon.println("Requesting current GPS/GNSS/GLONASS coordinates");
        if (modem.getGPS(&lat, &lon, &speed, &alt, &vsat, &usat, &accuracy, &year, &month, &day, &hour, &min, &sec)) {
            message["latitude"] = lat;
            message["longitude"] = lon;
            message["altitude"] = alt;
            message["speed"] = speed;
            char timestamp[25];
            sprintf(timestamp, "%04d-%02d-%02dT%02d:%02d:%02dZ", year, month, day, hour, min, sec);
            message["timestamp"] = timestamp;

            mqtt.beginPublish(topicCoordinates.c_str(), measureJson(message), true); // TODO retained = true. What does it mean?
            serializeJson(message, mqtt);
            mqtt.endPublish();

            break;
        } else {
            SerialMon.println("Couldn't get GPS/GNSS/GLONASS coordinates, retrying in 15s.");
            delay(15000L);
        }
    }

    SerialMon.println("Disabling GPS");
    modem.disableGPS();
    // Set SIM7000G GPIO4 LOW ,turn off GPS power
    // CMD:AT+SGPIO=0,4,1,0
    // Only in version 20200415 is there a function to control GPS power
    // modem.sendAT("+SGPIO=0,4,1,0");
    // if (modem.waitResponse(10000L) != 1) {
    //     SerialMon.println(" SGPIO=0,4,1,0 false ");
    // }
}

void setup() {
    // Set console baud rate
    SerialMon.begin(115200);
    delay(10);
    SerialAT.begin(UART_BAUD, SERIAL_8N1, MODEM_RX, MODEM_TX);
    // Set LED OFF
    pinMode(LED_PIN, OUTPUT);
    digitalWrite(LED_PIN, HIGH);

    pinMode(MODEM_PWRKEY, OUTPUT);
    digitalWrite(MODEM_PWRKEY, HIGH);
    // Starting the machine requires at least 1 second of low level, and with a level conversion, the levels are opposite
    delay(1000);
    digitalWrite(MODEM_PWRKEY, LOW);

    /*
    TODO unclear if it is needed - maybe for GPS
    MODEM_FLIGHT IO:25 Modulator flight mode control,
    need to enable modulator, this pin must be set to high
    */
    pinMode(MODEM_FLIGHT, OUTPUT);
    digitalWrite(MODEM_FLIGHT, HIGH);

    SerialMon.println("\nWait...");

    delay(1000);

    // Restart takes quite some time
    // To skip it, call init() instead of restart()
    DBG("Initializing modem...");
    if (!modem.init()) {
        DBG("Failed to restart modem, delaying 10s and retrying");
    }

    String ret;
    //    do {
    //        ret = modem.setNetworkMode(2);
    //        delay(500);
    //    } while (ret != "OK");
    ret = modem.setNetworkMode(2); // 2 = automatic
    DBG("setNetworkMode:", ret);

    String name = modem.getModemName();
    DBG("Modem Name:", name);

    String modemInfo = modem.getModemInfo();
    DBG("Modem Info:", modemInfo);

    if (GSM_PIN && modem.getSimStatus() != 3) {
        modem.simUnlock(GSM_PIN);
    }

    SerialMon.print("Waiting for network...");
    if (!modem.waitForNetwork()) {
        SerialMon.println(" fail");
        delay(10000);
        return;
    }
    SerialMon.println(" success");

    if (modem.isNetworkConnected()) {
        SerialMon.println("Network connected");
    }

    // GPRS connection parameters are usually set after network registration
    SerialMon.print(F("Connecting to GPRS: "));
    SerialMon.print(apn);
    if (!modem.gprsConnect(apn, gprsUser, gprsPass)) {
        SerialMon.println("fail");
        delay(10000);
        return;
    }
    SerialMon.println("success");

    // MQTT Broker setup
    mqtt.setServer(MQTT_BROKER, MQTT_PORT);
    mqtt.setCallback(mqttCallback);
}

void loop() {
    // Make sure we're still registered on the network
    if (!modem.isNetworkConnected()) {
        SerialMon.print("Network disconnected. Waiting for network... ");
        if (!modem.waitForNetwork(180000L, true)) {
            SerialMon.println("fail");
            delay(10000);
            return;
        }
        SerialMon.println("success");

        // and make sure GPRS/EPS is still connected
        if (!modem.isGprsConnected()) {
            SerialMon.println("GPRS disconnected!");
            SerialMon.print(F("Reconnecting to GPRS: "));
            SerialMon.print(apn);
            if (!modem.gprsConnect(apn, gprsUser, gprsPass)) {
                SerialMon.println(" fail");
                delay(10000);
                return;
            }
            SerialMon.println(" success");
        }
    }

    // TODO weird loop
    if (!mqtt.connected()) {
        SerialMon.println("=== MQTT NOT CONNECTED ===");
        // Reconnect every 10 seconds
        uint32_t t = millis();
        if (t - lastReconnectAttempt > 10000L) {
            lastReconnectAttempt = t;
            if (mqttConnect()) {
                lastReconnectAttempt = 0;
            }
        }
        delay(1000);
        return;
    }

    sendCoordinates();
    
    // TODO optimisation: sleep between sending data
    // esp_sleep_enable_timer_wakeup(TIME_TO_SLEEP * uS_TO_S_FACTOR);
    // delay(200);
    // esp_deep_sleep_start();
    // delay(10000);
    mqtt.loop();
}