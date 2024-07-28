package controller

import (
	"context"
	"fmt"

	"time"

	rabbitmqtopologyv1 "github.com/rabbitmq/messaging-topology-operator/api/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	errorutils "k8s.io/apimachinery/pkg/util/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	// "sigs.k8s.io/controller-runtime/pkg/source"
	pedgev1alpha1 "github.com/plmercereau/pedge/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// The suffix should not change: the rabbitmq operator takes ownership of it,
	// and still creates a -user-credentials secret even when asked otherwise. Investigate.
	deviceSecretSuffix     = "-user-credentials"
	configBuilderJobSuffix = "-config-build"
	configJobLabel         = "pedge.io/config-job"
	deviceClusterLabel     = "pedge.io/device-cluster"

	hashForDeviceAnnotation     = "pedge.io/device-hash"
	secretHashAnnotation        = "pedge.io/secret-hash"
	deviceClassHashAnnotation   = "pedge.io/device-class-hash"
	deviceClusterHashAnnotation = "pedge.io/device-cluster-hash"
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

	device := &pedgev1alpha1.Device{}
	if err := r.Get(ctx, req.NamespacedName, device); err != nil {
		if apierrors.IsNotFound(err) {
			logger.V(5).Info("Object was not found, not an error")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, fmt.Errorf("failed to get device object: %w", err)
	}

	// Collect errors as an aggregate to return together after all patches have been performed.
	var errs []error

	patchBase := client.MergeFrom(device.DeepCopy())
	deviceStatusCopy := device.Status.DeepCopy() // Patch call will erase the status

	result, err := r.reconcile(ctx, device)
	if err != nil {
		errs = append(errs, fmt.Errorf("error reconciling device object: %w", err))
	}

	if err := r.Patch(ctx, device, patchBase); err != nil && !apierrors.IsNotFound(err) {
		errs = append(errs, fmt.Errorf("failed to patch device object: %w", err))
	}

	device.Status = *deviceStatusCopy

	if err := r.Status().Patch(ctx, device, patchBase); err != nil && !apierrors.IsNotFound(err) {
		errs = append(errs, fmt.Errorf("failed to patch status for device object: %w", err))
	}

	if len(errs) > 0 {
		return ctrl.Result{}, errorutils.NewAggregate(errs)
	}

	return result, nil
}

func (r *DeviceReconciler) reconcile(ctx context.Context, device *pedgev1alpha1.Device) (ctrl.Result, error) {

	deviceClass, err := r.loadDeviceClass(ctx, device)
	if err != nil {
		return ctrl.Result{}, err
	}

	deviceCluster, err := r.loadDeviceCluster(ctx, device, deviceClass)
	if err != nil {
		return ctrl.Result{}, err
	}

	secret, err := r.createSecret(ctx, device)
	if err != nil {
		return ctrl.Result{}, err
	}

	if result, err := r.createJob(ctx, device, secret, deviceClass, deviceCluster); err != nil {
		return ctrl.Result{}, err
	} else if result != (ctrl.Result{}) {
		return result, nil
	}

	if _, err := r.createRabbitmqResources(ctx, device, deviceCluster); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *DeviceReconciler) createSecret(ctx context.Context, device *pedgev1alpha1.Device) (*corev1.Secret, error) {
	logger := ctrl.LoggerFrom(ctx)
	// Secret
	secretName := device.Name + deviceSecretSuffix
	var secret corev1.Secret

	if err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: device.Namespace}, &secret); err != nil {
		logger.Info("Creating a new secret " + secretName)
		data := map[string][]byte{
			"password": []byte(generateRandomPassword(16)),
			"username": []byte(device.Name),
		}
		secret = corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: device.Namespace,
				Annotations: map[string]string{
					hashForDeviceAnnotation: hashByteData(data),
				},
			},

			Data: data,
		}
		// ? Set the ownerRef for the secret to ensure it gets cleaned up when the device is deleted
		// if err := ctrl.SetControllerReference(device, secret, r.Scheme); err != nil {
		// 	return err
		// }

		if err := r.Create(ctx, &secret); err != nil {
			logger.Error(err, "Unable to create secret "+secretName)
			return nil, err
		}
	} else {
		patch := client.MergeFrom(secret.DeepCopy())
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		if _, exists := secret.Data["password"]; !exists {
			secret.Data["password"] = []byte(generateRandomPassword(16))
		}
		if _, exists := secret.Data["username"]; !exists || string(secret.Data["username"]) != device.Name {
			secret.Data["username"] = []byte(device.Name)
		}
		if secret.Annotations == nil {
			secret.Annotations = make(map[string]string)
		}
		if secret.Annotations == nil {
			secret.Annotations = make(map[string]string)
		}
		secret.Annotations[hashForDeviceAnnotation] = hashByteData(secret.Data)
		// patch the secret with the new labels - not using update but patch to avoid conflicts
		if err := r.Patch(ctx, &secret, patch); err != nil {
			logger.Error(err, "Unable to patch secret "+secretName)
			return nil, err
		}
	}

	if device.Labels == nil {
		device.Labels = make(map[string]string)
	}
	device.Labels[secretNameLabel] = device.Name + deviceSecretSuffix

	return &secret, nil
}

