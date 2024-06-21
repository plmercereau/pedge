- endpoint: get information from loadbalancer
- move netboot stuff to another repo
- publish the operator image
- publish the influx/grafana helm chart
- develop and publish an operator helm chart
- dedicated influxdb user for grafana
- dedicated minio user
- develop a rancher UI interface

  - CRUD devices
  - firmware build status
  - download firmwares
  - automate script for flashing devices
  - installation script (if possible)

## Next

- secure firmware - efuse etc
- OTA upgrades over MQTT/HTTPS
  - https://chatgpt.com/share/c35ae778-766e-41ca-adc1-1f3021af7fd8
- OperatorHub / OLM
  - [x] cert-manager
  - [x] rabbitmq
  - [x] rabbitmq message topology
  - [x] grafana
  - [ ] influxdb2
  - [ ] telegraf
  - [x] minio
