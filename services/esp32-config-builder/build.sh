#!/bin/bash

PARTITION_SIZE_DECIMAL=$((PARTITION_SIZE_HEXA))

export TMP_CSV_FILE="$(mktemp -d)/nvs.csv"

cleanup() {
  echo "Cleaning up..."
  rm -f "$TMP_CSV_FILE"
}

trap cleanup EXIT ERR

# Generate the CSV file from environment variables and secrets
python csv_gen.py
# Call the Python script with the converted argument
python nvs_partition_gen.py generate "$TMP_CSV_FILE" "$OUTPUT_DIR/nvs.bin" "$PARTITION_SIZE_DECIMAL"


sed "s/%PARTITION_OFFSET_HEXA%/$PARTITION_OFFSET_HEXA/" flash.sh > "$OUTPUT_DIR/flash.sh"
chmod +x "$OUTPUT_DIR/flash.sh"