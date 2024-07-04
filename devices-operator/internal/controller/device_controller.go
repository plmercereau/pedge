package controller

import (
	"context"
	"fmt"

	rabbitmqtopologyv1 "github.com/rabbitmq/messaging-topology-operator/api/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	// "sigs.k8s.io/controller-runtime/pkg/source"
	pedgev1alpha1 "github.com/plmercereau/pedge/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	// The suffix should not change: the rabbitmq operator takes ownership of it,
	// and still creates a -user-credentials secret even when asked otherwise. Investigate.
	deviceSecretSuffix      = "-user-credentials"
	secretNameLabel         = "pedge.io/secret-name"
	secretVersionAnnotation = "pedge.io/secret-version"
	serviceAnnotation       = "pedge.io/service"
)

// DeviceReconciler reconciles a Device object
type DeviceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=devices.pedge.io,resources=devices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devices.pedge.io,resources=devices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rabbitmq.com,resources=permissions;topicpermissions;users,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch

func (r *DeviceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Device")
	// Fetch the Device instance
	device := &pedgev1alpha1.Device{}
	if err := r.Get(ctx, req.NamespacedName, device); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "unable to fetch device")
		return ctrl.Result{}, err
	}
	logger.Info("Fetched device " + device.Name)

	// Sync resources
	if err := r.syncResources(ctx, device); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// syncResources creates or updates the associated resources
