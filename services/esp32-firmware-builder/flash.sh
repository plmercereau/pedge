#!/bin/bash
esptool.py \
    --chip esp32 \
    --baud 115200 \
    --before default_reset \
    --after hard_reset write_flash \
    -z \
    --flash_mode dio \
    --flash_freq 40m \
    --flash_size 4MB \
    0x1000 bootloader.bin \
    0x8000 partitions.bin \
    0xe000 boot_app0.bin \
    0x10000 firmware.bin
