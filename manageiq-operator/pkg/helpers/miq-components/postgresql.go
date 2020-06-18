package miqtools

import (
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func DefaultPostgresqlSecret(cr *miqv1alpha1.ManageIQ) *corev1.Secret {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}
	secret := map[string]string{
		"dbname":   "vmdb_production",
		"username": "root",
		"password": generatePassword(),
		"hostname": "postgresql",
		"port":     "5432",
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      postgresqlSecretName(cr),
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		StringData: secret,
	}
}

func postgresqlSecretName(cr *miqv1alpha1.ManageIQ) string {
	secretName := "postgresql-secrets"
	if cr.Spec.DatabaseSecret != "" {
		secretName = cr.Spec.DatabaseSecret
	}

	return secretName
}

func PostgresqlConfigMap(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*corev1.ConfigMap, controllerutil.MutateFn) {
	data := map[string]string{
		"01_miq_overrides.conf": postgresqlOverrideConf(),
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql-configs",
			Namespace: cr.ObjectMeta.Namespace,
		},
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, configMap, scheme); err != nil {
			return err
		}
		addAppLabel(cr.Spec.AppName, &configMap.ObjectMeta)
		configMap.Data = data
		return nil
	}

	return configMap, f
}

func PostgresqlPVC(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*corev1.PersistentVolumeClaim, controllerutil.MutateFn) {
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
			Name:      "postgresql",
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

func PostgresqlService(cr *miqv1alpha1.ManageIQ) (*corev1.Service, controllerutil.MutateFn) {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}

	selector := map[string]string{
		"name": "postgresql",
	}

	ports := []corev1.ServicePort{
		corev1.ServicePort{
			Name: "postgresql",
			Port: 5432,
		},
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql",
			Namespace: cr.ObjectMeta.Namespace,
		},
	}

	return service, serviceMutateFn(service, labels, ports, selector)
}

func PostgresqlDeployment(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*appsv1.Deployment, controllerutil.MutateFn, error) {
	deploymentLabels := map[string]string{
		"name": "postgresql",
		"app":  cr.Spec.AppName,
	}
	var initialDelaySecs int32 = 60

	container := corev1.Container{
		Name:            "postgresql",
		Image:           cr.Spec.PostgresqlImageName + ":" + cr.Spec.PostgresqlImageTag,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Ports: []corev1.ContainerPort{
			corev1.ContainerPort{
				ContainerPort: 5432,
				Protocol:      "TCP",
			},
		},
		ReadinessProbe: &corev1.Probe{
			InitialDelaySeconds: initialDelaySecs,
			Handler: corev1.Handler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(5432),
				},
			},
		},
		Env: []corev1.EnvVar{
			corev1.EnvVar{
				Name: "POSTGRESQL_USER",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: postgresqlSecretName(cr)},
						Key:                  "username",
					},
				},
			},
			corev1.EnvVar{
				Name: "POSTGRESQL_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: postgresqlSecretName(cr)},
						Key:                  "password",
					},
				},
			},
			corev1.EnvVar{
				Name: "POSTGRESQL_DATABASE",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: postgresqlSecretName(cr)},
						Key:                  "dbname",
					},
				},
			},
			corev1.EnvVar{
				Name:  "POSTGRESQL_MAX_CONNECTIONS",
				Value: cr.Spec.PostgresqlMaxConnections,
			},
			corev1.EnvVar{
				Name:  "POSTGRESQL_SHARED_BUFFERS",
				Value: cr.Spec.PostgresqlSharedBuffers,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			corev1.VolumeMount{Name: "miq-pgdb-volume", MountPath: "/var/lib/pgsql/data"},
			corev1.VolumeMount{Name: "miq-pg-configs", MountPath: "/opt/app-root/src/postgresql-cfg/"},
		},
	}

	err := addResourceReqs(cr.Spec.PostgresqlMemoryLimit, cr.Spec.PostgresqlMemoryRequest, cr.Spec.PostgresqlCpuLimit, cr.Spec.PostgresqlCpuRequest, &container)
	if err != nil {
		return nil, nil, err
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql",
			Namespace: cr.ObjectMeta.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: deploymentLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: deploymentLabels,
					Name:   "postgresql",
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
		var repNum int32 = 1
		deployment.Spec.Replicas = &repNum
		deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
		deployment.Spec.Template.Spec.Volumes = []corev1.Volume{
			corev1.Volume{
				Name: "miq-pgdb-volume",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: "postgresql",
					},
				},
			},
			corev1.Volume{
				Name: "miq-pg-configs",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql-configs"},
					},
				},
			},
		}
		return nil
	}

	return deployment, f, nil
}