func (r *DeviceReconciler) syncResources(ctx context.Context, device *pedgev1alpha1.Device) error {
	logger := log.FromContext(ctx)

	devicesCluster := &pedgev1alpha1.DevicesCluster{}
	if err := r.Get(ctx, types.NamespacedName{Name: device.Spec.DevicesClusterReference.Name, Namespace: device.Namespace}, devicesCluster); err != nil {
		logger.Error(err, "unable to fetch DevicesCluster")
		return err
	}
	// Secret
	secretName := device.Name + deviceSecretSuffix
	var secret corev1.Secret
	var secretVersion string
	if err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: device.Namespace}, &secret); err != nil {
		logger.Info("Creating new secret " + secretName)
		secret = corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: device.Namespace,
			},
			Data: map[string][]byte{
				"username": []byte(device.Name),
				"password": []byte(generateRandomPassword(16)),
			},
		}
		// Set the ownerRef for the secret to ensure it gets cleaned up when the device is deleted
		// if err := ctrl.SetControllerReference(device, secret, r.Scheme); err != nil {
		// 	return err
		// }

		if err := r.Create(ctx, &secret); err != nil {
			logger.Error(err, "Unable to create secret "+secretName)
			return err
		}
	} else {
		changed := false
		username := string(secret.Data["username"])
		if string(secret.Data["username"]) != device.Name {
			logger.Info("Mismatching username in " + secretName + ". Expected " + device.Name + " but found " + username + ". Overriding the value")
			secret.Data["username"] = []byte(device.Name)
			changed = true
		}
		if changed {
			if err := r.Update(ctx, &secret); err != nil {
				logger.Error(err, "Unable to update secret "+secretName)
				return err
			}
		}
	}
	secretVersion = secret.ResourceVersion

	// Check if labels map is nil
	if device.Labels == nil {
		device.Labels = make(map[string]string)
	}

	// Add the secret name to the device labels
	device.Labels[secretNameLabel] = secretName

	// Update the Device resource
	if err := r.Client.Update(ctx, device); err != nil {
		return err
	}

	// Define the desired RabbitMQ User resource
	user := &rabbitmqtopologyv1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name:      device.Name,
			Namespace: device.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(device, pedgev1alpha1.GroupVersion.WithKind("Device")),
			},
			Annotations: map[string]string{
				// Needed to trigger a reconciliation when the password changes
				secretVersionAnnotation: secretVersion,
			},
		},
		Spec: rabbitmqtopologyv1.UserSpec{
			RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
				Name: devicesCluster.Name,
			},
			ImportCredentialsSecret: &corev1.LocalObjectReference{
				Name: secretName,
			},
		},
	}

	// Define the desired RabbitMQ Permission resource
	permission := &rabbitmqtopologyv1.Permission{
		ObjectMeta: metav1.ObjectMeta{
			Name:      device.Name,
			Namespace: device.Namespace,
			// ? RabbitMQ does not want to loose ownership of the resource
			// OwnerReferences: []metav1.OwnerReference{
			// 	*metav1.NewControllerRef(device, pedgev1alpha1.GroupVersion.WithKind("Device")),
			// },
		},
		Spec: rabbitmqtopologyv1.PermissionSpec{
			Vhost: "/",
			UserReference: &corev1.LocalObjectReference{
				Name: device.Name,
			},
			Permissions: rabbitmqtopologyv1.VhostPermissions{
				Configure: "^mqtt-subscription-.*$",
				Write:     "^amq\\.topic$|^mqtt-subscription-.*$",
				Read:      "^amq\\.topic$|^mqtt-subscription-.*$",
			},
			RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
				Name:      devicesCluster.Name,
				Namespace: devicesCluster.Namespace,
			},
		},
	}

	// Define the desired RabbitMQ TopicPermission resource
	topicPermission := &rabbitmqtopologyv1.TopicPermission{
		ObjectMeta: metav1.ObjectMeta{
			Name:      device.Name,
			Namespace: device.Namespace,
			// ? RabbitMQ does not want to loose ownership of the resource
			// OwnerReferences: []metav1.OwnerReference{
			// 	*metav1.NewControllerRef(device, pedgev1alpha1.GroupVersion.WithKind("Device")),
			// },
		},
		Spec: rabbitmqtopologyv1.TopicPermissionSpec{
			Vhost: "/",
			UserReference: &corev1.LocalObjectReference{
				Name: device.Name,
			},
			Permissions: rabbitmqtopologyv1.TopicPermissionConfig{
				Exchange: "amq.topic",
				Write:    fmt.Sprintf("^%s\\.%s\\..+$", devicesCluster.Spec.MQTT.SensorsTopic, device.Name),
				// TODO narrow down the read permissions: the device should only be able to write to its own topic
				Read: fmt.Sprintf("^%s\\.%s\\..+$", devicesCluster.Spec.MQTT.SensorsTopic, device.Name),
			},
			RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
				Name:      devicesCluster.Name,
				Namespace: devicesCluster.Namespace,
			},
		},
	}

	resources := []client.Object{user, permission, topicPermission}
	for _, res := range resources {
		if err := r.CreateOrUpdate(ctx, res); err != nil {
			return err
		}
	}

	// Fetch the referenced DeviceClass instance
	var deviceClass *pedgev1alpha1.DeviceClass
	if device.Spec.DeviceClasseReference != (pedgev1alpha1.DeviceClassReference{}) {
		deviceClass = &pedgev1alpha1.DeviceClass{}
		if err := r.Get(ctx, types.NamespacedName{Name: device.Spec.DeviceClasseReference.Name, Namespace: device.Namespace}, deviceClass); err != nil {
			logger.Error(err, "unable to fetch deviceClass")
			return err
		}
	}

	service := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: "devicesCluster.Name", Namespace: "devicesCluster.Namespace"}, service); err != nil {
		logger.Error(err, "unable to fetch service")
		return err
	}
	// TODO check if the service is a LoadBalancer and has an IP. Warn if more than one IP
	// TODO use possible custom broker/port values in the DevicesCluster spec
	serviceIngress := service.Status.LoadBalancer.Ingress[0]
	var mqttBroker string
	if serviceIngress.Hostname != "" {
		mqttBroker = serviceIngress.Hostname
	} else {
		mqttBroker = serviceIngress.IP
	}
	mqttPortInt := -1
	for _, port := range service.Spec.Ports {
		if port.Name == "mqtt" {
			mqttPortInt = int(port.Port)
			break
		}
	}
	mqttPort := fmt.Sprint(mqttPortInt)

	jobName := device.Name + "-config-build"
	configBuilderImage := deviceClass.Spec.Config.Image

	desiredJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: device.Namespace,
			Labels: map[string]string{
				"job-name": jobName,
			},
		},
		Spec: batchv1.JobSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"job-name": jobName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"job-name": jobName,
					},
				},
				Spec: corev1.PodSpec{
					// The service account is used to be able to patch the secret. It is defined by the DevicesCluster
					ServiceAccountName: devicesCluster.Name + "-config-builder",
					InitContainers: []corev1.Container{
						{
							Name:            "build-config",
							Image:           fmt.Sprintf("%s:%s", configBuilderImage.Repository, configBuilderImage.Tag),
							ImagePullPolicy: configBuilderImage.PullPolicy,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "secrets",
									MountPath: "/secrets",
								},
								{
									Name:      "output",
									MountPath: "/output",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "MQTT_BROKER",
									Value: mqttBroker,
								},
								{
									Name:  "MQTT_PORT",
									Value: mqttPort,
								},
								{
									Name:  "MQTT_USERNAME",
									Value: device.Name,
								},
								{
									Name:  "MQTT_PASSWORD",
									Value: string(secret.Data["password"]),
								},
								{
									Name:  "SENSORS_TOPIC",
									Value: devicesCluster.Spec.MQTT.SensorsTopic,
								},
								{
									Name:      "AWS_SECRET_ACCESS_KEY",
									ValueFrom: deviceClass.Spec.Storage.SecretKey,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:    "update-secret",
							Image:   "bitnami/kubectl:1.30.2",
							Command: []string{"sh", "-c", fmt.Sprintf(`kubectl patch secret %s -n %s --type merge -p '{"data":{"config.bin":"'$(base64 -w0 /output/config.bin)'"}}'`, secretName, device.Namespace)},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "output",
									MountPath: "/output",
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyOnFailure,
					Volumes: []corev1.Volume{
						{
							Name: "secrets",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: secretName,
									Optional:   func(b bool) *bool { return &b }(true),
								},
							},
						},
						{
							Name: "output",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
			BackoffLimit: func(i int32) *int32 { return &i }(4),
		},
	}

	// Set the ownerRef for the Job to ensure it gets cleaned up when the device is deleted
	if err := ctrl.SetControllerReference(device, desiredJob, r.Scheme); err != nil {
		return err
	}

	// Check if the Job already exists
	existingJob := &batchv1.Job{}
	err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: device.Namespace}, existingJob)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create the Job if it does not exist
			logger.Info("Creating new Job ", "job", jobName)
			if err := r.Create(ctx, desiredJob); err != nil {
				logger.Error(err, "Failed to create new Job", "job", jobName)
				return err
			}
		} else {
			logger.Error(err, "Failed to get Job", "job", jobName)
			return err
		}
	} else {
		// Compare the existing Job with the desired Job
		if !jobSpecMatches(existingJob, desiredJob) {
			// If the Job specs differ, delete the existing Job and create a new one
			logger.Info("Deleting existing Job", "job", jobName)
			if err := r.Delete(ctx, existingJob); err != nil {
				logger.Error(err, "Failed to delete existing Job", "job", jobName)
				return err
			}
			logger.Info("Creating new Job", "job", jobName)
			if err := r.Create(ctx, desiredJob); err != nil {
				logger.Error(err, "Failed to create new Job", "job", jobName)
				return err
			}
		}
	}

	return nil
}

