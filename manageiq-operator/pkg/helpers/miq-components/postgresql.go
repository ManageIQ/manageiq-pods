package miqtools

import (
	miqv1alpha1 "github.com/manageiq/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

func NewPostgresqlConfigsConfigMap(cr *miqv1alpha1.Manageiq) *corev1.ConfigMap {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql-configs",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"01_miq_overrides.conf": readContentFromFile("pkg/cfres/postgresql_conf/01_miq_overrides.conf"),
		},
	}
}

func NewPostgresqlPVC(cr *miqv1alpha1.Manageiq) *corev1.PersistentVolumeClaim {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}
	storageReq, _ := resource.ParseQuantity(cr.Spec.DatabaseVolumeCapacity)
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql",
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
}

func NewPostgresqlService(cr *miqv1alpha1.Manageiq) *corev1.Service {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}
	selector := map[string]string{
		"name": "postgresql",
	}
	var port int32 = 5432
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name: "postgresql",
					Port: port,
				},
			},
			Selector: selector,
		},
	}
}

func NewPostgresqlDeployment(cr *miqv1alpha1.Manageiq) *appsv1.Deployment {
	DeploymentLabels := map[string]string{
		"app": cr.Spec.AppName,
	}
	PodLabels := map[string]string{
		"name": "postgresql",
		"app":  cr.Spec.AppName,
	}
	var RepNum int32 = 1
	var InitialDelaySecs int32 = 60
	memLimit, _ := resource.ParseQuantity(cr.Spec.PostgresqlMemoryLimit)
	memReq, _ := resource.ParseQuantity(cr.Spec.PostgresqlMemoryRequest)
	cpuReq, _ := resource.ParseQuantity(cr.Spec.PostgresqlCpuRequest)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    DeploymentLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: "Recreate",
			},
			Replicas: &RepNum,
			Selector: &metav1.LabelSelector{
				MatchLabels: PodLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "postgresql",
					Labels: PodLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:  "postgresql",
							Image: cr.Spec.PostgresqlImageName + ":" + cr.Spec.PostgresqlImageTag,
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									ContainerPort: 5432,
									Protocol:      "TCP",
								},
							},
							ReadinessProbe: &corev1.Probe{
								InitialDelaySeconds: InitialDelaySecs,
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
											LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql-secrets"},
											Key:                  "username",
										},
									},
								},
								corev1.EnvVar{
									Name: "POSTGRESQL_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql-secrets"},
											Key:                  "password",
										},
									},
								},
								corev1.EnvVar{
									Name: "POSTGRESQL_DATABASE",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql-secrets"},
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
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									"memory": memLimit,
								},
								Requests: corev1.ResourceList{
									"memory": memReq,
									"cpu":    cpuReq,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{Name: "miq-pgdb-volume", MountPath: "/var/lib/pgsql/data"},
								corev1.VolumeMount{Name: "miq-pg-configs", MountPath: "/opt/app-root/src/postgresql-cfg/"},
							},
						},
					},

					Volumes: []corev1.Volume{
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
					},
				},
			},
		},
	}
}