func (r *DeviceReconciler) loadDeviceClass(ctx context.Context, device *pedgev1alpha1.Device) (*pedgev1alpha1.DeviceClass, error) {
	logger := log.FromContext(ctx)
	deviceClass := &pedgev1alpha1.DeviceClass{}
	if err := r.Get(ctx, types.NamespacedName{Name: device.Spec.DeviceClassReference.Name, Namespace: device.Namespace}, deviceClass); err != nil {
		logger.Error(err, "unable to fetch deviceClass")
		return nil, err
	}
	return deviceClass, nil
}

func (r *DeviceReconciler) loadDeviceCluster(ctx context.Context, device *pedgev1alpha1.Device, deviceClass *pedgev1alpha1.DeviceClass) (*pedgev1alpha1.DeviceCluster, error) {
	logger := ctrl.LoggerFrom(ctx)
	deviceCluster := &pedgev1alpha1.DeviceCluster{}
	if err := r.Get(ctx, types.NamespacedName{Name: deviceClass.Spec.DeviceClusterReference.Name, Namespace: device.Namespace}, deviceCluster); err != nil {
		logger.Error(err, "unable to fetch DeviceCluster")
		return nil, err
	}
	if device.Labels == nil {
		device.Labels = make(map[string]string)
	}
	device.Labels[deviceClusterLabel] = deviceCluster.Name
	return deviceCluster, nil
}

