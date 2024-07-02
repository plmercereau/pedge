FROM python:3.9-slim-bullseye

# Install necessary system packages
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    build-essential \
    libssl-dev \
    libffi-dev \
    python3-dev \
    && rm -rf /var/lib/apt/lists/*

# Install the cryptography library
RUN pip install cryptography

# Create a directory for the application
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY nvs_partition_gen.py csv_gen.py build.sh /app/

RUN mkdir -p /output /secrets

ENV OUTPUT_FILE=/output/config.bin
ENV PARTITION_SIZE=0x6000

# Set the command to run the application
ENTRYPOINT ["/app/build.sh"]