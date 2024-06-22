
- publish the operator image
- publish the influx/grafana helm chart
- develop and publish an operator helm chart
- rename mqtt.queue.name to mqtt.sensorsTopic

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

- secure firmware - efuse etc
- OTA upgrades over MQTT/HTTPS
  - https://chatgpt.com/share/c35ae778-766e-41ca-adc1-1f3021af7fd8
- OLM: overkill