func (r *DeviceReconciler) createJob(ctx context.Context, device *pedgev1alpha1.Device, secret *corev1.Secret, deviceClass *pedgev1alpha1.DeviceClass, deviceCluster *pedgev1alpha1.DeviceCluster) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)
	jobName := device.Name + configBuilderJobSuffix
	existingJob := &batchv1.Job{}
	err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: device.Namespace}, existingJob)
	jobExists := false
	if err == nil {
		// Job exists, check if it is being deleted
		if existingJob.GetDeletionTimestamp() != nil {
			// Job is being deleted, requeue after some time
			logger.Info("Job is being deleted, requeuing", "job", jobName)
			return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		}
		// Requeue if the job is not completed yet (success or failure)
		if existingJob.Status.CompletionTime == nil {
			logger.Info("Job is not completed yet, requeuing", "job", jobName)
			return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		}
		jobExists = true

	} else {
		if !apierrors.IsNotFound(err) {
			logger.Error(err, "unable to fetch existing Job")
			return ctrl.Result{}, err
		}
	}

	if jobExists {
		// * Here we decide if we need to recreate the job, depending on the changes in the following resources through their hashes:
		// 1. secret hash
		// 2. device class hash
		// 3. device cluster hahs
		// TODO 4. DeviceGroup is different
		deviceAnnotations := device.Annotations
		if deviceAnnotations == nil {
			deviceAnnotations = make(map[string]string)
		}
		secretAnnotations := secret.Annotations
		if secretAnnotations == nil {
			secretAnnotations = make(map[string]string)
		}
		secretHash := secretAnnotations[hashForDeviceAnnotation]
		deviceClassAnnotations := deviceClass.Annotations
		if deviceClassAnnotations == nil {
			deviceClassAnnotations = make(map[string]string)
		}
		deviceClassHash := deviceClassAnnotations[hashForDeviceAnnotation]
		deviceClusterAnnotations := deviceCluster.Annotations
		if deviceClusterAnnotations == nil {
			deviceClusterAnnotations = make(map[string]string)
		}
		deviceClusterHash := deviceClusterAnnotations[hashForDeviceAnnotation]
		if (deviceAnnotations[secretHashAnnotation] != secretHash) || (deviceAnnotations[deviceClassHashAnnotation] != deviceClassHash) || (deviceAnnotations[deviceClusterHashAnnotation] != deviceClusterHash) {
			// Delete the existing Job
			if err := r.Delete(ctx, existingJob, client.PropagationPolicy(metav1.DeletePropagationForeground)); err != nil {
				logger.Error(err, "unable to delete existing Job", "job", jobName)
				return ctrl.Result{}, err
			}

			// Requeue until the job is deleted
			logger.Info("Existing config job deleted, requeuing for new job creation", "job", jobName)
			return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		} else {
			logger.Info("Config job already exists and is up to date", "job", jobName)
			return ctrl.Result{}, nil
		}
	}

	configBuilderImage := deviceClass.Spec.Config.Image
	outputMount := corev1.VolumeMount{
		Name:      "output",
		MountPath: "/output",
	}
	secretsMount := corev1.VolumeMount{
		Name:      "secrets",
		MountPath: "/secrets",
		ReadOnly:  true,
	}
	storageMount := corev1.VolumeMount{
		Name:      "data",
		MountPath: "/data",
	}

	// ! The template spec is immutable
	jobTemplate := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				configJobLabel: jobName,
			},
		},
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{
				{
					Name:            "build-config",
					Image:           fmt.Sprintf("%s:%s", configBuilderImage.Repository, configBuilderImage.Tag),
					ImagePullPolicy: configBuilderImage.PullPolicy,
					VolumeMounts: []corev1.VolumeMount{
						secretsMount,
						outputMount,
					},
					Env: []corev1.EnvVar{
						{
							Name:  "MQTT_BROKER",
							Value: deviceCluster.GetAnnotations()[mqttBrokerHostnameAnnotation],
						},
						{
							Name:  "MQTT_PORT",
							Value: deviceCluster.GetAnnotations()[mqttBrokerPortAnnotation],
						},
						{
							Name:  "MQTT_USERNAME",
							Value: device.Name,
						},
						{
							Name: "MQTT_PASSWORD",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: device.Name + deviceSecretSuffix,
									},
									Key: "password",
								},
							},
						},
						{
							Name:  "MQTT_SENSORS_TOPIC",
							Value: deviceCluster.Spec.MQTT.SensorsTopic,
						},
					},
				},
				{
					Name:  "store-hash",
					Image: "leplusorg/hash:sha-657c8c8",
					Command: []string{"sh", "-c", fmt.Sprintf(`mkdir -p %s/auth; echo -n "$PASSWORD" | argon2 saltItWithSalt -id -e > %s/auth/%s.hash`,
						storageMount.MountPath,
						storageMount.MountPath,
						device.Name)},
					VolumeMounts: []corev1.VolumeMount{
						storageMount,
					},
					Env: []corev1.EnvVar{
						{
							Name: "PASSWORD",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: device.Name + deviceSecretSuffix,
									},
									Key: "password",
								},
							},
						},
					},
				},
			},
			Containers: []corev1.Container{
				// Use a persistent volume to store the firmware
				{
					Name:  "upload-secret",
					Image: "busybox:1.36.1",
					Command: []string{"sh", "-c", fmt.Sprintf("mkdir -p %s/configurations/%s; tar -czf %s/configurations/%s/config.tgz -C %s .",
						storageMount.MountPath,
						device.Name,
						storageMount.MountPath,
						device.Name,
						outputMount.MountPath)},
					VolumeMounts: []corev1.VolumeMount{
						outputMount,
						storageMount,
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
			Volumes: []corev1.Volume{
				{
					Name: outputMount.Name,
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: storageMount.Name,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: deviceCluster.Name + artefactSuffix,
						},
					},
				},
				{
					Name: secretsMount.Name,
					VolumeSource: corev1.VolumeSource{
						Projected: &corev1.ProjectedVolumeSource{
							Sources: []corev1.VolumeProjection{
								{
									Secret: &corev1.SecretProjection{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: deviceCluster.Name + deviceClusterSecretSuffix,
										},
										Optional: func(b bool) *bool { return &b }(true),
									},
								},
								{
									Secret: &corev1.SecretProjection{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: device.Name + deviceSecretSuffix,
										},
										Optional: func(b bool) *bool { return &b }(true),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// * Set the new hashes to compare in the next reconciliations
	if device.Annotations == nil {
		device.Annotations = make(map[string]string)
	}
	device.Annotations[secretHashAnnotation] = secret.GetAnnotations()[hashForDeviceAnnotation]
	device.Annotations[deviceClassHashAnnotation] = deviceClass.GetAnnotations()[hashForDeviceAnnotation]
	device.Annotations[deviceClusterHashAnnotation] = deviceCluster.GetAnnotations()[hashForDeviceAnnotation]
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: device.Namespace,
			Labels: map[string]string{
				configJobLabel: jobName,
			},
		},
		Spec: batchv1.JobSpec{
			Template:     jobTemplate,
			BackoffLimit: func(i int32) *int32 { return &i }(4),
		},
	}

	// Set the ownerRef for the Job to ensure it gets cleaned up when the device is deleted
	if err := ctrl.SetControllerReference(device, job, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.Create(ctx, job); err != nil {
		return ctrl.Result{}, err
	}
	logger.Info("Job created", "job", job.Name)
	return ctrl.Result{}, nil
}

func (r *DeviceReconciler) createRabbitmqResources(ctx context.Context, device *pedgev1alpha1.Device, deviceCluster *pedgev1alpha1.DeviceCluster) (ctrl.Result, error) {
	// Define the desired RabbitMQ User resource
	user := &rabbitmqtopologyv1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name:      device.Name,
			Namespace: device.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(device, pedgev1alpha1.GroupVersion.WithKind("Device")),
			},
		},
		Spec: rabbitmqtopologyv1.UserSpec{
			RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
				Name: deviceCluster.Name,
			},
			ImportCredentialsSecret: &corev1.LocalObjectReference{
				Name: device.Name + deviceSecretSuffix,
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
				Name:      deviceCluster.Name,
				Namespace: deviceCluster.Namespace,
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
				Write:    fmt.Sprintf("^%s\\.%s\\..+$", deviceCluster.Spec.MQTT.SensorsTopic, device.Name),
				// TODO narrow down the read permissions: the device should only be able to write to its own topic
				Read: fmt.Sprintf("^%s\\.%s\\..+$", deviceCluster.Spec.MQTT.SensorsTopic, device.Name),
			},
			RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
				Name:      deviceCluster.Name,
				Namespace: deviceCluster.Namespace,
			},
		},
	}

	resources := []client.Object{user, permission, topicPermission}
	for _, res := range resources {
		if err := r.CreateOrUpdate(ctx, res); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil

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
		if apierrors.IsConflict(err) {
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
		Owns(&corev1.Secret{}).
		Owns(&batchv1.Job{}).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.mapSecretToDevice),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Watches(
			&pedgev1alpha1.DeviceClass{},
			handler.EnqueueRequestsFromMapFunc(r.mapDeviceClassToDevice),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		// Watch for changes in the DeviceCluster as it may affect the device e.g. when the MQTT broker or the artefact ingress changes
		Watches(
			&pedgev1alpha1.DeviceCluster{},
			handler.EnqueueRequestsFromMapFunc(r.mapDeviceClusterToDevice),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}

func (r *DeviceReconciler) mapSecretToDevice(ctx context.Context, secret client.Object) []reconcile.Request {
	// Fetch the list of Devices that use this secret
	var devices pedgev1alpha1.DeviceList
	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{secretNameLabel: secret.GetName()}}
	labelMap, _ := metav1.LabelSelectorAsMap(&labelSelector)
	listOps := &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelMap),
		Namespace:     secret.GetNamespace(),
	}
	if err := r.List(context.Background(), &devices, listOps); err != nil {
		// Handle error
		return nil
	}

	// Create reconcile requests for each device
	var requests []reconcile.Request
	for _, device := range devices.Items {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      device.Name,
				Namespace: device.Namespace,
			},
		})
	}
	return requests
}

