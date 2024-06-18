- Operator

  - Rename "MQTTServer" to "DeviceCluster"
    - endpoint: get information from loadbalancer
  - Device secrets: rabbitmq username/password + PIN
  - firmware job

    - firmware:
      storage: # TODO
      endpoint: http://minio-service.esp-cluster.svc.cluster.local:9000
      bucket: firmware # TODO add a secret for the access key and secret
      accessKeyId: ""
      secretAccessKey: ""

  - store the firmware to an S3 bucket

- configure influxdb/grafana
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
