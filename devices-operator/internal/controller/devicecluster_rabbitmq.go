package controller

import (
	"context"
	"fmt"

	pedgev1alpha1 "github.com/plmercereau/pedge/api/v1alpha1"
	rabbitmqv1 "github.com/rabbitmq/cluster-operator/api/v1beta1"
	rabbitmqtopologyv1 "github.com/rabbitmq/messaging-topology-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// syncResources creates or updates the associated resources
func (r *DeviceClusterReconciler) syncRabbitmqCluster(ctx context.Context, server *pedgev1alpha1.DeviceCluster) error {
	// Define the desired RabbitMQ Cluster resource
	cluster := &rabbitmqv1.RabbitmqCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name,
			Namespace: server.Namespace,
		},
		Spec: rabbitmqv1.RabbitmqClusterSpec{
			Service: rabbitmqv1.RabbitmqClusterServiceSpec{
				Type: "LoadBalancer",
			},
			Rabbitmq: rabbitmqv1.RabbitmqClusterConfigurationSpec{
				AdditionalPlugins: []rabbitmqv1.Plugin{"rabbitmq_mqtt"},
				// TODO only for testing purposes
				EnvConfig: `RABBITMQ_LOGS=""`,
				AdditionalConfig: `
log.console = true
log.console.level = debug
`,
			},
		},
	}

	if err := controllerutil.SetOwnerReference(server, cluster, r.Scheme); err != nil {
		return err
	}

	vhost := &rabbitmqtopologyv1.Vhost{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name + "-default",
			Namespace: server.Namespace,
		},
		Spec: rabbitmqtopologyv1.VhostSpec{
			Name: "/",
			RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
				Name:      server.Name,
				Namespace: server.Namespace,
			},
		},
	}
	if err := controllerutil.SetOwnerReference(server, vhost, r.Scheme); err != nil {
		return err
	}
	// We only create the vhost if it doesn't exist. The RabbitMQ messaging topology operator does not allow to modify it.
	// TODO we should also block some updates on the devices cluster name - through a validation webhook
	existingVhost := vhost.DeepCopyObject().(client.Object)
	if err := r.Get(ctx, client.ObjectKeyFromObject(vhost), existingVhost); err != nil && errors.IsNotFound(err) {
		if err := r.Create(ctx, vhost); err != nil {
			return err
		}
	}

	queue := &rabbitmqtopologyv1.Queue{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name,
			Namespace: server.Namespace,
		},
		Spec: rabbitmqtopologyv1.QueueSpec{
			Name:  server.Spec.MQTT.SensorsTopic,
			Vhost: vhost.Name,
			RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
				Name:      server.Name,
				Namespace: server.Namespace,
			},
		},
	}
	if err := controllerutil.SetOwnerReference(server, queue, r.Scheme); err != nil {
		return err
	}

	// We only create the queue if it doesn't exist. The RabbitMQ messaging topology operator does not allow to modify it.
	existingQueue := queue.DeepCopyObject().(client.Object)
	if err := r.Get(ctx, client.ObjectKeyFromObject(queue), existingQueue); err != nil && errors.IsNotFound(err) {
		if err := r.Create(ctx, queue); err != nil {
			return err
		}
	}

	var listenerSecret corev1.Secret
	secretName := listenerUserName + deviceSecretSuffix
	if err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: server.Namespace}, &listenerSecret); err != nil {
		listenerSecret = corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: server.Namespace,
			},
			Data: map[string][]byte{
				"username": []byte(listenerUserName),
				"password": []byte(generateRandomPassword(16)),
			},
		}

		if err := r.Create(ctx, &listenerSecret); err != nil {
			return err
		}
	} else {
		if string(listenerSecret.Data["username"]) != listenerUserName {
			listenerSecret.Data["username"] = []byte(listenerUserName)
			if err := r.Update(ctx, &listenerSecret); err != nil {
				return err
			}
		}
	}

	// Store MQTT URL/TOPIC/USERNAME/PASSWORD in influxdb-auth in the influxdb namespace
	if server.Spec.InfluxDB != (pedgev1alpha1.InfluxDB{}) {
		influxDBSecret := &corev1.Secret{}
		if err := r.Get(ctx, types.NamespacedName{Name: server.Spec.InfluxDB.SecretReference.Name, Namespace: server.Spec.InfluxDB.Namespace}, influxDBSecret); err != nil {
			// If a secret name is provided, then it must exist
			// TODO in such cases, create an Event for the user to understand why their reconcile is failing.
			return err
		}

		changed := false

		// Add reloader.stakater.com/match: "true" to the secret to trigger a reload of the telegraf config
		if influxDBSecret.Annotations == nil {
			influxDBSecret.Annotations = make(map[string]string)
			// check if reloader.stakater.com/match exists and is set to "true"
			if influxDBSecret.Annotations["reloader.stakater.com/match"] != "true" {
				influxDBSecret.Annotations["reloader.stakater.com/match"] = "true"
				changed = true
			}
		}

		if influxDBSecret.Data["MQTT_USERNAME"] == nil || string(influxDBSecret.Data["MQTT_USERNAME"]) != listenerUserName {
			influxDBSecret.Data["MQTT_USERNAME"] = []byte(listenerUserName)
			changed = true
		}
		if influxDBSecret.Data["MQTT_PASSWORD"] == nil || string(influxDBSecret.Data["MQTT_PASSWORD"]) != string(listenerSecret.Data["password"]) {
			influxDBSecret.Data["MQTT_PASSWORD"] = listenerSecret.Data["password"]
			changed = true
		}
		if influxDBSecret.Data["MQTT_TOPIC"] == nil || string(influxDBSecret.Data["MQTT_TOPIC"]) != server.Spec.MQTT.SensorsTopic {
			influxDBSecret.Data["MQTT_TOPIC"] = []byte(server.Spec.MQTT.SensorsTopic)
			changed = true
		}
		mqttUrl := fmt.Sprintf("tcp://%s.%s.svc:%s", server.Name, server.Namespace, "1883")
		if influxDBSecret.Data["MQTT_URL"] == nil || string(influxDBSecret.Data["MQTT_URL"]) != mqttUrl {
			influxDBSecret.Data["MQTT_URL"] = []byte(mqttUrl)
			changed = true
		}
		// ! For some reason, telegraf does not allow admin-token as an environment variable, maybe because of the dash
		if influxDBSecret.Data["INFLUXDB_TOKEN"] == nil || string(influxDBSecret.Data["INFLUXDB_TOKEN"]) != string(influxDBSecret.Data["admin-token"]) {
			influxDBSecret.Data["INFLUXDB_TOKEN"] = influxDBSecret.Data["admin-token"]
			changed = true
		}
		if changed {
			if err := r.Update(ctx, influxDBSecret); err != nil {
				return err
			}
		}
	}

	listenerUser := &rabbitmqtopologyv1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name:      listenerUserName,
			Namespace: server.Namespace,
			Annotations: map[string]string{
				// Needed to trigger a reconciliation when the password changes
				secretVersionAnnotation: listenerSecret.GetResourceVersion(),
			},
		},
		Spec: rabbitmqtopologyv1.UserSpec{
			RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
				Name:      server.Name,
				Namespace: server.Namespace,
			},
			ImportCredentialsSecret: &corev1.LocalObjectReference{
				Name: listenerSecret.Name,
			},
		},
	}
	// TODO it seems rabbitmq is already watching/owning the user, and when set to server, the permissions are not applied
	// For now, accept users are not deleted with the devices cluster...
	// controllerutil.SetOwnerReference(server, listenerUser, r.Scheme)

	// Define the desired RabbitMQ Permission resource
	listenerPermission := &rabbitmqtopologyv1.Permission{
		ObjectMeta: metav1.ObjectMeta{
			Name:      listenerUser.Name,
			Namespace: listenerUser.Namespace,
		},
		Spec: rabbitmqtopologyv1.PermissionSpec{
			Vhost: "/",
			UserReference: &corev1.LocalObjectReference{
				Name: listenerUser.Name,
			},
			Permissions: rabbitmqtopologyv1.VhostPermissions{
				Configure: "^mqtt-subscription-.*$",
				Write:     "^amq\\.topic$|^mqtt-subscription-.*$",
				Read:      "^amq\\.topic$|^mqtt-subscription-.*$",
			},
			RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
				Name:      server.Name,
				Namespace: server.Namespace,
			},
		},
	}
	// controllerutil.SetOwnerReference(server, listenerPermission, r.Scheme)

	// Define the desired RabbitMQ TopicPermission resource
	listenerTopicPermission := &rabbitmqtopologyv1.TopicPermission{
		ObjectMeta: metav1.ObjectMeta{
			Name:      listenerUser.Name,
			Namespace: listenerUser.Namespace,
		},
		Spec: rabbitmqtopologyv1.TopicPermissionSpec{
			Vhost: "/",
			UserReference: &corev1.LocalObjectReference{
				Name: listenerUser.Name,
			},
			Permissions: rabbitmqtopologyv1.TopicPermissionConfig{
				Exchange: "amq.topic",
				Write:    "",
				Read:     fmt.Sprintf("^%s\\..+$", server.Spec.MQTT.SensorsTopic),
			},
			RabbitmqClusterReference: rabbitmqtopologyv1.RabbitmqClusterReference{
				Name:      server.Name,
				Namespace: server.Namespace,
			},
		},
	}
	// controllerutil.SetOwnerReference(server, listenerTopicPermission, r.Scheme)

	resources := []client.Object{cluster, listenerUser, listenerPermission, listenerTopicPermission}
	for _, res := range resources {
		if err := r.CreateOrUpdate(ctx, res); err != nil {
			return err
		}
	}
	return nil
}
