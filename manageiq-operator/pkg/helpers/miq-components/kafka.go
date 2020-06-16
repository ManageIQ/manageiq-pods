package miqtools

import (
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DefaultKafkaSecret(cr *miqv1alpha1.ManageIQ) *corev1.Secret {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}
	secret := map[string]string{
		"username": "root",
		"password": generatePassword(),
		"hostname": "kafka",
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kafkaSecretName(cr),
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		StringData: secret,
	}
}

func kafkaSecretName(cr *miqv1alpha1.ManageIQ) string {
	secretName := "kafka-secrets"
	if cr.Spec.KafkaSecret != "" {
		secretName = cr.Spec.KafkaSecret
	}

	return secretName
}

func KafkaPVC(cr *miqv1alpha1.ManageIQ) *corev1.PersistentVolumeClaim {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}
	storageReq, _ := resource.ParseQuantity(cr.Spec.KafkaVolumeCapacity)
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kafka-data",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": storageReq,
				},
			},
		},
	}

	if cr.Spec.StorageClassName != "" {
		pvc.Spec.StorageClassName = &cr.Spec.StorageClassName
	}

	return pvc
}

func ZookeeperPVC(cr *miqv1alpha1.ManageIQ) *corev1.PersistentVolumeClaim {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}
	storageReq, _ := resource.ParseQuantity(cr.Spec.ZookeeperVolumeCapacity)
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "zookeeper-data",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": storageReq,
				},
			},
		},
	}

	if cr.Spec.StorageClassName != "" {
		pvc.Spec.StorageClassName = &cr.Spec.StorageClassName
	}

	return pvc
}

func KafkaService(cr *miqv1alpha1.ManageIQ) *corev1.Service {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}
	selector := map[string]string{
		"name": "kafka",
	}
	var port int32 = 9092
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kafka",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name: "kafka",
					Port: port,
				},
			},
			Selector: selector,
		},
	}
}

func ZookeeperService(cr *miqv1alpha1.ManageIQ) *corev1.Service {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}
	selector := map[string]string{
		"name": "zookeeper",
	}
	var port int32 = 2181
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "zookeeper",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name: "zookeeper",
					Port: port,
				},
			},
			Selector: selector,
		},
	}
}

func KafkaDeployment(cr *miqv1alpha1.ManageIQ) (*appsv1.Deployment, error) {
	container := corev1.Container{
		Name:            "kafka",
		Image:           cr.Spec.KafkaImageName + ":" + cr.Spec.KafkaImageTag,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Ports: []corev1.ContainerPort{
			corev1.ContainerPort{
				ContainerPort: 9092,
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
			corev1.EnvVar{
				Name:  "KAFKA_ZOOKEEPER_CONNECT",
				Value: "zookeeper:2181",
			},
			corev1.EnvVar{
				Name:  "ALLOW_PLAINTEXT_LISTENER",
				Value: "yes",
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			corev1.VolumeMount{Name: "kafka-data", MountPath: "/bitnami/kafka"},
		},
	}

	err := addResourceReqs(cr.Spec.KafkaMemoryLimit, cr.Spec.KafkaMemoryRequest, cr.Spec.KafkaCpuLimit, cr.Spec.KafkaCpuRequest, &container)
	if err != nil {
		return nil, err
	}

	deploymentLabels := map[string]string{
		"app": cr.Spec.AppName,
	}
	podLabels := map[string]string{
		"name": "kafka",
		"app":  cr.Spec.AppName,
	}
	var repNum int32 = 1
	var termSecs int64 = 10

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kafka",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    deploymentLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: "Recreate",
			},
			Replicas: &repNum,
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "kafka",
					Labels: podLabels,
				},
				Spec: corev1.PodSpec{
					Containers:                    []corev1.Container{container},
					TerminationGracePeriodSeconds: &termSecs,
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: "kafka-data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "kafka-data",
								},
							},
						},
					},
				},
			},
		},
	}

	return deployment, nil
}

func ZookeeperDeployment(cr *miqv1alpha1.ManageIQ) (*appsv1.Deployment, error) {
	container := corev1.Container{
		Name:            "zookeeper",
		Image:           cr.Spec.ZookeeperImageName + ":" + cr.Spec.ZookeeperImageTag,
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
		return nil, err
	}

	deploymentLabels := map[string]string{
		"app": cr.Spec.AppName,
	}
	podLabels := map[string]string{
		"name": "zookeeper",
		"app":  cr.Spec.AppName,
	}
	var repNum int32 = 1

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "zookeeper",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    deploymentLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: "Recreate",
			},
			Replicas: &repNum,
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "zookeeper",
					Labels: podLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{container},
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: "zookeeper-data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "zookeeper-data",
								},
							},
						},
					},
				},
			},
		},
	}

	return deployment, nil
}
