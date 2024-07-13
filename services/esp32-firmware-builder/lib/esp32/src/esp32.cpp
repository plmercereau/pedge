#include "esp32.h"
#include <Arduino.h>

String getESP32ChipID() {
  uint64_t chipid = ESP.getEfuseMac();
  char chipID[17];
  snprintf(chipID, sizeof(chipID), "%04X%08X", (uint16_t)(chipid>>32), (uint32_t)chipid);
  return String(chipID);
}
