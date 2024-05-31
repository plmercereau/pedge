FROM python:3.9.19-slim

ENV APP_VERSION="6.1.15"

RUN apt update && \
    apt install -y git && \
    apt-get clean autoclean && apt-get autoremove --yes && rm -rf /var/lib/{apt,dpkg,cache,log}/ && \
    pip install -U platformio==${APP_VERSION} && \
    mkdir -p /workspace && \
    adduser --disabled-password pio && \
    chown -R pio:pio /workspace

USER pio
RUN platformio platform install espressif32

WORKDIR /workspace
COPY platformio.ini /workspace/
RUN platformio pkg install

COPY . ./

CMD ["/workspace/build.sh"]