func (r *DeviceReconciler) mapDeviceClassToDevice(ctx context.Context, deviceClass client.Object) []reconcile.Request {
	devices := &pedgev1alpha1.DeviceList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("spec.deviceClassReference.name", deviceClass.GetName()),
		Namespace:     deviceClass.GetNamespace(),
	}
	if err := r.List(ctx, devices, listOps); err != nil {
		return []reconcile.Request{}
	}

	// Create reconcile requests for each device
	var requests []reconcile.Request
	for _, device := range devices.Items {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      device.Name,
				Namespace: device.Namespace,
			},
		})
	}
	return requests
}

func (r *DeviceReconciler) mapDeviceClusterToDevice(ctx context.Context, cluster client.Object) []reconcile.Request {
	devices := &pedgev1alpha1.DeviceList{}
	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{deviceClusterLabel: cluster.GetName()}}
	labelMap, _ := metav1.LabelSelectorAsMap(&labelSelector)
	listOps := &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelMap),
		Namespace:     cluster.GetNamespace(),
	}
	if err := r.List(ctx, devices, listOps); err != nil {
		return []reconcile.Request{}
	}

	// Create reconcile requests for each device
	var requests []reconcile.Request
	for _, device := range devices.Items {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      device.Name,
				Namespace: device.Namespace,
			},
		})
	}
	return requests

}
