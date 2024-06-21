package controller

import (
	"context"

	pedgev1alpha1 "github.com/plmercereau/pedge/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// FirmwareReconciler reconciles a Firmware object
type FirmwareReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=devices.pedge.io,resources=firmwares,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devices.pedge.io,resources=firmwares/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

func (r *FirmwareReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the Firmware instance
	firmware := &pedgev1alpha1.Firmware{}
	err := r.Get(ctx, req.NamespacedName, firmware)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Uunable to fetch Firmware")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *FirmwareReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pedgev1alpha1.Firmware{}).
		Owns(&pedgev1alpha1.Device{}).
		Complete(r)
}
