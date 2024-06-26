# pedge

## Installation

### Install InfluxDB and Grafana

```sh
export DEVICE_CLUSTER_NAME=devices-cluster
helm install influxdb-grafana oci://ghcr.io/plmercereau/pedge-charts/influxdb-grafana \
    --set=grafana.ingress.enabled=true \
    --set=grafana.ingress.ingressClassName=traefik
```

### Install the Devices operator

```sh
helm install devices-operator oci://ghcr.io/plmercereau/pedge-charts/devices-operator \
    --set=traefik.enabled=true \
    --set=cert-manager.enabled=true \
    --set=rabbitmq-operator.enabled=true \
    --set=minio-operator.enabled=true
```

### Add a Devices cluster

```sh
kubectl apply -f - <<EOF
apiVersion: devices.pedge.io/v1alpha1
kind: DeviceCluster
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
