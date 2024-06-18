package controller

import (
	"context"

	rabbitmqv1 "github.com/rabbitmq/cluster-operator/api/v1beta1"
	rabbitmqtopologyv1 "github.com/rabbitmq/messaging-topology-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	// TODO change this
	devicesv1alpha1 "github.com/example/memcached-operator/api/v1alpha1"
)

// MQTTServerReconciler reconciles a MQTTServer object
type MQTTServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=devices.pedge.io,resources=mqttservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devices.pedge.io,resources=mqttservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=devices.pedge.io,resources=mqttservers/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=rabbitmq.com,resources=rabbitmqclusters;queues,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

func (r *MQTTServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the MQTTServer instance
	server := &devicesv1alpha1.MQTTServer{}
	err := r.Get(ctx, req.NamespacedName, server)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "unable to fetch MQTTServer")
		return ctrl.Result{}, err
	}

	// Check if the MQTTServer is marked for deletion
	if server.GetDeletionTimestamp() != nil {
		if containsString(server.GetFinalizers(), deviceFinalizer) {
			// Finalize the server
			if err := r.finalizeMQTTServer(ctx, server); err != nil {
				return ctrl.Result{}, err
			}
			// Remove finalizer
			server.SetFinalizers(removeString(server.GetFinalizers(), deviceFinalizer))
			if err := r.Update(ctx, server); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR
	if !containsString(server.GetFinalizers(), deviceFinalizer) {
		server.SetFinalizers(append(server.GetFinalizers(), deviceFinalizer))
		if err := r.Update(ctx, server); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Sync resources
	if err := r.syncResources(ctx, server); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// finalizeMQTTServer handles cleanup logic when a MQTTServer is deleted
func (r *MQTTServerReconciler) finalizeMQTTServer(ctx context.Context, server *devicesv1alpha1.MQTTServer) error {
	resources := []client.Object{
		&rabbitmqv1.RabbitmqCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      server.Name,
				Namespace: server.Namespace,
			},
		},
		&rabbitmqtopologyv1.Queue{
			ObjectMeta: metav1.ObjectMeta{
				Name:      server.Name,
				Namespace: server.Namespace,
			},
		},
	}

	for _, res := range resources {
		if err := r.Delete(ctx, res); client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}

// syncResources creates or updates the associated resources
func (r *MQTTServerReconciler) syncResources(ctx context.Context, server *devicesv1alpha1.MQTTServer) error {
	// Define the desired RabbitMQ Cluster resource
	cluster := &rabbitmqv1.RabbitmqCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name,
			Namespace: server.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(server, devicesv1alpha1.GroupVersion.WithKind("MQTTServer")),
			},
		},
		Spec: rabbitmqv1.RabbitmqClusterSpec{
			Service: rabbitmqv1.RabbitmqClusterServiceSpec{
				Type: "LoadBalancer", // TODO
			},
			Rabbitmq: rabbitmqv1.RabbitmqClusterConfigurationSpec{
				AdditionalPlugins: []rabbitmqv1.Plugin{"rabbitmq_mqtt"},
			},
		},
	}

	queue := &rabbitmqtopologyv1.Queue{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name,
			Namespace: server.Namespace,
		},
		Spec: rabbitmqtopologyv1.QueueSpec{
			Name: server.Spec.Queue.Name,
			RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
				Name:      server.Name,
				Namespace: server.Namespace,
			},
		},
	}

	resources := []client.Object{cluster, queue}
	for _, res := range resources {
		if err := r.CreateOrUpdate(ctx, res); err != nil {
			return err
		}
	}
	return nil
}

// CreateOrUpdate creates or updates a resource
func (r *MQTTServerReconciler) CreateOrUpdate(ctx context.Context, obj client.Object) error {
	key := types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}
	existing := obj.DeepCopyObject().(client.Object)
	err := r.Get(ctx, key, existing)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			return r.Create(ctx, obj)
		}
		return err
	}
	obj.SetResourceVersion(existing.GetResourceVersion())
	return r.Update(ctx, obj)
}

// SetupWithManager sets up the controller with the Manager
func (r *MQTTServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devicesv1alpha1.MQTTServer{}).
		Complete(r)
}
