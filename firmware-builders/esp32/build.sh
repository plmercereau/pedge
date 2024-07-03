#!/bin/sh
platformio run --environment $TARGET
cp .pio/build/$TARGET/*.bin /firmware/
