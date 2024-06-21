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
	// TODO try change the suffix now that we are "managing" the secret, and as the rabbitmq operator takes ownership of it
	deviceSecretSuffix        = "-user-credentials"
	secretNameLabel           = "pedge.io/secret-name"
	secretVersionAnnotation   = "pedge.io/secret-version"
	firmwareVersionAnnotation = "pedge.io/firmware-version"
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

func (r *DeviceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Device")
	// Fetch the Device instance
	device := &pedgev1alpha1.Device{}
	err := r.Get(ctx, req.NamespacedName, device)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "unable to fetch device")
		return ctrl.Result{}, err
	}
	logger.Info("Fetched device " + device.Name)
	// Fetch the referenced Firmware instance
	var firmware *pedgev1alpha1.Firmware
	if device.Spec.FirmwareReference != (pedgev1alpha1.FirmwareReference{}) {
		firmware = &pedgev1alpha1.Firmware{}
		err := r.Get(ctx, types.NamespacedName{Name: device.Spec.FirmwareReference.Name, Namespace: device.Namespace}, firmware)
		if err != nil {
			logger.Error(err, "unable to fetch firmware")
			return ctrl.Result{}, err
		}
	}

	var deviceCluster *pedgev1alpha1.DeviceCluster
	deviceCluster = &pedgev1alpha1.DeviceCluster{}
	if err := r.Get(ctx, types.NamespacedName{Name: device.Spec.DeviceClusterReference.Name, Namespace: device.Namespace}, deviceCluster); err != nil {
		logger.Error(err, "unable to fetch DeviceCluster")
		return ctrl.Result{}, err
	}

	// Sync resources
	if err := r.syncResources(ctx, device, firmware, deviceCluster); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// syncResources creates or updates the associated resources
func (r *DeviceReconciler) syncResources(ctx context.Context, device *pedgev1alpha1.Device, firmware *pedgev1alpha1.Firmware, deviceCluster *pedgev1alpha1.DeviceCluster) error {
	logger := log.FromContext(ctx)

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
				"gsm_pin":  []byte(""),
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
		if secret.Data["gsm_pin"] == nil {
			secret.Data["gsm_pin"] = []byte("")
			changed = true
		}
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
				Name: deviceCluster.Name,
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
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(device, pedgev1alpha1.GroupVersion.WithKind("Device")),
			},
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
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(device, pedgev1alpha1.GroupVersion.WithKind("Device")),
			},
		},
		Spec: rabbitmqtopologyv1.TopicPermissionSpec{
			Vhost: "/",
			UserReference: &corev1.LocalObjectReference{
				Name: device.Name,
			},
			Permissions: rabbitmqtopologyv1.TopicPermissionConfig{
				Exchange: "amq.topic",
				Write:    fmt.Sprintf("^%s\\.%s\\..+$", deviceCluster.Spec.Queue.Name, device.Name),
				// TODO narrow down the read permissions: the device should only be able to write to its own topic
				Read: fmt.Sprintf("^%s\\.%s\\..+$", deviceCluster.Spec.Queue.Name, device.Name),
			},
			RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
				Name:      deviceCluster.Name,
				Namespace: deviceCluster.Namespace,
			},
		},
	}

	resources := []client.Object{user, permission, topicPermission}

	if firmware != nil {
		jobName := device.Name + "-firmware-build"
		builderImage := firmware.Spec.Builder.Image
		job := &batchv1.Job{}
		err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: device.Namespace}, job)
		if err != nil && errors.IsNotFound(err) {
			logger.Info("Creating new firmware build job " + jobName)

			job := &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      jobName,
					Namespace: device.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(device, pedgev1alpha1.GroupVersion.WithKind("Device")),
					},
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						// TODO check this: re-trigger the job when the firmware or secret changes. It may mean we need to update the firmware first?
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								firmwareVersionAnnotation: firmware.GetResourceVersion(),
								secretVersionAnnotation:   secretVersion,
							},
						},
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Name:            "build-firmware",
									Image:           fmt.Sprintf("%s/%s:%s", builderImage.Registry, builderImage.Repository, builderImage.Tag),
									ImagePullPolicy: corev1.PullPolicy(builderImage.PullPolicy),
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "firmware",
											MountPath: "/firmware",
										},
									},
									Env: []corev1.EnvVar{
										{
											Name: "GSM_PIN",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: secretName,
													},
													Key: "gsm_pin",
												},
											},
										},
										{
											Name:  "MQTT_BROKER",
											Value: "TODO", // TODO get from the DeviceCluster / the loadbalancer
										},
										{
											Name:  "MQTT_PORT",
											Value: "1883", // TODO get from the DeviceCluster / the loadbalancer
										},
										{
											Name:  "DEVICE_NAME",
											Value: device.Name,
										},
										{
											Name: "MQTT_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: secretName,
													},
													Key: "password",
												},
											},
										},
									},
								},
							},
							Containers: []corev1.Container{
								{
									Name:    "upload-firmware",
									Image:   "amazon/aws-cli:2.16.6",
									Command: []string{"sh", "-c", fmt.Sprintf("aws --no-verify-ssl s3 cp /firmware/firmware.bin s3://%s/%s.bin", bucketName, device.Name)},
									Env: []corev1.EnvVar{
										{
											Name:  "AWS_ENDPOINT_URL_S3",
											Value: fmt.Sprintf("https://minio.%s.svc.cluster.local", deviceCluster.Namespace),
										},
										{
											Name:  "AWS_DEFAULT_REGION",
											Value: bucketRegion,
										},
										{
											Name: "AWS_ACCESS_KEY_ID",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: deviceCluster.Name + deviceClusterSecretSuffix,
													},
													Key: s3AccessKeyId,
												},
											},
										},
										{
											Name: "AWS_SECRET_ACCESS_KEY",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: deviceCluster.Name + deviceClusterSecretSuffix,
													},
													Key: s3SecretAccessKey,
												},
											},
										},
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "firmware",
											MountPath: "/firmware",
										},
									},
								},
							},
							RestartPolicy: corev1.RestartPolicyOnFailure,
							Volumes: []corev1.Volume{
								{
									Name: "firmware",
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
			if err := ctrl.SetControllerReference(device, job, r.Scheme); err != nil {
				return err
			}
			resources = append(resources, job)
		}
	}

	for _, res := range resources {
		if err := r.CreateOrUpdate(ctx, res); err != nil {
			return err
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
		// TODO Watch DeviceCluster, too
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForSecret),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Watches(
			&pedgev1alpha1.Firmware{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForFirmware),
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

func (r *DeviceReconciler) findObjectsForFirmware(ctx context.Context, firmware client.Object) []reconcile.Request {
	attachedDevices := &pedgev1alpha1.FirmwareList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(".spec.firmwareReference.name", firmware.GetName()),
		Namespace:     firmware.GetNamespace(),
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
