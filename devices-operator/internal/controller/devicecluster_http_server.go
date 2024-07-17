package controller

import (
	"context"
	"fmt"

	pedgev1alpha1 "github.com/plmercereau/pedge/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	artefactServiceSuffix = "-artefacts"
)

//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *DeviceClusterReconciler) syncHttpServer(ctx context.Context, server *pedgev1alpha1.DeviceCluster) error {
	artefactsServiceImage := server.Spec.Artefacts.Image
	replicasValue := int32(1)
	containerPort := int32(5000)
	volumeName := server.Name + "-data"
	labels := map[string]string{"app": server.Name}
	labelSelector := metav1.LabelSelector{MatchLabels: labels}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name + artefactServiceSuffix,
			Namespace: server.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(server, pedgev1alpha1.GroupVersion.WithKind("DeviceCluster")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicasValue,
			Selector: &labelSelector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  server.Name,
							Image: fmt.Sprintf("%s:%s", artefactsServiceImage.Repository, artefactsServiceImage.Tag),
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: containerPort,
								},
							},
							Resources: corev1.ResourceRequirements{
								// ?
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      volumeName,
									MountPath: "/data",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: volumeName,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "devices", // server.Name + "-data", // TODO
								},
							},
						},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{},
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name + artefactServiceSuffix,
			Namespace: server.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(server, pedgev1alpha1.GroupVersion.WithKind("DeviceCluster")),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt32(containerPort),
				},
			},
			Selector: labels,
		},
	}

	if err := r.CreateOrUpdate(ctx, deployment); err != nil {
		return err
	}
	if err := r.CreateOrUpdate(ctx, service); err != nil {
		return err
	}

	if server.Spec.Artefacts.Ingress.Enabled {
		pathTypePrefix := networkingv1.PathTypePrefix
		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      server.Name,
				Namespace: server.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(server, pedgev1alpha1.GroupVersion.WithKind("DeviceCluster")),
				},
			},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{
						Host: server.Spec.Artefacts.Ingress.Hostname,
						IngressRuleValue: networkingv1.IngressRuleValue{
							HTTP: &networkingv1.HTTPIngressRuleValue{
								Paths: []networkingv1.HTTPIngressPath{
									{
										Path:     "/",
										PathType: &pathTypePrefix,
										Backend: networkingv1.IngressBackend{
											Service: &networkingv1.IngressServiceBackend{
												Name: server.Name + artefactServiceSuffix,
												Port: networkingv1.ServiceBackendPort{
													Number: 80,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		if err := r.CreateOrUpdate(ctx, ingress); err != nil {
			return err
		}

	}
	return nil
}
