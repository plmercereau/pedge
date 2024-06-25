package controller

import (
	"context"
	"fmt"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	pedgev1alpha1 "github.com/plmercereau/pedge/api/v1alpha1"
	rabbitmqv1 "github.com/rabbitmq/cluster-operator/api/v1beta1"
	rabbitmqtopologyv1 "github.com/rabbitmq/messaging-topology-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	s3AccessKeyId             = "accesskey" // ! cannot be changed - depends on the MinIO operator
	s3SecretAccessKey         = "secretkey" // ! cannot be changed - depends on the MinIO operator
	deviceClusterSecretSuffix = "-device-cluster"
	bucketName                = "firmwares"
	bucketRegion              = "default"
	listenerUserName          = "device-listener"
)

// DeviceClusterReconciler reconciles a DeviceCluster object
type DeviceClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=devices.pedge.io,resources=deviceclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devices.pedge.io,resources=deviceclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rabbitmq.com,resources=rabbitmqclusters;vhosts;queues;permissions;topicpermissions;users,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=minio.min.io,resources=tenants,verbs=get;list;watch;create;update;patch;delete

func (r *DeviceClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the DeviceCluster instance
	server := &pedgev1alpha1.DeviceCluster{}
	err := r.Get(ctx, req.NamespacedName, server)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "unable to fetch DeviceCluster")
		return ctrl.Result{}, err
	}

	// Secret
	secretName := server.Name + deviceClusterSecretSuffix
	// TODO decidated minio user
	// * see https://github.com/minio/operator/blob/master/examples/kustomization/base/storage-user.yaml
	// * and https://github.com/minio/operator/blob/fd7ede7ba9b5e0c4730284afff84c1350933f848/examples/kustomization/base/tenant.yaml#L33
	var secret corev1.Secret
	if err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: server.Namespace}, &secret); err != nil {
		logger.Info("Creating new secret " + secretName)
		accessKey := generateRandomPassword(20)
		secretKey := generateRandomPassword(40)
		secret = corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: server.Namespace,
			},
			Data: map[string][]byte{
				s3AccessKeyId:     []byte(accessKey),
				s3SecretAccessKey: []byte(secretKey),
				"config.env":      []byte(fmt.Sprintf("export MINIO_ROOT_USER=%s\nexport MINIO_ROOT_PASSWORD=%s", accessKey, secretKey)),
			},
		}
		// Set the ownerRef for the secret to ensure it gets cleaned up when the device cluster is deleted
		// if err := ctrl.SetControllerReference(server, secret, r.Scheme); err != nil {
		// 	return ctrl.Result{}, err
		// }

		if err := r.Create(ctx, &secret); err != nil {
			logger.Error(err, "Unable to create secret "+secretName)
			return ctrl.Result{}, err
		}
	} else {
		changed := false
		s3Key := string(secret.Data[s3AccessKeyId])
		if s3Key == "" {
			logger.Info("Creating a default value for" + s3AccessKeyId + " in secret " + secretName)
			s3Key = generateRandomPassword(20)
			secret.Data[s3AccessKeyId] = []byte(s3Key)
			changed = true
		}
		s3Secret := string(secret.Data[s3SecretAccessKey])
		if s3Secret == "" {
			logger.Info("Creating a default value for" + s3SecretAccessKey + " in secret " + secretName)
			s3Secret = generateRandomPassword(40)
			secret.Data[s3SecretAccessKey] = []byte(s3Secret)
			changed = true
		}
		configEnv := fmt.Sprintf("export MINIO_ROOT_USER=%s\nexport MINIO_ROOT_PASSWORD=%s", s3Key, s3Secret)
		if string(secret.Data["config.env"]) != configEnv {
			logger.Info("Update config.env value for in secret " + secretName)
			secret.Data["config.env"] = []byte(configEnv)
			changed = true
		}
		if changed {
			logger.Info("Updating secret " + secretName)
			if err := r.Update(ctx, &secret); err != nil {
				logger.Error(err, "unable to update secret "+secretName)
				return ctrl.Result{}, err
			}
		}

	}

	// Sync resources
	if err := r.syncResources(ctx, server); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// syncResources creates or updates the associated resources
