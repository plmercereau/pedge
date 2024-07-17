package controller

import (
	"context"
	"fmt"

	pedgev1alpha1 "github.com/plmercereau/pedge/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *DeviceClusterReconciler) ensurePersistence(ctx context.Context, deviceCluster *pedgev1alpha1.DeviceCluster) error {
	capacity := corev1.ResourceList{
		corev1.ResourceStorage: resource.MustParse("10Gi"),
	}
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: deviceCluster.Name + artefactSuffix,
		},
		Spec: corev1.PersistentVolumeSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			Capacity: capacity,
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/data",
					Type: new(corev1.HostPathType),
				},
			},
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
		},
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deviceCluster.Name + artefactSuffix,
			Namespace: deviceCluster.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: capacity,
			},
		},
	}

	// Check if the PV already exists, and if not, create it
	if err := r.Get(ctx, client.ObjectKeyFromObject(pv), pv); err != nil {
		if err := r.Client.Create(ctx, pv); err != nil {
			return fmt.Errorf("failed to create PV: %w", err)
		}
	}

	// Check if the PVC already exists, and if not, create it
	if err := r.Get(ctx, client.ObjectKeyFromObject(pvc), pvc); err != nil {
		if err := r.Client.Create(ctx, pvc); err != nil {
			return fmt.Errorf("failed to create PVC: %w", err)
		}
	}

	return nil
}
