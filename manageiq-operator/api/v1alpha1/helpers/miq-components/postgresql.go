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

func ManagePostgresqlSecret(cr *miqv1alpha1.ManageIQ, client client.Client, scheme *runtime.Scheme) (*corev1.Secret, controllerutil.MutateFn) {
	secretKey := types.NamespacedName{Namespace: cr.ObjectMeta.Namespace, Name: cr.Spec.DatabaseSecret}
	secret := &corev1.Secret{}
	secretErr := client.Get(context.TODO(), secretKey, secret)
	if secretErr != nil {
		secret = defaultPostgresqlSecret(cr)
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, secret, scheme); err != nil {
			return err
		}

		addAppLabel(cr.Spec.AppName, &secret.ObjectMeta)
		addBackupLabel(cr.Spec.BackupLabelName, &secret.ObjectMeta)

		if certSecret := InternalCertificatesSecret(cr, client); certSecret.Data["postgresql_crt"] != nil && certSecret.Data["postgresql_key"] != nil && string(secret.Data["hostname"]) == "postgresql" {
			d := map[string]string{
				"rootcertificate": string(certSecret.Data["root_crt"]),
				"sslmode":         "verify-full",
			}
			secret.StringData = d
		}

		return nil
	}

	return secret, f
}

func defaultPostgresqlSecret(cr *miqv1alpha1.ManageIQ) *corev1.Secret {
	secretData := map[string]string{
		"dbname":   "vmdb_production",
		"username": "root",
		"password": generatePassword(),
		"hostname": "postgresql",
		"port":     "5432",
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.DatabaseSecret,
			Namespace: cr.ObjectMeta.Namespace,
		},
		StringData: secretData,
	}

	return secret
}

func postgresqlSecret(cr *miqv1alpha1.ManageIQ, client client.Client) *corev1.Secret {
	secretKey := types.NamespacedName{Namespace: cr.Namespace, Name: cr.Spec.DatabaseSecret}
	secret := &corev1.Secret{}
	client.Get(context.TODO(), secretKey, secret)

	return secret
}

func PostgresqlConfigMap(cr *miqv1alpha1.ManageIQ, client client.Client, scheme *runtime.Scheme) (*corev1.ConfigMap, controllerutil.MutateFn) {
	postgresOverrideConfig := postgresqlOverrideConf()

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

		if configMap.Data == nil {
			configMap.Data = map[string]string{}
		}
		configMap.Data["01_miq_overrides.conf"] = postgresOverrideConfig

		if secret := InternalCertificatesSecret(cr, client); secret.Data["postgresql_crt"] != nil && secret.Data["postgresql_key"] != nil {
			configMap.Data["02_ssl.conf"] = postgresqlSslConf()
		} else {
			delete(configMap.Data, "02_ssl.conf")
		}

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

func PostgresqlService(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*corev1.Service, controllerutil.MutateFn) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql",
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
		service.Spec.Ports[0].Name = "postgresql"
		service.Spec.Ports[0].Port = 5432
		service.Spec.Selector = map[string]string{"name": "postgresql"}
		return nil
	}

	return service, f
}

func PostgresqlDeployment(cr *miqv1alpha1.ManageIQ, client client.Client, scheme *runtime.Scheme) (*appsv1.Deployment, controllerutil.MutateFn, error) {
	deploymentLabels := map[string]string{
		"name": "postgresql",
		"app":  cr.Spec.AppName,
	}
	var initialDelaySecs int32 = 60

	container := corev1.Container{
		Name:            "postgresql",
		Image:           cr.Spec.PostgresqlImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Ports: []corev1.ContainerPort{
			corev1.ContainerPort{
				ContainerPort: 5432,
				Protocol:      "TCP",
			},
		},
		ReadinessProbe: &corev1.Probe{
			InitialDelaySeconds: initialDelaySecs,
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(5432),
				},
			},
		},
		Env: []corev1.EnvVar{
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
		addBackupLabel(cr.Spec.BackupLabelName, &deployment.ObjectMeta)
		addBackupLabel(cr.Spec.BackupLabelName, &deployment.Spec.Template.ObjectMeta)
		addBackupAnnotation("miq-pgdb-volume", &deployment.Spec.Template.ObjectMeta)
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

		volumeMount := corev1.VolumeMount{Name: "env-file", MountPath: "/run/secrets/postgresql", ReadOnly: true}
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = addOrUpdateVolumeMount(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, volumeMount)
		secret := corev1.SecretVolumeSource{
			SecretName: cr.Spec.DatabaseSecret,
			Items: []corev1.KeyToPath{
				corev1.KeyToPath{Key: "dbname", Path: "POSTGRESQL_DATABASE"},
				corev1.KeyToPath{Key: "password", Path: "POSTGRESQL_PASSWORD"},
				corev1.KeyToPath{Key: "username", Path: "POSTGRESQL_USER"},
			},
		}
		deployment.Spec.Template.Spec.Volumes = addOrUpdateVolume(deployment.Spec.Template.Spec.Volumes, corev1.Volume{Name: "env-file", VolumeSource: corev1.VolumeSource{Secret: &secret}})

		addInternalCertificate(cr, deployment, client, "postgresql", "/opt/app-root/src/certificates")

		return nil
	}

	return deployment, f, nil
}
