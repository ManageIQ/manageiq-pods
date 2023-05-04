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

		service.Spec.Ports = []corev1.ServicePort{
			corev1.ServicePort{
				Name: "kafka",
				Port: 9092,
			},
			corev1.ServicePort{
				Name: "controller",
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
		service.Spec.Selector = map[string]string{"name": "zookeeper"}
		return nil
	}

	return service, f
}

func updateKafkaEnv(cr *miqv1alpha1.ManageIQ, client client.Client, c *corev1.Container) {
	c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "KAFKA_ENABLE_KRAFT", Value: "yes"})
	c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "KAFKA_CFG_PROCESS_ROLES", Value: "broker,controller"})
	c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "KAFKA_CFG_CONTROLLER_LISTENER_NAMES", Value: "CONTROLLER"})
	c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "KAFKA_CFG_LISTENERS", Value: "INTERNAL://:9092,CONTROLLER://:9093"})
	c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "KAFKA_CFG_INTER_BROKER_LISTENER_NAME", Value: "INTERNAL"})
	c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "KAFKA_CFG_ADVERTISED_LISTENERS", Value: "INTERNAL://kafka:9092"})
	c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "KAFKA_CFG_CONTROLLER_QUORUM_VOTERS", Value: "1@kafka:9093"})
	c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "KAFKA_CFG_BROKER_ID", Value: "1"})

	certSecret := InternalCertificatesSecret(cr, client)
	if certSecret.Data["kafka_truststore"] != nil && certSecret.Data["kafka_keystore"] != nil && certSecret.Data["kafka_keystore_pass"] != nil {
		c.Env = addOrUpdateEnvVar(c.Env,
			corev1.EnvVar{
				Name: "KAFKA_INTER_BROKER_USER",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: kafkaSecretName(cr)},
						Key:                  "username",
					},
				},
			},
		)
		c.Env = addOrUpdateEnvVar(c.Env,
			corev1.EnvVar{
				Name: "KAFKA_INTER_BROKER_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: kafkaSecretName(cr)},
						Key:                  "password",
					},
				},
			},
		)
		c.Env = addOrUpdateEnvVar(c.Env,
			corev1.EnvVar{
				Name: "KAFKA_CERTIFICATE_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: certSecret.Name},
						Key:                  "kafka_keystore_pass",
					},
				},
			},
		)
		c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP", Value: "INTERNAL:SASL_SSL,CONTROLLER:SASL_SSL"})
		c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "KAFKA_CFG_LISTENER_NAME_INTERNAL_SSL_CLIENT_AUTH", Value: "required"})
		c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "KAFKA_CFG_LISTENER_NAME_CONTROLLER_SSL_CLIENT_AUTH", Value: "required"})
		c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "KAFKA_CFG_SASL_MECHANISM_INTER_BROKER_PROTOCOL", Value: "PLAIN"})
		c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "KAFKA_CFG_SASL_MECHANISM_CONTROLLER_PROTOCOL", Value: "PLAIN"})
		c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "KAFKA_CFG_SASL_ENABLED_MECHANISMS", Value: "PLAIN"})
	} else {
		c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP", Value: "INTERNAL:PLAINTEXT,CONTROLLER:PLAINTEXT"})
		c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "ALLOW_PLAINTEXT_LISTENER", Value: "yes"})
	}
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
		Env: []corev1.EnvVar{},
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
		updateKafkaEnv(cr, client, &deployment.Spec.Template.Spec.Containers[0])
		addKafkaStores(cr, deployment, client, "/opt/bitnami/kafka/config/certs")

		return nil
	}

	return deployment, f, nil
}

func ZookeeperDeployment(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*appsv1.Deployment, controllerutil.MutateFn, error) {
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
		},
		Env: []corev1.EnvVar{
			corev1.EnvVar{
				Name:  "ALLOW_ANONYMOUS_LOGIN",
				Value: "yes",
			},
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
		return nil
	}

	return deployment, f, nil
}
