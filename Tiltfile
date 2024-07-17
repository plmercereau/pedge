load('ext://helm_resource', 'helm_resource', 'helm_repo')

helm_repo('traefik', 'https://traefik.github.io/charts', resource_name='traefik-repo')
helm_resource('traefik', 
    'traefik/traefik',
    namespace='traefik',
    flags = ['--version=28.2.0',  '--create-namespace'],
    resource_deps=['traefik-repo'])

helm_resource('cert-manager', 
    'oci://registry-1.docker.io/bitnamicharts/cert-manager',
    flags = ['--version=1.2.1', '--set=installCRDs=true'])

helm_resource('rabbitmq-cluster-operator', 
    'oci://registry-1.docker.io/bitnamicharts/rabbitmq-cluster-operator', 
    namespace='rabbitmq-system',
    flags = ['--version=4.2.10', '--create-namespace', '--set=useCertManager=true'],
    resource_deps = [ 'cert-manager'])


# TODO wait for rabbitmq-cluster-operator
k8s_yaml(helm('./charts/influxdb-grafana', values=['./charts/influxdb-grafana/values.yaml']))

docker_build('ghcr.io/plmercereau/pedge/devices-operator:0.0.1', './devices-operator')

# TODO watch files
k8s_kind('DeviceClass', image_object={'json_path': '{.spec.builder.image}', 'repo_field': 'repository', 'tag_field': 'tag'})
docker_build('ghcr.io/plmercereau/pedge/esp32-firmware-builder:latest', './services/esp32-firmware-builder')

k8s_kind('DeviceClass', image_object={'json_path': '{.spec.config.image}', 'repo_field': 'repository', 'tag_field': 'tag'})
docker_build('ghcr.io/plmercereau/pedge/esp32-config-builder:latest', './services/esp32-config-builder')

k8s_kind('DeviceCluster', image_object={'json_path': '{.spec.artefacts.image}', 'repo_field': 'repository', 'tag_field': 'tag'})
docker_build('ghcr.io/plmercereau/pedge/firmware-artefacts:latest', './services/firmware-artefacts',)

k8s_yaml(kustomize('./devices-operator/config/default'))
k8s_yaml(kustomize('./examples/manifests'))

# TODO document in the readme, development section
local_resource( 
    'ngrok',
    serve_cmd='ngrok start mqtt',
    readiness_probe=probe(
      period_secs=5,
      http_get=http_get_action(port=4040, path="/api")
   ) 
)

local_resource(
    'update-mqtt-hostname',
    """
    URL=$(curl -s localhost:4040/api/tunnels | jq -r '.tunnels[0].public_url')
    HOST=$(echo $URL | awk -F[/:] '{print $4}')
    PORT=$(echo $URL | awk -F[/:] '{print $5}')
    kubectl patch DeviceCluster my-cluster --type='json' -p="[{'op': 'replace', 'path': '/spec/mqtt/hostname', 'value': '$HOST'},{'op': 'replace', 'path': '/spec/mqtt/port', 'value': $PORT}]"
    """,
    resource_deps=['ngrok', 'my-cluster']
)