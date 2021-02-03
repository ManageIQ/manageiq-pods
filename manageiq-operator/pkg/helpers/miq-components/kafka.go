package miqtools

import (
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func DefaultKafkaSecret(cr *miqv1alpha1.ManageIQ) *corev1.Secret {
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
		service.Spec.Ports[0].Name = "kafka"
		service.Spec.Ports[0].Port = 9092
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

func KafkaDeployment(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*appsv1.Deployment, controllerutil.MutateFn, error) {
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
				Spec: corev1.PodSpec{},
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
		if len(cr.Spec.AppAnnotations) > 0 {
			deployment.Spec.Template.ObjectMeta.Annotations = cr.Spec.AppAnnotations
		}
		deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
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
