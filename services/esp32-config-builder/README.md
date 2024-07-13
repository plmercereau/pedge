# config-builder-esp32

A simple docker image to build an ESP32 NVS partition from secrets


## Usage
```sh
docker run --rm -v $(pwd):/secrets -v $(pwd):/output ghcr.io/plmercereau/pedge/config-builder-esp32-nvs:latest
```