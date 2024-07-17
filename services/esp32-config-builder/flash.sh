#!/bin/bash
esptool.py --chip esp32 --baud 115200 write_flash %PARTITION_OFFSET_HEXA% nvs.bin
