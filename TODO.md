- rename "firmware" to "device installation"
  - and create a new "firmware" resource that can also be built, but not linked to a specific device (firmware without the specifics of a device)
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

## Next

- review the firmware build/flash system
  - store secrets using NVS / efuse
- secure firmware - efuse etc
- OTA upgrades over MQTT/HTTPS
  - https://chatgpt.com/share/c35ae778-766e-41ca-adc1-1f3021af7fd8
- OLM: overkill
