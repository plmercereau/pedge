- Operator

  - Rename "MQTTServer" to "DeviceCluster"
    - endpoint: get information from loadbalancer
  - Integrate telegraf
    - either
      - add a telegraf CRD + controller
        - find a way to hybrid go+helm operator
          - https://github.com/operator-framework/helm-operator-plugins
      - when a new DeviceCluster / MQTTServer is added:
        - add a Telegraf CR
  - Same pattern for influxdb2 and grafana?

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
