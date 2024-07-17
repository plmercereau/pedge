package controller

import (
	"context"
	"fmt"

	pedgev1alpha1 "github.com/plmercereau/pedge/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	listenerUserName        = "device-listener"
	secretVersionAnnotation = "pedge.io/secret-version"
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
	// Fetch the DeviceCluster instance
	deviceCluster := &pedgev1alpha1.DeviceCluster{}
	err := r.Get(ctx, req.NamespacedName, deviceCluster)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "unable to fetch DeviceCluster. ")
		return ctrl.Result{}, err
	}

	if err := r.ensurePersistence(ctx, deviceCluster); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.syncRabbitmqCluster(ctx, deviceCluster); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.syncHttpServer(ctx, deviceCluster); err != nil {
		return ctrl.Result{}, err
	}

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
	attachedDevices := &pedgev1alpha1.DeviceClassList{}
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
