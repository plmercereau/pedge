package controller

import (
	"context"
	"fmt"
	"strings"

	pedgev1alpha1 "github.com/plmercereau/pedge/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	errorutils "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	deviceClusterSecretSuffix    = "-device-cluster"
	secretVersionAnnotation      = "pedge.io/secret-version"
	mqttBrokerHostnameAnnotation = "pedge.io/mqtt-broker-hostname"
	mqttBrokerPortAnnotation     = "pedge.io/mqtt-broker-port"
)

// DeviceClusterReconciler reconciles a DeviceCluster object
type DeviceClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=devices.pedge.io,resources=deviceclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devices.pedge.io,resources=deviceclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=core,resources=secrets;serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rabbitmq.com,resources=rabbitmqclusters;vhosts;queues;permissions;topicpermissions;users,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims;persistentvolumes,verbs=get;list;watch;create;update;patch;delete

func (r *DeviceClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	deviceCluster := &pedgev1alpha1.DeviceCluster{}
	if err := r.Get(ctx, req.NamespacedName, deviceCluster); err != nil {
		if apierrors.IsNotFound(err) {
			logger.V(5).Info("Object was not found, not an error")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, fmt.Errorf("failed to get device object: %w", err)
	}

	// Collect errors as an aggregate to return together after all patches have been performed.
	var errs []error

	patchBase := client.MergeFrom(deviceCluster.DeepCopy())
	deviceStatusCopy := deviceCluster.Status.DeepCopy() // Patch call will erase the status

	result, err := r.reconcile(ctx, deviceCluster)
	if err != nil {
		errs = append(errs, fmt.Errorf("error reconciling device object: %w", err))
	}

	if err := r.Patch(ctx, deviceCluster, patchBase); err != nil && !apierrors.IsNotFound(err) {
		errs = append(errs, fmt.Errorf("failed to patch device object: %w", err))
	}

	deviceCluster.Status = *deviceStatusCopy

	if err := r.Status().Patch(ctx, deviceCluster, patchBase); err != nil && !apierrors.IsNotFound(err) {
		errs = append(errs, fmt.Errorf("failed to patch status for device object: %w", err))
	}

	if len(errs) > 0 {
		return ctrl.Result{}, errorutils.NewAggregate(errs)
	}

	return result, nil

}

func (r *DeviceClusterReconciler) reconcile(ctx context.Context, deviceCluster *pedgev1alpha1.DeviceCluster) (ctrl.Result, error) {
	if err := r.ensurePersistence(ctx, deviceCluster); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.syncRabbitmqCluster(ctx, deviceCluster); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.syncHttpServer(ctx, deviceCluster); err != nil {
		return ctrl.Result{}, err
	}

	hashForDevice := map[string][]byte{
		mqttBrokerHostnameAnnotation: []byte(deviceCluster.GetAnnotations()[mqttBrokerHostnameAnnotation]),
		mqttBrokerPortAnnotation:     []byte(deviceCluster.GetAnnotations()[mqttBrokerPortAnnotation]),
		"sensors-topic":              []byte(deviceCluster.Spec.MQTT.SensorsTopic),
	}
	// Get the secret for the device cluster, so we can add it to the hash that retriggers a device config builder job
	secret := &corev1.Secret{}
	secretName := deviceCluster.Name + deviceClusterSecretSuffix
	if err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: deviceCluster.Namespace}, secret); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("failed to get secret for device cluster: %w", err)
		}
	} else {
		hashForDevice[hashForDeviceAnnotation] = []byte(hashByteData(secret.Data))
	}

	// * create a hash depending on changes that impact the device config builder
	deviceCluster.Annotations[hashForDeviceAnnotation] = hashByteData(hashForDevice)

	return ctrl.Result{}, nil
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

// SetupWithManager sets up the controller with the Manager
func (r *DeviceClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pedgev1alpha1.DeviceCluster{}).
		// Watch for changes in the service as it may impact the mqtt broker url and port
		Watches(
			&corev1.Service{},
			handler.EnqueueRequestsFromMapFunc(r.mapServiceToDeviceCluster),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.mapSecretToDeviceCluster),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}

func (r *DeviceClusterReconciler) mapServiceToDeviceCluster(ctx context.Context, service client.Object) []reconcile.Request {
	cluster := &pedgev1alpha1.DeviceCluster{}
	err := r.Get(ctx, types.NamespacedName{Name: service.GetName(), Namespace: service.GetNamespace()}, cluster)
	if err != nil {
		return []reconcile.Request{}
	}
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
			},
		},
	}
}

// Watch logic for the secret related to the device cluster
func (r *DeviceClusterReconciler) mapSecretToDeviceCluster(ctx context.Context, secret client.Object) []reconcile.Request {
	// skip if the secret does not end with deviceClusterSecretSuffix
	if !strings.HasSuffix(secret.GetName(), deviceClusterSecretSuffix) {
		return []reconcile.Request{}
	}
	// strip the suffix to get the device cluster name
	deviceClusterName := strings.TrimSuffix(secret.GetName(), deviceClusterSecretSuffix)
	cluster := &pedgev1alpha1.DeviceCluster{}
	err := r.Get(ctx, types.NamespacedName{Name: deviceClusterName, Namespace: secret.GetNamespace()}, cluster)
	if err != nil {
		return []reconcile.Request{}
	}
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
			},
		},
	}

}
