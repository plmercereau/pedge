#!/bin/bash

# Convert the hexadecimal PARTITION_SIZE to a decimal number
PARTITION_SIZE_DECIMAL=$((PARTITION_SIZE))

export TMP_CSV_FILE="$(mktemp -d)/nvs.csv"

cleanup() {
  echo "Cleaning up..."
  rm -f "$TMP_CSV_FILE"
}

trap cleanup EXIT ERR

# Generate the CSV file from environment variables and secrets
python csv_gen.py
# Call the Python script with the converted argument
python nvs_partition_gen.py generate "$TMP_CSV_FILE" "$OUTPUT_FILE" "$PARTITION_SIZE_DECIMAL"

