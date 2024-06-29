- before the following:

  - understand how esp32 encryption works
    - we may need to generate and store an encryption key in a secret - but it seems esp-idf works without it
    - can we build/install/reinstall the firmware code without the encryption key?
    - can we "only" encrypt the SVN? In that case, we could build it from a job dependent on the `device` resource, and store the result into an s3 object

- validation hooks
- review "firmware"
  - Device
    - spec.deviceClassReference.name
    - device-related config stored in `device.data` key-values
      - automatically add MQTT credentials e.g. username/password, broker, port, topic name(s)
      - custom mapping, same system as for DeviceClass
      - create a job that generates the `secret.<device name>.data.file`
        - each DeviceClass would have a `spec.config.builder` docker image
        - avoid an infinite loop: https://chatgpt.com/share/618cc21b-1c80-41c3-b242-e65b0e96e6bd
  - DeviceClass spec
    - `spec.firmware.builder` (same as firmware.spec.builder)
    - specific device class-wide secrets - how, maybe a list of secret keys in the specs e.g.
      - `deviceClass.<name>.secrets.keys = ["PROVISIONNER_WIFI_SSID"]`
      - `secret.<name>-device-class.stringData.PROVISIONNER_WIFI_SSID = "abcd"`
    - the DeviceClass resource should build:
      - the firmware
      - a script that fetches the secrets and create the secrets file (could be actually standard to any device class)
      - an installation script that flashes the firmware + the secrets
      - an installation script that flashes the firmware without the secrets?
  - DevicesCluster
    - specific cluster-wide secrets, same system as for DeviceClass
  - adapt the lilygo firmware
    - installation:
      - build the code, generate the partition table, generate the NVS csv file, encrypt, flash
      - what about the encryption key?
    - code
      - try to open the file and save its key-values to NVS
    - how to build
      - with a file and a partition
      - without a file - would it keep the secrets
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

- review the firmware build/flash system
  - store secrets using NVS / efuse
- secure firmware - efuse etc
- OTA upgrades over MQTT/HTTPS
  - https://chatgpt.com/share/c35ae778-766e-41ca-adc1-1f3021af7fd8
- OLM: overkill
