# pedge

## Installation

### Prerequisites

```sh
helm repo add traefik https://traefik.github.io/charts
helm install traefik traefik/traefik --version=28.2.0 --create-namespace

helm install cert-manager oci://registry-1.docker.io/bitnamicharts/cert-manager \
    --version=1.2.1 \
    --set=installCRDs=true
```

### Install InfluxDB and Grafana

```sh
export DEVICE_CLUSTER_NAME=devices-cluster
helm install influxdb-grafana oci://ghcr.io/plmercereau/pedge-charts/influxdb-grafana \
    --set=grafana.ingress.enabled=true \
    --set=grafana.ingress.ingressClassName=traefik
```

### Set an s3 bucket


### Install the Devices operator

```sh
helm install devices-operator oci://ghcr.io/plmercereau/pedge-charts/devices-operator \
    --set=rabbitmq-operator.enabled=true \
```

### Add a Devices cluster

```sh
kubectl apply -f - <<EOF
apiVersion: devices.pedge.io/v1alpha1
kind: DevicesCluster
metadata:
  name: devices-cluster
spec:
  mqtt:
    sensorsTopic: sensors
  influxdb:
    name: influxdb
    namespace: default
    secretReference:
      name: influxdb-auth
EOF
```

## Development

```sh
tilt up
```
