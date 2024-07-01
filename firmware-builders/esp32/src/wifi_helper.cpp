#include "wifi_helper.h"

void setupWifi() {
  Preferences preferences;
  preferences.begin(CUSTOM_NAMESPACE, false);

  String ssid = preferences.getString(WIFI_SSID_KEY, "");
  String password = preferences.getString(WIFI_PASSWORD_KEY, "");

  if (ssid.isEmpty() || password.isEmpty()) {
    Serial.println("No stored WiFi credentials found.");
    while (true);
    // TODO BETTER METHOD TO STOP THE DEVICE?
  } else {
    Serial.println("Stored WiFi credentials found.");
    Serial.print("Connecting to ");
    Serial.println(ssid);

    // attempt to connect to Wifi network
    while (WiFi.status() != WL_CONNECTED) {
      Serial.print(".");
      WiFi.begin(ssid.c_str(), password.c_str());
      delay(1000);  // wait 1 second for connection
    }

    Serial.println("");
    Serial.println("WiFi connected.");
    Serial.print("IP address: ");
    Serial.println(WiFi.localIP());
  }

  preferences.end();
}