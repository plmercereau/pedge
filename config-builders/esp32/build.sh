#!/bin/bash

# Convert the hexadecimal PARTITION_SIZE to a decimal number
PARTITION_SIZE_DECIMAL=$((PARTITION_SIZE))

export CSV_FILE="$(mktemp -d)/nvs.csv"
python csv_gen.py

# Call the Python script with the converted argument
python nvs_partition_gen.py generate "$CSV_FILE" "$OUTPUT_FILE" "$PARTITION_SIZE_DECIMAL"