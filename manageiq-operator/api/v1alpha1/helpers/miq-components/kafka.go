package miqtools

import (
	"context"

	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func ManageKafkaSecret(cr *miqv1alpha1.ManageIQ, client client.Client, scheme *runtime.Scheme) (*corev1.Secret, controllerutil.MutateFn) {
	secretKey := types.NamespacedName{Namespace: cr.ObjectMeta.Namespace, Name: cr.Spec.KafkaSecret}
	secret := &corev1.Secret{}
	secretErr := client.Get(context.TODO(), secretKey, secret)
	if secretErr != nil {
		secret = defaultKafkaSecret(cr)
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, secret, scheme); err != nil {
			return err
		}

		addAppLabel(cr.Spec.AppName, &secret.ObjectMeta)
		addBackupLabel(cr.Spec.BackupLabelName, &secret.ObjectMeta)

		return nil
	}

	return secret, f
}

func defaultKafkaSecret(cr *miqv1alpha1.ManageIQ) *corev1.Secret {
	secretData := map[string]string{
		"username": "root",
		"password": generatePassword(),
		"hostname": "kafka",
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kafkaSecretName(cr),
			Namespace: cr.ObjectMeta.Namespace,
		},
		StringData: secretData,
	}

	addAppLabel(cr.Spec.AppName, &secret.ObjectMeta)
	addBackupLabel(cr.Spec.BackupLabelName, &secret.ObjectMeta)

	return secret
}

func kafkaSecretName(cr *miqv1alpha1.ManageIQ) string {
	secretName := "kafka-secrets"
	if cr.Spec.KafkaSecret != "" {
		secretName = cr.Spec.KafkaSecret
	}

	return secretName
}

func KafkaPVC(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*corev1.PersistentVolumeClaim, controllerutil.MutateFn) {
	storageReq, _ := resource.ParseQuantity(cr.Spec.KafkaVolumeCapacity)

	resources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			"storage": storageReq,
		},
	}

	accessModes := []corev1.PersistentVolumeAccessMode{
		"ReadWriteOnce",
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kafka-data",
			Namespace: cr.ObjectMeta.Namespace,
		},
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, pvc, scheme); err != nil {
			return err
		}

		addAppLabel(cr.Spec.AppName, &pvc.ObjectMeta)
		addBackupLabel(cr.Spec.BackupLabelName, &pvc.ObjectMeta)
		pvc.Spec.AccessModes = accessModes
		pvc.Spec.Resources = resources

		if cr.Spec.StorageClassName != "" {
			pvc.Spec.StorageClassName = &cr.Spec.StorageClassName
		}
		return nil
	}

	return pvc, f
}

func ZookeeperPVC(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*corev1.PersistentVolumeClaim, controllerutil.MutateFn) {
	storageReq, _ := resource.ParseQuantity(cr.Spec.DatabaseVolumeCapacity)

	resources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			"storage": storageReq,
		},
	}

	accessModes := []corev1.PersistentVolumeAccessMode{
		"ReadWriteOnce",
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "zookeeper-data",
			Namespace: cr.ObjectMeta.Namespace,
		},
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, pvc, scheme); err != nil {
			return err
		}

		addAppLabel(cr.Spec.AppName, &pvc.ObjectMeta)
		addBackupLabel(cr.Spec.BackupLabelName, &pvc.ObjectMeta)
		pvc.Spec.AccessModes = accessModes
		pvc.Spec.Resources = resources

		if cr.Spec.StorageClassName != "" {
			pvc.Spec.StorageClassName = &cr.Spec.StorageClassName
		}
		return nil
	}

	return pvc, f
}

func KafkaService(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*corev1.Service, controllerutil.MutateFn) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kafka",
			Namespace: cr.ObjectMeta.Namespace,
		},
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, service, scheme); err != nil {
			return err
		}

		addAppLabel(cr.Spec.AppName, &service.ObjectMeta)
		if len(service.Spec.Ports) == 0 {
			service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{})
		}
		// service.Spec.Ports[0].Name = "kafka"
		// service.Spec.Ports[0].Port = 9092

		service.Spec.Ports = []corev1.ServicePort{
			corev1.ServicePort{
				Name: "kafka",
				Port: 9092,
			},
			corev1.ServicePort{
				Name: "localhost",
				Port: 9093,
			},
		}
		service.Spec.Selector = map[string]string{"name": "kafka"}
		return nil
	}

	return service, f
}

