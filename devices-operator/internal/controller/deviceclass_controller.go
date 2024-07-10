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
	if err := r.Get(ctx, req.NamespacedName, deviceClass); err != nil {
		logger.Error(err, "Unable to fetch DeviceClass")
		return ctrl.Result{}, err
	}

	jobName := deviceClass.Name + "-firmware-build"
	builderImage := deviceClass.Spec.Builder.Image
	firmwareMount := corev1.VolumeMount{
		Name:      "firmware",
		MountPath: "/firmware",
	}
	storageMount := corev1.VolumeMount{
		Name:      "data",
		MountPath: "/data",
		SubPath:   "firmwares",
	}
	desiredJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: deviceClass.Namespace,
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
					InitContainers: []corev1.Container{
						{
							Name:            "build-firmware",
							Image:           fmt.Sprintf("%s:%s", builderImage.Repository, builderImage.Tag),
							ImagePullPolicy: builderImage.PullPolicy,
							VolumeMounts:    []corev1.VolumeMount{firmwareMount},
							Env:             deviceClass.Spec.Builder.Env,
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "upload-firmware",
							Image: "busybox:1.36.1",
							Command: []string{"sh", "-c", fmt.Sprintf("mkdir -p %s/%s; tar -czf %s/%s/firmware.tgz %s/*",
								storageMount.MountPath,
								deviceClass.Name,
								storageMount.MountPath,
								deviceClass.Name,
								firmwareMount.MountPath)},
							VolumeMounts: []corev1.VolumeMount{firmwareMount, storageMount},
						},
					},
					RestartPolicy: corev1.RestartPolicyOnFailure,
					Volumes: []corev1.Volume{
						{
							Name: firmwareMount.Name,
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: storageMount.Name,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: deviceClass.Spec.DevicesClusterReference.Name + persistentVolumeClaimSuffix,
								},
							},
						},
					},
				},
			},
			BackoffLimit: func(i int32) *int32 { return &i }(4),
		},
	}

	// Set the ownerRef for the Job to ensure it gets cleaned up when the deviceClass is deleted
	if err := ctrl.SetControllerReference(deviceClass, desiredJob, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Check if the Job already exists
	existingJob := &batchv1.Job{}
	if err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: deviceClass.Namespace}, existingJob); err != nil {
		if errors.IsNotFound(err) {
			// Create the Job if it does not exist
			logger.Info("Creating new Job ", "job", jobName)
			if err := r.Create(ctx, desiredJob); err != nil {
				logger.Error(err, "Failed to create new Job", "job", jobName)
				return ctrl.Result{}, err
			}
		} else {
			logger.Error(err, "Failed to get Job", "job", jobName)
			return ctrl.Result{}, err
		}
	} else {
		// Compare the existing Job with the desired Job
		if !jobSpecMatches(existingJob, desiredJob) {
			// If the Job specs differ, delete the existing Job and create a new one
			logger.Info("Deleting existing Job", "job", jobName)
			if err := r.Delete(ctx, existingJob); err != nil {
				logger.Error(err, "Failed to delete existing Job", "job", jobName)
				return ctrl.Result{}, err
			}
			logger.Info("Creating new Job", "job", jobName)
			if err := r.Create(ctx, desiredJob); err != nil {
				logger.Error(err, "Failed to create new Job", "job", jobName)
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *DeviceClassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pedgev1alpha1.DeviceClass{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}
