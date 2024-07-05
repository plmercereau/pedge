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

helm_repo('minio', 'https://operator.min.io')
helm_resource('minio-operator',
    'minio/operator',
    namespace='minio-operator',
    flags = ['--version=5.0.15', '--create-namespace']
)

k8s_yaml(kustomize('./examples/manifests'))

# TODO ideally we should create a separate resource, this puts stuff in the devices operator
# TODO e.g. [event: tenant devices-cluster] Tenant is missing root credentials
k8s_resource('pedge-devices-controller-manager', 
    objects=['storage:Tenant:default'],
    resource_deps=['rabbitmq-cluster-operator', 'minio-operator'])

# TODO wait for rabbitmq-cluster-operator
k8s_yaml(helm('./charts/influxdb-grafana', values=['./charts/influxdb-grafana/values.yaml']))

docker_build('ghcr.io/plmercereau/pedge/devices-operator:0.0.1', './devices-operator')

k8s_kind('DeviceClass', image_object={'json_path': '{.spec.builder.image}', 'repo_field': 'repository', 'tag_field': 'tag'})
docker_build('ghcr.io/plmercereau/pedge/firmware-builder-esp32:latest', './firmware-builders/esp32')
k8s_kind('DeviceClass', image_object={'json_path': '{.spec.config.image}', 'repo_field': 'repository', 'tag_field': 'tag'})
docker_build('ghcr.io/plmercereau/pedge/config-builder-esp32:latest', './config-builders/esp32')

k8s_yaml(kustomize ( './devices-operator/config/default'))