#!/bin/bash
platformio run --environment $TARGET
cp .pio/build/$TARGET/*.bin /firmware/
cp /home/pio/.platformio/packages/framework-arduinoespressif32/tools/partitions/boot_app0.bin /firmware/
cp flash.sh /firmware/