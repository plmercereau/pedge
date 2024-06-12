#!/bin/sh
platformio run --environment esp32dev
# TODO only copy the right file(s)
cp .pio/build/esp32dev/firmware.* /firmware/ 2>/dev/null
