$IDF_PATH/components/nvs_flash/nvs_partition_generator/nvs_partition_gen.py generate nvs_data.csv nvs.bin 24576

idf.py -p /dev/cu.usbserial-210 flash
# Instead of idf.py:
# Flash the bootloader
esptool.py --chip esp32 --port /dev/cu.usbserial-210 --baud 115200 write_flash 0x1000 build/bootloader/bootloader.bin
# Flash the partition table
esptool.py --chip esp32 --port /dev/cu.usbserial-210 --baud 115200 write_flash 0x8000 build/partition_table/partition-table.bin
# Flash the application binary
esptool.py --chip esp32 --port /dev/cu.usbserial-210 --baud 115200 write_flash 0x10000 build/your_application.bin


esptool.py --chip esp32 --port /dev/cu.usbserial-210 --baud 115200 write_flash 0x9000 nvs.bin