func (r *DeviceClusterReconciler) syncResources(ctx context.Context, server *pedgev1alpha1.DeviceCluster) error {
	// Define the desired RabbitMQ Cluster resource
	cluster := &rabbitmqv1.RabbitmqCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name,
			Namespace: server.Namespace,
		},
		Spec: rabbitmqv1.RabbitmqClusterSpec{
			Service: rabbitmqv1.RabbitmqClusterServiceSpec{
				Type: "LoadBalancer",
			},
			Rabbitmq: rabbitmqv1.RabbitmqClusterConfigurationSpec{
				AdditionalPlugins: []rabbitmqv1.Plugin{"rabbitmq_mqtt"},
				// TODO only for testing purposes
				EnvConfig: `RABBITMQ_LOGS=""`,
				AdditionalConfig: `
log.console = true
log.console.level = debug
`,
			},
		},
	}
	controllerutil.SetOwnerReference(server, cluster, r.Scheme)

	vhost := &rabbitmqtopologyv1.Vhost{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name + "-default",
			Namespace: server.Namespace,
		},
		Spec: rabbitmqtopologyv1.VhostSpec{
			Name: "/",
			RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
				Name:      server.Name,
				Namespace: server.Namespace,
			},
		},
	}
	controllerutil.SetOwnerReference(server, vhost, r.Scheme)
	// We only create the vhost if it doesn't exist. The RabbitMQ messaging topology operator does not allow to modify it.
	// TODO we should also block some updates on the device cluster name - through a validation webhook
	existingVhost := vhost.DeepCopyObject().(client.Object)
	if err := r.Get(ctx, client.ObjectKeyFromObject(vhost), existingVhost); err != nil && errors.IsNotFound(err) {
		return r.Create(ctx, vhost)
	}

	queue := &rabbitmqtopologyv1.Queue{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name,
			Namespace: server.Namespace,
		},
		Spec: rabbitmqtopologyv1.QueueSpec{
			Name:  server.Spec.MQTT.SensorsTopic,
			Vhost: vhost.Name,
			RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
				Name:      server.Name,
				Namespace: server.Namespace,
			},
		},
	}
	controllerutil.SetOwnerReference(server, queue, r.Scheme)
	// We only create the queue if it doesn't exist. The RabbitMQ messaging topology operator does not allow to modify it.
	existingQueue := queue.DeepCopyObject().(client.Object)
	if err := r.Get(ctx, client.ObjectKeyFromObject(queue), existingQueue); err != nil && errors.IsNotFound(err) {
		return r.Create(ctx, queue)
	}

	var listenerSecret corev1.Secret
	secretName := listenerUserName + deviceSecretSuffix
	if err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: server.Namespace}, &listenerSecret); err != nil {
		listenerSecret = corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: server.Namespace,
			},
			Data: map[string][]byte{
				"username": []byte(listenerUserName),
				"password": []byte(generateRandomPassword(16)),
			},
		}

		if err := r.Create(ctx, &listenerSecret); err != nil {
			return err
		}
	} else {
		if string(listenerSecret.Data["username"]) != listenerUserName {
			listenerSecret.Data["username"] = []byte(listenerUserName)
			if err := r.Update(ctx, &listenerSecret); err != nil {
				return err
			}
		}
	}

	// Store MQTT URL/TOPIC/USERNAME/PASSWORD in influxdb-auth in the influxdb namespace
	if server.Spec.InfluxDB != (pedgev1alpha1.InfluxDB{}) {
		influxDBSecret := &corev1.Secret{}
		err := r.Get(ctx, types.NamespacedName{Name: server.Spec.InfluxDB.SecretReference.Name, Namespace: server.Spec.InfluxDB.Namespace}, influxDBSecret)
		if err != nil {
			// If a secret name is provided, then it must exist
			// TODO in such cases, create an Event for the user to understand why their reconcile is failing.
			return err
		}

		changed := false

		// Add reloader.stakater.com/match: "true" to the secret to trigger a reload of the telegraf config
		if influxDBSecret.Annotations == nil {
			influxDBSecret.Annotations = make(map[string]string)
			// check if reloader.stakater.com/match exists and is set to "true"
			if influxDBSecret.Annotations["reloader.stakater.com/match"] != "true" {
				influxDBSecret.Annotations["reloader.stakater.com/match"] = "true"
				changed = true
			}
		}

		if influxDBSecret.Data["MQTT_USERNAME"] == nil || string(influxDBSecret.Data["MQTT_USERNAME"]) != listenerUserName {
			influxDBSecret.Data["MQTT_USERNAME"] = []byte(listenerUserName)
			changed = true
		}
		if influxDBSecret.Data["MQTT_PASSWORD"] == nil || string(influxDBSecret.Data["MQTT_PASSWORD"]) != string(listenerSecret.Data["password"]) {
			influxDBSecret.Data["MQTT_PASSWORD"] = listenerSecret.Data["password"]
			changed = true
		}
		if influxDBSecret.Data["MQTT_TOPIC"] == nil || string(influxDBSecret.Data["MQTT_TOPIC"]) != server.Spec.MQTT.SensorsTopic {
			influxDBSecret.Data["MQTT_TOPIC"] = []byte(server.Spec.MQTT.SensorsTopic)
			changed = true
		}
		mqttUrl := fmt.Sprintf("tcp://%s.%s.svc:%s", server.Name, server.Namespace, "1883")
		if influxDBSecret.Data["MQTT_URL"] == nil || string(influxDBSecret.Data["MQTT_URL"]) != mqttUrl {
			influxDBSecret.Data["MQTT_URL"] = []byte(mqttUrl)
			changed = true
		}
		// ! For some reason, telegraf does not allow admin-token as an environment variable, maybe because of the dash
		if influxDBSecret.Data["INFLUXDB_TOKEN"] == nil || string(influxDBSecret.Data["INFLUXDB_TOKEN"]) != string(influxDBSecret.Data["admin-token"]) {
			influxDBSecret.Data["INFLUXDB_TOKEN"] = influxDBSecret.Data["admin-token"]
			changed = true
		}
		if changed {
			if err := r.Update(ctx, influxDBSecret); err != nil {
				return err
			}
		}
	}

	listenerUser := &rabbitmqtopologyv1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name:      listenerUserName,
			Namespace: server.Namespace,
			Annotations: map[string]string{
				// Needed to trigger a reconciliation when the password changes
				secretVersionAnnotation: listenerSecret.GetResourceVersion(),
			},
		},
		Spec: rabbitmqtopologyv1.UserSpec{
			RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
				Name:      server.Name,
				Namespace: server.Namespace,
			},
			ImportCredentialsSecret: &corev1.LocalObjectReference{
				Name: listenerSecret.Name,
			},
		},
	}
	// TODO it seems rabbitmq is already watching/owning the user, and when set to server, the permissions are not applied
	// For now, accept users are not deleted with the device cluster...
	// controllerutil.SetOwnerReference(server, listenerUser, r.Scheme)

	// Define the desired RabbitMQ Permission resource
	listenerPermission := &rabbitmqtopologyv1.Permission{
		ObjectMeta: metav1.ObjectMeta{
			Name:      listenerUser.Name,
			Namespace: listenerUser.Namespace,
		},
		Spec: rabbitmqtopologyv1.PermissionSpec{
			Vhost: "/",
			UserReference: &corev1.LocalObjectReference{
				Name: listenerUser.Name,
			},
			Permissions: rabbitmqtopologyv1.VhostPermissions{
				Configure: "^mqtt-subscription-.*$",
				Write:     "^amq\\.topic$|^mqtt-subscription-.*$",
				Read:      "^amq\\.topic$|^mqtt-subscription-.*$",
			},
			RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
				Name:      server.Name,
				Namespace: server.Namespace,
			},
		},
	}
	// controllerutil.SetOwnerReference(server, listenerPermission, r.Scheme)

	// Define the desired RabbitMQ TopicPermission resource
	listenerTopicPermission := &rabbitmqtopologyv1.TopicPermission{
		ObjectMeta: metav1.ObjectMeta{
			Name:      listenerUser.Name,
			Namespace: listenerUser.Namespace,
		},
		Spec: rabbitmqtopologyv1.TopicPermissionSpec{
			Vhost: "/",
			UserReference: &corev1.LocalObjectReference{
				Name: listenerUser.Name,
			},
			Permissions: rabbitmqtopologyv1.TopicPermissionConfig{
				Exchange: "amq.topic",
				Write:    "",
				Read:     fmt.Sprintf("^%s\\..+$", server.Spec.MQTT.SensorsTopic),
			},
			RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
				Name:      server.Name,
				Namespace: server.Namespace,
			},
		},
	}
	// controllerutil.SetOwnerReference(server, listenerTopicPermission, r.Scheme)

	// self-signed certificate. If exposing over the internet, use cert-manager + letsencrypt
	requestAutoCert := true
	tenant := &miniov2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name,
			Namespace: server.Namespace,
			// the minio controller is not behaving well when using controllerutil.SetOwnerReference
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(server, pedgev1alpha1.GroupVersion.WithKind("DeviceCluster")),
			},
		},
		Spec: miniov2.TenantSpec{
			Pools: []miniov2.Pool{
				{
					Name:             "minio-pool-default",
					Servers:          4, // TODO may be a bit too much, try a lower value
					VolumesPerServer: 4,
					VolumeClaimTemplate: &corev1.PersistentVolumeClaim{
						ObjectMeta: metav1.ObjectMeta{
							Name: "data",
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							AccessModes: []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"},
							Resources: corev1.VolumeResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceStorage: resource.MustParse("2Gi"),
								},
							},
						},
					},
				},
			},
			Buckets: []miniov2.Bucket{
				{
					Name:          bucketName,
					Region:        bucketRegion,
					ObjectLocking: false,
				},
			},
			// CredsSecret is not working anymore: https://github.com/minio/operator/blob/master/pkg/apis/minio.min.io/v2/types.go#L356C2-L356C15
			Configuration: &corev1.LocalObjectReference{
				Name: server.Name + deviceClusterSecretSuffix,
			},
			RequestAutoCert: &requestAutoCert,
		},
	}

	resources := []client.Object{cluster, tenant, listenerUser, listenerPermission, listenerTopicPermission}
	for _, res := range resources {
		if err := r.CreateOrUpdate(ctx, res); err != nil {
			return err
		}
	}
	return nil
}

