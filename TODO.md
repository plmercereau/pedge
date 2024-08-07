- Include the lilygo code into the new esp32 firmware builder
- develop a rancher UI interface

  - CRUD devices
  - firmware build status
  - download firmwares
  - automate script for flashing devices
  - installation script (if possible)

## Backlog

- DeviceGroup
  - used for loading common settings e.g. wifi settings
  - use a selector e.g. on labels
    - the selector would then be used to mount the secret in the config builder
      - device > device group > device class > devices cluster
      - think of another label to sort device groups e.g. `priority`
  - `spec.secret.name`
- validation/default hooks
- implement the firmware build job in the same spirit as the config builder job
- device sensors
  - should we determine sensors when building the firmware, or in the config partition/file?
- k8s events in controllers
- move netboot stuff to another repo
- dedicated influxdb user for grafana
- create service accounts
  - only allowed to create/manage a device cluster
  - only allowed to manage devices
- Development instructions  
  - make ngrok optional

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
