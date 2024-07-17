package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"

	pedgev1alpha1 "github.com/plmercereau/pedge/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	secretTypeLabel           = "pedge.io/secret-type"
	deviceNameLabel           = "pedge.io/device-name"
	deviceSecretHashLabel     = "pedge.io/device-secret-hash"
	mqttDeviceSecretHashLabel = "pedge.io/mqtt-device-secret-hash"
	maxRetries                = 5
	initialBackoff            = 100 * time.Millisecond
)

// DeviceSecretWatcherReconciler reconciles a DeviceSecretWatcher object
type DeviceSecretWatcherReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=secrets/status,verbs=get

func (r *DeviceSecretWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the Secret instance
	var secret corev1.Secret
	if err := r.Get(ctx, req.NamespacedName, &secret); client.IgnoreNotFound(err) != nil {
		logger.Error(err, "unable to fetch Secret")
		return ctrl.Result{}, err
	}

	deviceName := secret.Labels[deviceNameLabel]
	logger.Info("Secret fetched", "name", secret.Name)

	secretHasher := sha256.New()
	keys := make([]string, 0, len(secret.Data))

	for k := range secret.Data {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	for _, key := range keys {
		secretHasher.Write([]byte(key))
		secretHasher.Write(secret.Data[key])
	}

	secretHash := hex.EncodeToString(secretHasher.Sum(nil))

	mqttHasher := sha256.New()
	mqttHasher.Write(secret.Data["username"])
	mqttHasher.Write(secret.Data["password"])
	mqttHash := hex.EncodeToString(mqttHasher.Sum(nil))

	backoff := initialBackoff
	retryCount := 0

	if retryCount == maxRetries {
		logger.Error(nil, "Failed to update device after retries", "name", deviceName)
		deviceGR := schema.GroupResource{Group: "devices.pedge.io", Resource: "devices"}
		return ctrl.Result{}, errors.NewConflict(deviceGR, deviceName, nil)
	}

	backoff = initialBackoff
	for retryCount = 0; retryCount < maxRetries; retryCount++ {
		patch := client.MergeFrom(secret.DeepCopy())
		// Check if there is a password, required for RabbitMQ
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		if _, exists := secret.Data["password"]; !exists {
			secret.Data["password"] = []byte(generateRandomPassword(16))
		}
		if _, exists := secret.Data["username"]; !exists || string(secret.Data["username"]) != deviceName {
			secret.Data["username"] = []byte(deviceName)
		}
		if err := r.Patch(ctx, &secret, patch); err != nil {
			if errors.IsConflict(err) {
				logger.Info("Conflict when updating secret, retrying", "name", secret.Name)
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return ctrl.Result{}, err
		}
		break
	}

	if retryCount == maxRetries {
		logger.Error(nil, "Failed to update secret after retries", "name", secret.Name)
		secretGR := schema.GroupResource{Group: "core", Resource: "secrets"}
		return ctrl.Result{}, errors.NewConflict(secretGR, secret.Name, nil)
	}

	// We need to possibly update the device in order to sync the RabbitMQ User credentials
	// TODO check if the device controller updates the user. Or, do it from here and forget about the mqttDeviceSecretHashLabel annotation on the device
	for retryCount = 0; retryCount < maxRetries; retryCount++ {
		// Fetch the latest version of the device instance
		device := &pedgev1alpha1.Device{}
		if err := r.Get(ctx, types.NamespacedName{Name: deviceName, Namespace: secret.Namespace}, device); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}

		// Create a copy of the device to create the patch
		patch := client.MergeFrom(device.DeepCopy())

		// Update the annotations
		if device.Annotations == nil {
			device.Annotations = make(map[string]string)
		}

		if device.Annotations[mqttDeviceSecretHashLabel] == mqttHash && device.Annotations[deviceSecretHashLabel] == secretHash {
			logger.Info("No change in the secret that need to be reflected in the device", "name", secret.Name)
			break
		}
		device.Annotations[mqttDeviceSecretHashLabel] = mqttHash
		device.Annotations[deviceSecretHashLabel] = secretHash

		// Apply the patch
		if err := r.Patch(ctx, device, patch); err != nil {
			if errors.IsConflict(err) {
				logger.Info("Conflict when updating device, retrying", "name", device.Name)
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return ctrl.Result{}, err
		}
		break
	}

	return ctrl.Result{}, nil
}

func (r *DeviceSecretWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	labelSelector := labels.SelectorFromSet(labels.Set{secretTypeLabel: "device"})

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		WithEventFilter(predicate.NewPredicateFuncs(func(object client.Object) bool {
			if !labelSelector.Matches(labels.Set(object.GetLabels())) {
				return false
			}
			_, exists := object.GetLabels()[deviceNameLabel]
			return exists
		})).
		Complete(r)
}
