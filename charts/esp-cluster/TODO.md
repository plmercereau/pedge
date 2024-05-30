- create jobs to build the firmware
  - platformio builder image
    - include the firmware code inside the docker image
    - env vars: user name, SIM PIN, password
- store the firmware to an S3 bucket
- convert the chart to an operator
  - "Device" CRD
  - "DeviceCluster" CRD
- configure prometheus/grafana
  - MQTT exporter
  - Grafana map
- develop a rancher UI interface
  - CRUD devices
  - firmware build status
  - download firmwares
  - installation script (if possible)

## Next

- secure firmware - efuse etc
- OTA upgrades over MQTT/HTTPS
  - https://chatgpt.com/share/c35ae778-766e-41ca-adc1-1f3021af7fd8