// creates or updates a resource
func (r *DeviceClusterReconciler) CreateOrUpdate(ctx context.Context, obj client.Object) error {
	key := types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}
	existing := obj.DeepCopyObject().(client.Object)
	err := r.Get(ctx, key, existing)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			return r.Create(ctx, obj)
		}
		return err
	}
	for i := 0; i < 5; i++ { // Retry up to 5 times
		obj.SetResourceVersion(existing.GetResourceVersion())
		err = r.Update(ctx, obj)
		if err == nil {
			return nil
		}
		if errors.IsConflict(err) {
			// Fetch the latest version of the object
			err = r.Get(ctx, key, existing)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return fmt.Errorf("failed to update resource %s/%s after multiple attempts: %w", obj.GetNamespace(), obj.GetName(), err)
}

// SetupWithManager sets up the controller with the Manager
func (r *DeviceClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pedgev1alpha1.DeviceCluster{}).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForInfluxSecret),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}

func (r *DeviceClusterReconciler) findObjectsForInfluxSecret(ctx context.Context, secret client.Object) []reconcile.Request {
	attachedDevices := &pedgev1alpha1.FirmwareList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.AndSelectors(
			fields.OneTermEqualSelector(".spec.influxDB.secretReference.name", secret.GetName()),
			fields.OneTermEqualSelector(".spec.influxDB.namespace", secret.GetNamespace())),
	}
	err := r.List(ctx, attachedDevices, listOps)
	if err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(attachedDevices.Items))
	for i, item := range attachedDevices.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}