func ZookeeperService(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*corev1.Service, controllerutil.MutateFn) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "zookeeper",
			Namespace: cr.ObjectMeta.Namespace,
		},
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, service, scheme); err != nil {
			return err
		}

		addAppLabel(cr.Spec.AppName, &service.ObjectMeta)
		if len(service.Spec.Ports) == 0 {
			service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{})
		}
		service.Spec.Ports[0].Name = "zookeeper"
		service.Spec.Ports[0].Port = 2181

		// service.Spec.Ports = []corev1.ServicePort{
		// 	corev1.ServicePort{
		// 		Name: "zookeeperplain",
		// 		Port: 2181,
		// 	},
		// 	corev1.ServicePort{
		// 		Name: "zookeeper",
		// 		Port: 2182,
		// 	},
		// }
		service.Spec.Selector = map[string]string{"name": "zookeeper"}
		return nil
	}

	return service, f
}

func KafkaDeployment(cr *miqv1alpha1.ManageIQ, client client.Client, scheme *runtime.Scheme) (*appsv1.Deployment, controllerutil.MutateFn, error) {
	deploymentLabels := map[string]string{
		"name": "kafka",
		"app":  cr.Spec.AppName,
	}

	container := corev1.Container{
		Name:            "kafka",
		Image:           cr.Spec.KafkaImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Ports: []corev1.ContainerPort{
			corev1.ContainerPort{
				ContainerPort: 9092,
			},
			corev1.ContainerPort{
				ContainerPort: 9093,
			},
		},
		LivenessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(9092),
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(9092),
				},
			},
		},
		Env: []corev1.EnvVar{
			corev1.EnvVar{
				Name: "KAFKA_BROKER_USER",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: kafkaSecretName(cr)},
						Key:                  "username",
					},
				},
			},
			corev1.EnvVar{
				Name: "KAFKA_BROKER_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: kafkaSecretName(cr)},
						Key:                  "password",
					},
				},
			},
			// corev1.EnvVar{
			// 	Name:  "KAFKA_BROKER_USER",
			// 	Value: "user",
			// },
			// corev1.EnvVar{
			// 	Name:  "KAFKA_BROKER_PASSWORD",
			// 	Value: "password",
			// },
			// corev1.EnvVar{
			// 	Name:  "KAFKA_INTER_BROKER_PASSWORD",
			// 	Value: "password",
			// },
			corev1.EnvVar{
				Name:  "KAFKA_ClIENT_USERS",
				Value: "user",
			},
			corev1.EnvVar{
				Name:  "KAFKA_CLIENT_PASSWORDS",
				Value: "password",
			},
			corev1.EnvVar{
				Name:  "KAFKA_ZOOKEEPER_USER",
				Value: "user",
			},
			corev1.EnvVar{
				Name:  "KAFKA_ZOOKEEPER_PASSWORD",
				Value: "password",
			},
			corev1.EnvVar{
				Name:  "ALLOW_PLAINTEXT_LISTENER",
				Value: "yes",
			},
			corev1.EnvVar{
				Name:  "KAFKA_ZOOKEEPER_CONNECT",
				Value: "zookeeper:2181",
			},
			corev1.EnvVar{
				Name:  "KAFKA_ZOOKEEPER_PROTOCOL",
				Value: "SASL",
			},
			// corev1.EnvVar{
			// 	Name:  "KAFKA_ZOOKEEPER_TLS_KEYSTORE_PASSWORD",
			// 	Value: "nasar123",
			// },
			// corev1.EnvVar{
			// 	Name:  "KAFKA_ZOOKEEPER_TLS_TRUSTSTORE_PASSWORD",
			// 	Value: "nasar123",
			// },
			// corev1.EnvVar{ // NOTE: may need to adjust path to /opt/.. and/or change name to zookeeper.truststore.jks
			// 	Name:  "KAFKA_ZOOKEEPER_TLS_TRUSTSTORE_FILE",
			// 	Value: "/opt/bitnami/kafka/config/certs/kafka.truststore.jks",
			// },
			corev1.EnvVar{
				Name:  "KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP",
				Value: "INTERNAL:SASL_SSL,CLIENT:SASL_SSL",
			},
			// corev1.EnvVar{
			// 	Name:  "KAFKA_CFG_LISTENERS",
			// 	Value: "INTERNAL://:9093,CLIENT://:9092",
			// },
			// corev1.EnvVar{
			// 	Name:  "KAFKA_CFG_INTER_BROKER_LISTENER_NAME",
			// 	Value: "INTERNAL",
			// },
			// corev1.EnvVar{
			// 	Name:  "KAFKA_CFG_ADVERTISED_LISTENERS",
			// 	Value: "INTERNAL://localhost:9093,CLIENT://kafka:9092",
			// },
			corev1.EnvVar{
				Name:  "KAFKA_CFG_LISTENERS",
				Value: "INTERNAL://:9092,CLIENT://:9093",
			},
			corev1.EnvVar{
				Name:  "KAFKA_CFG_INTER_BROKER_LISTENER_NAME",
				Value: "INTERNAL",
			},
			corev1.EnvVar{
				Name:  "KAFKA_CFG_ADVERTISED_LISTENERS",
				Value: "INTERNAL://kafka:9092,CLIENT://localhost:9093",
			},
			corev1.EnvVar{
				Name:  "KAFKA_CFG_LISTENER_NAME_INTERNAL_SSL_CLIENT_AUTH",
				Value: "required",
			},
			corev1.EnvVar{
				Name:  "KAFKA_CFG_LISTENER_NAME_CLIENT_SSL_CLIENT_AUTH",
				Value: "required",
			},
			// corev1.EnvVar{
			// 	Name:  "KAFKA_CFG_LISTENERS",
			// 	Value: "SASL_SSL://kafka:9092",
			// },
			// corev1.EnvVar{
			// 	Name:  "KAFKA_CFG_ADVERTISED_LISTENERS",
			// 	Value: "SASL_SSL://kafka:9092",
			// },
			// corev1.EnvVar{
			// 	Name:  "KAFKA_CFG_INTER_BROKER_LISTENER_NAME",
			// 	Value: "SASL_SSL",
			// },
			corev1.EnvVar{
				Name:  "KAFKA_CFG_SASL_MECHANISM_INTER_BROKER_PROTOCOL",
				Value: "PLAIN",
			},
			corev1.EnvVar{
				Name:  "KAFKA_CERTIFICATE_PASSWORD",
				Value: "nasar123",
			},
			// corev1.EnvVar{
			// 	Name:  "KAFKA_CFG_LISTENER_NAME_SASL_SSL_PLAIN_SASL_JAAS_CONFIG",
			// 	Value: "KafkaServer {org.apache.kafka.common.security.plain.PlainLoginModule required username=\"user\" password=\"password\" user_admin=\"password\";};",
			// },
			corev1.EnvVar{
				Name:  "BITNAMI_DEBUG",
				Value: "true",
			},
			// corev1.EnvVar{
			// 	Name:  "KAFKA_OPTS",
			// 	Value: "-Djava.security.auth.login.config=/opt/bitnami/kafka/config/kafka_jaas.conf",
			// },
			// corev1.EnvVar{
			// 	Name:  "KAFKA_CFG_SSL_ENDPOINT_IDENTIFICATION_ALGORITHM",
			// 	Value: "",
			// },
		},
		VolumeMounts: []corev1.VolumeMount{
			corev1.VolumeMount{Name: "kafka-data", MountPath: "/bitnami/kafka"},
		},
	}

	err := addResourceReqs(cr.Spec.KafkaMemoryLimit, cr.Spec.KafkaMemoryRequest, cr.Spec.KafkaCpuLimit, cr.Spec.KafkaCpuRequest, &container)
	if err != nil {
		return nil, nil, err
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kafka",
			Namespace: cr.ObjectMeta.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: deploymentLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: deploymentLabels,
					Name:   "kafka",
				},
				Spec: corev1.PodSpec{
					Hostname: "kafka",
				},
			},
		},
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, deployment, scheme); err != nil {
			return err
		}
		addAppLabel(cr.Spec.AppName, &deployment.ObjectMeta)
		addBackupAnnotation("kafka-data", &deployment.Spec.Template.ObjectMeta)
		addBackupLabel(cr.Spec.BackupLabelName, &deployment.ObjectMeta)
		addBackupLabel(cr.Spec.BackupLabelName, &deployment.Spec.Template.ObjectMeta)
		var repNum int32 = 1
		deployment.Spec.Replicas = &repNum
		deployment.Spec.Strategy = appsv1.DeploymentStrategy{
			Type: "Recreate",
		}
		deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
		deployment.Spec.Template.Spec.Containers[0].SecurityContext = DefaultSecurityContext()
		deployment.Spec.Template.Spec.ServiceAccountName = defaultServiceAccountName(cr.Spec.AppName)
		var termSecs int64 = 10
		deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = &termSecs
		deployment.Spec.Template.Spec.Volumes = []corev1.Volume{
			corev1.Volume{
				Name: "kafka-data",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: "kafka-data",
					},
				},
			},
		}
		addKafkaStores(cr, deployment, client, "/opt/bitnami/kafka/config/certs")

		return nil
	}

	return deployment, f, nil
}

