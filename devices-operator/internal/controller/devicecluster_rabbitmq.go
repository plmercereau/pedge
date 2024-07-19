package controller

import (
	"context"
	"errors"
	"fmt"

	pedgev1alpha1 "github.com/plmercereau/pedge/api/v1alpha1"
	rabbitmqv1 "github.com/rabbitmq/cluster-operator/api/v1beta1"
	rabbitmqtopologyv1 "github.com/rabbitmq/messaging-topology-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// syncResources creates or updates the associated resources
func (r *DeviceClusterReconciler) syncRabbitmqCluster(ctx context.Context, deviceCluster *pedgev1alpha1.DeviceCluster) error {
	logger := log.FromContext(ctx)

	// TODO use possible custom broker/port values in the DeviceCluster spec instead of the service
	service := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: deviceCluster.Name, Namespace: deviceCluster.Namespace}, service); err != nil {
		logger.Error(err, "unable to fetch service")
		return err
	}

	var mqttBroker string
	if deviceCluster.Spec.MQTT.Hostname != "" {
		mqttBroker = deviceCluster.Spec.MQTT.Hostname
	} else {
		serviceIngress := service.Status.LoadBalancer.Ingress[0]
		if serviceIngress.Hostname != "" {
			mqttBroker = serviceIngress.Hostname
		} else {
			mqttBroker = serviceIngress.IP
		}
	}
	if mqttBroker == "" {
		return errors.New("no MQTT broker found")
	}

	mqttPortInt := 1883
	if deviceCluster.Spec.MQTT.Port != 0 {
		mqttPortInt = int(deviceCluster.Spec.MQTT.Port)
	} else {
		for _, port := range service.Spec.Ports {
			if port.Name == "mqtt" {
				mqttPortInt = int(port.Port)
				break
			}
		}
	}

	mqttPort := fmt.Sprint(mqttPortInt)
	deviceCluster.GetAnnotations()[mqttBrokerHostnameAnnotation] = mqttBroker
	deviceCluster.GetAnnotations()[mqttBrokerPortAnnotation] = mqttPort

	// Define the desired RabbitMQ Cluster resource
	cluster := &rabbitmqv1.RabbitmqCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deviceCluster.Name,
			Namespace: deviceCluster.Namespace,
		},
		Spec: rabbitmqv1.RabbitmqClusterSpec{
			Service: rabbitmqv1.RabbitmqClusterServiceSpec{
				Type: "LoadBalancer",
			},
			Rabbitmq: rabbitmqv1.RabbitmqClusterConfigurationSpec{
				AdditionalPlugins: []rabbitmqv1.Plugin{"rabbitmq_mqtt"},
			},
		},
	}

	if err := controllerutil.SetOwnerReference(deviceCluster, cluster, r.Scheme); err != nil {
		return err
	}

	if err := r.CreateOrUpdate(ctx, cluster); err != nil {
		return err
	}

	vhost := &rabbitmqtopologyv1.Vhost{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deviceCluster.Name + "-default",
			Namespace: deviceCluster.Namespace,
		},
		Spec: rabbitmqtopologyv1.VhostSpec{
			Name: "/",
			RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
				Name:      deviceCluster.Name,
				Namespace: deviceCluster.Namespace,
			},
		},
	}
	if err := controllerutil.SetOwnerReference(deviceCluster, vhost, r.Scheme); err != nil {
		return err
	}
	// We only create the vhost if it doesn't exist. The RabbitMQ messaging topology operator does not allow to modify it.
	// TODO we should also block some updates on the devices cluster name - through a validation webhook
	existingVhost := vhost.DeepCopyObject().(client.Object)
	if err := r.Get(ctx, client.ObjectKeyFromObject(vhost), existingVhost); err != nil && apierrors.IsNotFound(err) {
		if err := r.Create(ctx, vhost); err != nil {
			return err
		}
	}

	queue := &rabbitmqtopologyv1.Queue{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deviceCluster.Name,
			Namespace: deviceCluster.Namespace,
		},
		Spec: rabbitmqtopologyv1.QueueSpec{
			Name:  deviceCluster.Spec.MQTT.SensorsTopic,
			Vhost: vhost.Name,
			RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
				Name:      deviceCluster.Name,
				Namespace: deviceCluster.Namespace,
			},
		},
	}
	if err := controllerutil.SetOwnerReference(deviceCluster, queue, r.Scheme); err != nil {
		return err
	}

	// We only create the queue if it doesn't exist. The RabbitMQ messaging topology operator does not allow to modify it.
	existingQueue := queue.DeepCopyObject().(client.Object)
	if err := r.Get(ctx, client.ObjectKeyFromObject(queue), existingQueue); err != nil && apierrors.IsNotFound(err) {
		if err := r.Create(ctx, queue); err != nil {
			return err
		}
	}

	if deviceCluster.Spec.MQTT.Users != nil {
		// loop on users
		for _, user := range deviceCluster.Spec.MQTT.Users {
			var userSecret corev1.Secret
			secretName := user.Name + deviceSecretSuffix
			if err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: deviceCluster.Namespace}, &userSecret); err != nil {
				userSecret = corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      secretName,
						Namespace: deviceCluster.Namespace,
					},
					Data: map[string][]byte{
						"username": []byte(user.Name),
						"password": []byte(generateRandomPassword(16)),
					},
				}

				if err := r.Create(ctx, &userSecret); err != nil {
					return err
				}
			} else {
				changed := false
				if _, exists := userSecret.Data["username"]; !exists {
					userSecret.Data["username"] = []byte(user.Name)
					changed = true
				}
				if _, exists := userSecret.Data["password"]; !exists {
					userSecret.Data["password"] = []byte(generateRandomPassword(16))
					changed = true
				}
				if changed {
					if err := r.Update(ctx, &userSecret); err != nil {
						return err
					}

				}
			}

			rabbitmqUser := &rabbitmqtopologyv1.User{
				ObjectMeta: metav1.ObjectMeta{
					Name:      user.Name,
					Namespace: deviceCluster.Namespace,
					Annotations: map[string]string{
						// Needed to trigger a reconciliation when the password changes
						secretVersionAnnotation: userSecret.GetResourceVersion(),
					},
				},
				Spec: rabbitmqtopologyv1.UserSpec{
					RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
						Name:      deviceCluster.Name,
						Namespace: deviceCluster.Namespace,
					},
					ImportCredentialsSecret: &corev1.LocalObjectReference{
						Name: userSecret.Name,
					},
				},
			}
			// TODO it seems rabbitmq is already watching/owning the user, and when set to deviceCluster, the permissions are not applied
			// For now, accept users are not deleted with the devices cluster...
			// controllerutil.SetOwnerReference(deviceCluster, listenerUser, r.Scheme)

			// Define the desired RabbitMQ Permission resource
			userPermission := &rabbitmqtopologyv1.Permission{
				ObjectMeta: metav1.ObjectMeta{
					Name:      user.Name,
					Namespace: rabbitmqUser.Namespace,
				},
				Spec: rabbitmqtopologyv1.PermissionSpec{
					Vhost: "/",
					UserReference: &corev1.LocalObjectReference{
						Name: user.Name,
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
			// controllerutil.SetOwnerReference(deviceCluster, listenerPermission, r.Scheme)

			// Define the desired RabbitMQ TopicPermission resource
			// ! For now, we only support the "reader" role
			if user.Role != "reader" {
				return fmt.Errorf("Role %s is not supported", user.Role)
			}
			userTopicPermission := &rabbitmqtopologyv1.TopicPermission{
				ObjectMeta: metav1.ObjectMeta{
					Name:      user.Name,
					Namespace: rabbitmqUser.Namespace,
				},
				Spec: rabbitmqtopologyv1.TopicPermissionSpec{
					Vhost: "/",
					UserReference: &corev1.LocalObjectReference{
						Name: user.Name,
					},
					Permissions: rabbitmqtopologyv1.TopicPermissionConfig{
						Exchange: "amq.topic",
						Write:    "",
						Read:     fmt.Sprintf("^%s\\..+$", deviceCluster.Spec.MQTT.SensorsTopic),
					},
					RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
						Name:      deviceCluster.Name,
						Namespace: deviceCluster.Namespace,
					},
				},
			}
			// controllerutil.SetOwnerReference(deviceCluster, listenerTopicPermission, r.Scheme)
			resources := []client.Object{rabbitmqUser, userPermission, userTopicPermission}
			for _, res := range resources {
				if err := r.CreateOrUpdate(ctx, res); err != nil {
					return err
				}
			}

			for _, inject := range user.InjectSecrets {
				remoteSecret := &corev1.Secret{}
				if err := r.Get(ctx, types.NamespacedName{Name: inject.Name, Namespace: inject.Namespace}, remoteSecret); err != nil {
					// If a secret name is provided, then it must exist
					// TODO in such cases, create an Event for the user to understand why their reconcile is failing.
					// TODO also create an option "createIfNotExists" to create the secret if it does not exist
					logger.Error(err, "Secret injection: failed to get secret", "name", inject.Name, "namespace", inject.Namespace)
					return err
				}

				changed := false
				if inject.Annotations != nil {
					// Add reloader.stakater.com/match: "true" to the secret to trigger a reload of the telegraf config
					for key, value := range inject.Annotations {
						if remoteSecret.GetAnnotations()[key] != value {
							remoteSecret.GetAnnotations()[key] = value
							changed = true
						}
					}
				}

				usernameKey := inject.Mapping.Username
				if remoteSecret.Data[usernameKey] == nil || string(remoteSecret.Data[usernameKey]) != user.Name {
					remoteSecret.Data[usernameKey] = []byte(user.Name)
					changed = true
				}
				passwordKey := inject.Mapping.Password
				if remoteSecret.Data[passwordKey] == nil || string(remoteSecret.Data[passwordKey]) != string(userSecret.Data["password"]) {
					remoteSecret.Data[passwordKey] = userSecret.Data["password"]
					changed = true
				}
				sensorsTopicKey := inject.Mapping.SensorsTopic
				if sensorsTopicKey != "" {
					if remoteSecret.Data[sensorsTopicKey] == nil || string(remoteSecret.Data[sensorsTopicKey]) != deviceCluster.Spec.MQTT.SensorsTopic {
						remoteSecret.Data[sensorsTopicKey] = []byte(deviceCluster.Spec.MQTT.SensorsTopic)
						changed = true
					}
				}
				brokerUrlKey := inject.Mapping.BrokerUrl
				if brokerUrlKey != "" {
					mqttUrl := fmt.Sprintf("tcp://%s.%s.svc:%s", deviceCluster.Name, deviceCluster.Namespace, "1883")
					if remoteSecret.Data[brokerUrlKey] == nil || string(remoteSecret.Data[brokerUrlKey]) != mqttUrl {
						remoteSecret.Data[brokerUrlKey] = []byte(mqttUrl)
						changed = true
					}
				}
				if changed {
					if err := r.Update(ctx, remoteSecret); err != nil {
						return err
					}
				}
			}

		}
	}

	return nil
}