// CreateOrUpdate creates or updates a resource
func (r *DeviceReconciler) CreateOrUpdate(ctx context.Context, obj client.Object) error {
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

// * https://book.kubebuilder.io/reference/watching-resources/externally-managed
func (r *DeviceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pedgev1alpha1.Device{}).
		Owns(&batchv1.Job{}).
		// TODO Watch DevicesCluster, too (the sensors topic name can change)
		// Watch for changes in the secret of the device
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForSecret),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		// Watch for changes in the class of the device
		Watches(
			&pedgev1alpha1.DeviceClass{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForDeviceClass),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		// Watch for changes in the service exposing the devices cluster
		Watches(
			&corev1.Service{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForService),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}

func (r *DeviceReconciler) findObjectsForSecret(ctx context.Context, secret client.Object) []reconcile.Request {
	attachedDevices := &pedgev1alpha1.DeviceList{}
	listOps := &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{
			secretNameLabel: secret.GetName(),
		}),
		Namespace: secret.GetNamespace(),
	}
	if err := r.List(ctx, attachedDevices, listOps); err != nil {
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

func (r *DeviceReconciler) findObjectsForDeviceClass(ctx context.Context, deviceClass client.Object) []reconcile.Request {
	attachedDevices := &pedgev1alpha1.DeviceList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(".spec.deviceClassReference.name", deviceClass.GetName()),
		Namespace:     deviceClass.GetNamespace(),
	}
	if err := r.List(ctx, attachedDevices, listOps); err != nil {
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

func (r *DeviceReconciler) findObjectsForService(ctx context.Context, service client.Object) []reconcile.Request {
	attachedDevices := &pedgev1alpha1.DeviceList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(".metadata.name", service.GetName()),
		Namespace:     service.GetNamespace(),
	}
	if err := r.List(ctx, attachedDevices, listOps); err != nil {
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
