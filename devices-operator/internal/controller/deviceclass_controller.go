package controller

import (
	"context"
	"fmt"

	pedgev1alpha1 "github.com/plmercereau/pedge/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// DeviceClassReconciler reconciles a DeviceClass object
type DeviceClassReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=devices.pedge.io,resources=deviceclasses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devices.pedge.io,resources=deviceclasses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete

func (r *DeviceClassReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the DeviceClass instance
	deviceClass := &pedgev1alpha1.DeviceClass{}
	err := r.Get(ctx, req.NamespacedName, deviceClass)

	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Unable to fetch DeviceClass")
		return ctrl.Result{}, err
	}

	if deviceClass.Spec.Builder != (pedgev1alpha1.FirmwareBuilder{}) {
		logger.Info("Found firmware builder. Generating job...")
		jobName := deviceClass.Name + "-firmware-build"
		builderImage := deviceClass.Spec.Builder.Image

		service := &corev1.Service{}
		// TODO that's a tricky one, we need to link the minio service to the firmware job
		err := r.Get(ctx, types.NamespacedName{Name: "devicesCluster.Name", Namespace: "devicesCluster.Namespace"}, service)
		// ! err := r.Get(ctx, types.NamespacedName{Name: devicesCluster.Name, Namespace: devicesCluster.Namespace}, service)
		if err != nil {
			logger.Error(err, "unable to fetch service")
			return ctrl.Result{}, err
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
		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobName,
				Namespace: deviceClass.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(deviceClass, pedgev1alpha1.GroupVersion.WithKind("DeviceClass")),
				},
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					// By setting the annotation, we make sure that the Job is re-triggered when the deviceClass changes
					// Re-trigger the job when the deviceClass, secret changes or service changes
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							// ??? We don't need annotate changes in the firmare as for now only the builder image can change
							serviceAnnotation: fmt.Sprintf("%s:%s", mqttBroker, mqttPort),
						},
					},
					Spec: corev1.PodSpec{
						InitContainers: []corev1.Container{
							{
								Name:            "build-firmware",
								Image:           fmt.Sprintf("%s/%s:%s", builderImage.Registry, builderImage.Repository, builderImage.Tag),
								ImagePullPolicy: builderImage.PullPolicy,
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "firmware",
										MountPath: "/firmware",
									},
								},
								Env: []corev1.EnvVar{},
							},
						},
						Containers: []corev1.Container{
							{
								Name:  "upload-firmware",
								Image: "amazon/aws-cli:2.16.6",
								// TODO get "firmware" bucket name and credentials
								Command: []string{"sh", "-c", fmt.Sprintf("aws --no-verify-ssl s3 cp /firmware/firmware.bin s3://%s/%s.bin", "firmware", deviceClass.Name)},
								Env:     []corev1.EnvVar{
									// {
									// 	Name: "AWS_ENDPOINT_URL_S3",
									// 	// TODO that's a tricky one, we need to link the minio service to the firmware job
									// 	Value: fmt.Sprintf("https://minio.%s.svc.cluster.local", devicesCluster.Namespace),
									// },
									// {
									// 	Name:  "AWS_DEFAULT_REGION",
									// 	Value: bucketRegion,
									// },
									// {
									// 	Name: "AWS_ACCESS_KEY_ID",
									// 	ValueFrom: &corev1.EnvVarSource{
									// 		SecretKeyRef: &corev1.SecretKeySelector{
									// 			LocalObjectReference: corev1.LocalObjectReference{
									// 				Name: devicesCluster.Name + devicesClusterSecretSuffix,
									// 			},
									// 			Key: s3AccessKeyId,
									// 		},
									// 	},
									// },
									// {
									// 	Name: "AWS_SECRET_ACCESS_KEY",
									// 	ValueFrom: &corev1.EnvVarSource{
									// 		SecretKeyRef: &corev1.SecretKeySelector{
									// 			LocalObjectReference: corev1.LocalObjectReference{
									// 				Name: devicesCluster.Name + devicesClusterSecretSuffix,
									// 			},
									// 			Key: s3SecretAccessKey,
									// 		},
									// 	},
									// },
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
		// Set the ownerRef for the Job to ensure it gets cleaned up when the deviceClass is deleted
		if err := ctrl.SetControllerReference(deviceClass, job, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.CreateOrUpdate(ctx, job); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// CreateOrUpdate creates or updates a resource
func (r *DeviceClassReconciler) CreateOrUpdate(ctx context.Context, obj client.Object) error {
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
func (r *DeviceClassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pedgev1alpha1.DeviceClass{}).
		Complete(r)
}
