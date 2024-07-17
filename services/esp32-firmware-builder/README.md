# ESP32 Firmware Builder
## TODO / Notes

```

python esp_idf_nvs_partition_gen/nvs_partition_gen.py generate ../nvs_data.csv ../nvs.bin 24576


esptool.py --chip esp32 --port /dev/cu.usbserial-210 --baud 115200 write_flash 0x9000 nvs.bin

```


https://community.platformio.org/t/help-needed-with-flashing-nvs-partition-preferences-library/34832