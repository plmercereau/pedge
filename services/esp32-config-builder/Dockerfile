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
COPY . /app/

RUN mkdir -p /output /secrets

ENV OUTPUT_DIR=/output
ENV SECRETS_PATH=/secrets
ENV PARTITION_SIZE_HEXA=0x6000
ENV PARTITION_OFFSET_HEXA=0x9000

# Set the command to run the application
ENTRYPOINT ["/app/build.sh"]