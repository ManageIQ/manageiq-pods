package miqtools

import (
	miqv1alpha1 "github.com/manageiq/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/google/uuid"
)

func NewOrchestratorDeployment(cr *miqv1alpha1.Manageiq) *appsv1.Deployment {
	deploymentLabels := map[string]string{
		"app": cr.Spec.AppName,
	}

	podLabels := map[string]string{
		"name": "orchestrator",
		"app":  cr.Spec.AppName,
	}

	var repNum int32 = 1
	var termSecs int64 = 90
	memLimit, _ := resource.ParseQuantity(cr.Spec.OrchestratorMemoryLimit)
	memReq, _ := resource.ParseQuantity(cr.Spec.OrchestratorMemoryRequest)
	cpuReq, _ := resource.ParseQuantity(cr.Spec.OrchestratorCpuRequest)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "orchestrator",
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
					Name:   "orchestrator",
					Labels: podLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:  "orchestrator",
							Image: cr.Spec.OrchestratorImageNamespace + "/" + cr.Spec.OrchestratorImageName + ":" + cr.Spec.OrchestratorImageTag,
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{"pidof", "MIQ Server"},
									},
								},
								InitialDelaySeconds: 480,
								TimeoutSeconds:      3,
							},
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name:  "ALLOW_INSECURE_SESSION",
									Value: "true",
								},
								corev1.EnvVar{
									Name:  "APP_NAME",
									Value: cr.Spec.AppName,
								},
								corev1.EnvVar{
									Name: "APPLICATION_ADMIN_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: "app-secrets"},
											Key:                  "admin-password",
										},
									},
								},
								corev1.EnvVar{
									Name:  "GUID",
									Value: uuid.New().String(),
								},

								corev1.EnvVar{
									Name:  "DATABASE_REGION",
									Value: "0",
								},
								corev1.EnvVar{
									Name: "DATABASE_HOSTNAME",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: postgresqlSecretName(cr)},
											Key:                  "hostname",
										},
									},
								},
								corev1.EnvVar{
									Name: "DATABASE_NAME",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: postgresqlSecretName(cr)},
											Key:                  "dbname",
										},
									},
								},
								corev1.EnvVar{
									Name: "DATABASE_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: postgresqlSecretName(cr)},
											Key:                  "password",
										},
									},
								},
								corev1.EnvVar{
									Name: "DATABASE_PORT",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: postgresqlSecretName(cr)},
											Key:                  "port",
										},
									},
								},
								corev1.EnvVar{
									Name: "DATABASE_USER",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: postgresqlSecretName(cr)},
											Key:                  "username",
										},
									},
								},
								corev1.EnvVar{
									Name: "ENCRYPTION_KEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: "app-secrets"},
											Key:                  "encryption-key",
										},
									},
								},
								corev1.EnvVar{
									Name:  "CONTAINER_IMAGE_NAMESPACE",
									Value: cr.Spec.OrchestratorImageNamespace,
								},
								corev1.EnvVar{
									Name:  "IMAGE_PULL_SECRET",
									Value: "",
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
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						corev1.LocalObjectReference{Name: ""},
					},
					TerminationGracePeriodSeconds: &termSecs,

					ServiceAccountName: cr.Spec.AppName + "-orchestrator",
				},
			},
		},
	}
}
