- Operator

  - Rename "MQTTServer" to "DeviceCluster"
    - endpoint: get information from loadbalancer
  - Rename "esp\*" to "~mcu~"
  - move netboot stuff to another repo
  - Integrate telegraf/influxdb2/grafana
    - dedicated chart
    - https://helm.sh/docs/chart_template_guide/yaml_techniques/#yaml-anchors
    - move charts/esp-cluster to operator/helm-charts?

- configure influxdb/grafana
  - Grafana map
- develop a rancher UI interface

  - CRUD devices
  - firmware build status
  - download firmwares
  - installation script (if possible)

- OperatorHub / OLM
  - [x] cert-manager
  - [x] rabbitmq
  - [x] rabbitmq message topology
  - [x] grafana
  - [ ] influxdb2
  - [ ] telegraf
  - [x] minio

## Next

- secure firmware - efuse etc
- OTA upgrades over MQTT/HTTPS
  - https://chatgpt.com/share/c35ae778-766e-41ca-adc1-1f3021af7fd8