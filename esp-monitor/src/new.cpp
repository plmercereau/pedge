// TODO https://github.com/Xinyuan-LilyGO/T-SIM7600X/blob/master/examples/MqttClient/MqttClient.ino

#include <Arduino.h> 
#include "utilities.h"

#define TINY_GSM_MODEM_SIM7600

// Set serial for debug console (to the Serial Monitor, default speed 115200)
#define SerialMon Serial

// Set serial for AT commands (to the module)
// Use Hardware Serial on Mega, Leonardo, Micro
#define SerialAT Serial1

// See all AT commands, if wanted
//#define DUMP_AT_COMMANDS

// Define the serial console for debug prints, if needed
#define TINY_GSM_DEBUG SerialMon

// Add a reception delay, if needed.
// This may be needed for a fast processor at a slow baud rate.
// #define TINY_GSM_YIELD() { delay(2); }

// Your GPRS credentials, if any
// const char apn[]      = "YourAPN";
const char apn[]      = "cmnet";
const char gprsUser[] = "";
const char gprsPass[] = "";

#include <TinyGsmClient.h>
#include <Ticker.h>

#ifdef DUMP_AT_COMMANDS
#include <StreamDebugger.h>
StreamDebugger debugger(SerialAT, SerialMon);
TinyGsm        modem(debugger);
#else
TinyGsm        modem(SerialAT);
#endif
TinyGsmClient client(modem);

#include <ArduinoHttpClient.h>
// Server details
const char server[]   = "vsh.pp.ua";
const char resource[] = "/TinyGSM/logo.txt";
const int  port       = 80;
HttpClient    http(client, server, port);
void request() {
  SerialMon.print(F("Performing HTTP GET request... "));
  int err = http.get(resource);
  if (err != 0) {
    SerialMon.println(F("failed to connect"));
    delay(10000);
    return;
  }

  int status = http.responseStatusCode();
  SerialMon.print(F("Response status code: "));
  SerialMon.println(status);
  if (!status) {
    delay(10000);
    return;
  }

  String body = http.responseBody();
  SerialMon.println(F("Response:"));
  SerialMon.println(body);

  SerialMon.print(F("Body length is: "));
  SerialMon.println(body.length());

  // Shutdownw

  http.stop();
  SerialMon.println(F("Server disconnected"));
}

void setup()
{
    // Set console baud rate
    Serial.begin(115200);
    delay(10);
    SerialAT.begin(UART_BAUD, SERIAL_8N1, MODEM_RX, MODEM_TX);
    // Set LED OFF
    pinMode(LED_PIN, OUTPUT);
    digitalWrite(LED_PIN, HIGH);

    pinMode(MODEM_PWRKEY, OUTPUT);
    digitalWrite(MODEM_PWRKEY, HIGH);
    delay(300);
    digitalWrite(MODEM_PWRKEY, LOW);

    /*
    MODEM_FLIGHT IO:25 Modulator flight mode control,
    need to enable modulator, this pin must be set to high
    */
    pinMode(MODEM_FLIGHT, OUTPUT);
    digitalWrite(MODEM_FLIGHT, HIGH);

    Serial.println("\nWait...");

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

    #ifdef GSM_PIN
    // Unlock your SIM card with a PIN if needed
    if (modem.getSimStatus() != 3) {
        modem.simUnlock(GSM_PIN);
    }
    #endif

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
    SerialMon.print(F("Connecting to "));
    SerialMon.print(apn);
    if (!modem.gprsConnect(apn, gprsUser, gprsPass)) {
        SerialMon.println(" fail");
        delay(10000);
        return;
    }
    SerialMon.println(" success");

    if (modem.isGprsConnected()) {
        SerialMon.println("GPRS connected");
    }


}

void loop()
{

    // Make sure we're still registered on the network
    if (!modem.isNetworkConnected()) {
        SerialMon.println("Network disconnected");
        if (!modem.waitForNetwork(180000L, true)) {
            SerialMon.println(" fail");
            delay(10000);
            return;
        }
        if (modem.isNetworkConnected()) {
            SerialMon.println("Network re-connected");
        }

        // and make sure GPRS/EPS is still connected
        if (!modem.isGprsConnected()) {
            SerialMon.println("GPRS disconnected!");
            SerialMon.print(F("Connecting to "));
            SerialMon.print(apn);
            if (!modem.gprsConnect(apn, gprsUser, gprsPass)) {
                SerialMon.println(" fail");
                delay(10000);
                return;
            }
            if (modem.isGprsConnected()) {
                SerialMon.println("GPRS reconnected");
            }
        }
    }
    request();


}