#!/bin/bash

SERVER=devices.local
DEVICE=esp32cam


TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT ERR

flash_firmware() {
    FIRMARE_URL=$SERVER/firmwares/$DEVICE/firmware.tgz
    FIRMWARE_DIR=$TEMP_DIR/firmware
    mkdir -p $FIRMWARE_DIR

    curl $FIRMARE_URL --output - | tar -xz -C $FIRMWARE_DIR
    cd $FIRMWARE_DIR
    ./flash.sh
}

flash_config() {
    CONFIG_URL=$SERVER/configurations/$DEVICE/config.tgz
    CONFIG_DIR=$TEMP_DIR/configuration
    mkdir -p $CONFIG_DIR

    curl $CONFIG_URL --output - | tar -xz -C $CONFIG_DIR
    cd $CONFIG_DIR
    ./flash.sh
}

case $1 in
    firmware)
        flash_firmware
        ;;
    config)
        flash_config
        ;;
    all)
        update_firmware
        flash_config
        ;;
    *)
        echo "Usage: $0 {firmware|config|all}"
        exit 1
        ;;
esac