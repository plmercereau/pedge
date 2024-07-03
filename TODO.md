- storage
  - use the minio chart instead of the mino-operator
  - "region"
  - maybe usefull to create a `Storage` CRD to avoid config duplication?
    - `deviceClass.spec.storageReference`
- device config partition/file
  - config-builder-esp32 docker image
    - mount secrets as volumes, and generate the csv file, then generate the bin
      - https://chatgpt.com/share/3e202c7f-90b0-4863-81f1-6f8847e84d86
  - make sure the job restarts when dependent secrets/config changed
  - store the .bin in the device secret as `config.bin` - but avoid infinite loop
- Include the lilygo code into the new esp32 firmware builder
- DeviceGroup
  - used for loading common settings e.g. wifi settings
  - use a selector e.g. on labels
    - the selector would then be used to mount the secret in the config builder
      - device > device group > device class > devices cluster
      - think of another label to sort device groups e.g. `priority`
  - `spec.secret.name`
- validation/default hooks
- flash scripts per device class
  - flash firmware + bootloader + partitions
  - fetch config from Kubernetes
  - flash config (nvs)
  - where to store the script? Inside the deviceClass artefact? Into a device artefact? ???
- device sensors
  - should we determine sensors when building the firmware, or in the config partition/file?
- develop a rancher UI interface

  - CRUD devices
  - firmware build status
  - download firmwares
  - automate script for flashing devices
  - installation script (if possible)

## Backlog

- k8s events in controllers
- move netboot stuff to another repo
- dedicated influxdb user for grafana
- dedicated minio user
- create service accounts
  - only allowed to create/manage a device cluster
  - only allowed to manage devices

## Next

- encrypt the config partition (and the firmware, too)
- wifi change over Bluetooth
- way to structure MQTT messages
  - https://github.com/homieiot/convention
  - Tasmota conventions
  - Sparkplug B conventions
- OTA upgrades over MQTT/HTTPS
  - https://chatgpt.com/share/c35ae778-766e-41ca-adc1-1f3021af7fd8
- OLM: overkill
