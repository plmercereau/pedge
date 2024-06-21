#!/bin/sh
platformio run --environment esp32dev
cp .pio/build/esp32dev/firmware.bin /firmware/
