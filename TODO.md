- include the devices server in the operator
- include pvc/pv in the operator
- Include the lilygo code into the new esp32 firmware builder
- rebuild devices configs when the hostname/port/sensor topic changes
- DeviceGroup
  - used for loading common settings e.g. wifi settings
  - use a selector e.g. on labels
    - the selector would then be used to mount the secret in the config builder
      - device > device group > device class > devices cluster
      - think of another label to sort device groups e.g. `priority`
  - `spec.secret.name`
- rename DeviceCluster to DeviceCluster (singular)
- develop a rancher UI interface

  - CRUD devices
  - firmware build status
  - download firmwares
  - automate script for flashing devices
  - installation script (if possible)

## Backlog

- find a better name for 'firmware-http'
- validation/default hooks
- device sensors
  - should we determine sensors when building the firmware, or in the config partition/file?
- maybe usefull to create a `Storage` CRD to avoid config duplication?
  - `deviceClass.spec.storageReference`
- k8s events in controllers
- move netboot stuff to another repo
- dedicated influxdb user for grafana
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