func ZookeeperDeployment(cr *miqv1alpha1.ManageIQ, client client.Client, scheme *runtime.Scheme) (*appsv1.Deployment, controllerutil.MutateFn, error) {
	deploymentLabels := map[string]string{
		"name": "zookeeper",
		"app":  cr.Spec.AppName,
	}

	container := corev1.Container{
		Name:            "zookeeper",
		Image:           cr.Spec.ZookeeperImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Ports: []corev1.ContainerPort{
			corev1.ContainerPort{
				ContainerPort: 2181,
			},
			// corev1.ContainerPort{
			// 	ContainerPort: 2182,
			// },
		},
		Env: []corev1.EnvVar{
			corev1.EnvVar{
				Name:  "ZOO_SERVER_USERS",
				Value: "user",
			},
			corev1.EnvVar{
				Name:  "ZOO_SERVER_PASSWORDS",
				Value: "password",
			},
			corev1.EnvVar{
				Name:  "ZOO_ENABLE_AUTH",
				Value: "yes",
			},
			corev1.EnvVar{
				Name:  "ZOO_CLIENT_USER",
				Value: "user",
			},
			corev1.EnvVar{
				Name:  "ZOO_CLIENT_PASSWORD",
				Value: "password",
			},
			// corev1.EnvVar{
			// 	Name:  "ALLOW_ANONYMOUS_LOGIN",
			// 	Value: "yes",
			// },
			// corev1.EnvVar{
			// 	Name:  "ZOO_TLS_CLIENT_ENABLE",
			// 	Value: "yes",
			// },
			// corev1.EnvVar{
			// 	Name:  "ZOO_TLS_CLIENT_AUTH",
			// 	Value: "none",
			// },
			// corev1.EnvVar{
			// 	Name:  "ZOO_TLS_CLIENT_KEYSTORE_FILE",
			// 	Value: "/bitnami/zookeeper/certs/zookeeper.keystore.jks",
			// },
			// corev1.EnvVar{
			// 	Name:  "ZOO_TLS_CLIENT_KEYSTORE_PASSWORD",
			// 	Value: "nasar123",
			// },
			// corev1.EnvVar{
			// 	Name:  "ZOO_TLS_CLIENT_TRUSTSTORE_FILE",
			// 	Value: "/bitnami/zookeeper/certs/kafka.truststore.jks",
			// },
			// corev1.EnvVar{
			// 	Name:  "ZOO_TLS_CLIENT_TRUSTSTORE_PASSWORD",
			// 	Value: "nasar123",
			// },
			// corev1.EnvVar{
			// 	Name:  "ZOO_TLS_PORT_NUMBER",
			// 	Value: "2182",
			// },
			corev1.EnvVar{
				Name:  "BITNAMI_DEBUG",
				Value: "true",
			},
			// corev1.EnvVar{
			// 	Name:  "ZOO_LOG_LEVEL",
			// 	Value: "DEBUG",
			// },
		},
		VolumeMounts: []corev1.VolumeMount{
			corev1.VolumeMount{Name: "zookeeper-data", MountPath: "/bitnami/zookeeper"},
		},
	}

	err := addResourceReqs(cr.Spec.ZookeeperMemoryLimit, cr.Spec.ZookeeperMemoryRequest, cr.Spec.ZookeeperCpuLimit, cr.Spec.ZookeeperCpuRequest, &container)
	if err != nil {
		return nil, nil, err
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "zookeeper",
			Namespace: cr.ObjectMeta.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: deploymentLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: deploymentLabels,
					Name:   "zookeeper",
				},
				Spec: corev1.PodSpec{},
				// Spec: corev1.PodSpec{
				// 	Hostname: "zookeeper",
				// },
			},
		},
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, deployment, scheme); err != nil {
			return err
		}
		addAppLabel(cr.Spec.AppName, &deployment.ObjectMeta)
		addBackupAnnotation("zookeeper-data", &deployment.Spec.Template.ObjectMeta)
		addBackupLabel(cr.Spec.BackupLabelName, &deployment.ObjectMeta)
		addBackupLabel(cr.Spec.BackupLabelName, &deployment.Spec.Template.ObjectMeta)
		var repNum int32 = 1
		deployment.Spec.Replicas = &repNum
		deployment.Spec.Strategy = appsv1.DeploymentStrategy{
			Type: "Recreate",
		}
		addAnnotations(cr.Spec.AppAnnotations, &deployment.Spec.Template.ObjectMeta)
		deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
		deployment.Spec.Template.Spec.Containers[0].SecurityContext = DefaultSecurityContext()
		deployment.Spec.Template.Spec.ServiceAccountName = defaultServiceAccountName(cr.Spec.AppName)
		deployment.Spec.Template.Spec.Volumes = []corev1.Volume{
			corev1.Volume{
				Name: "zookeeper-data",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: "zookeeper-data",
					},
				},
			},
		}
		// addZookeeperStores(cr, deployment, client, "/bitnami/zookeeper/certs")

		return nil
	}

	return deployment, f, nil
}
