package miqtools

import (
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strconv"
	"strings"
)

func OrchestratorServiceAccount(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*corev1.ServiceAccount, controllerutil.MutateFn) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      orchestratorObjectName(cr),
			Namespace: cr.ObjectMeta.Namespace,
		},
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, sa, scheme); err != nil {
			return err
		}

		if cr.Spec.ImagePullSecret != "" {
			addSAPullSecret(sa, cr.Spec.ImagePullSecret)
		}

		return nil
	}

	return sa, f
}

func OrchestratorRole(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*rbacv1.Role, controllerutil.MutateFn) {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      orchestratorObjectName(cr),
			Namespace: cr.ObjectMeta.Namespace,
		},
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, role, scheme); err != nil {
			return err
		}

		role.Rules = []rbacv1.PolicyRule{
			rbacv1.PolicyRule{
				APIGroups: []string{""},
				Resources: []string{"pods", "pods/finalizers"},
				Verbs:     []string{"*"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments", "deployments/scale"},
				Verbs:     []string{"*"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"extensions"},
				Resources: []string{"deployments", "deployments/scale"},
				Verbs:     []string{"*"},
			},
		}

		return nil
	}

	return role, f
}

func OrchestratorRoleBinding(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*rbacv1.RoleBinding, controllerutil.MutateFn) {
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      orchestratorObjectName(cr),
			Namespace: cr.ObjectMeta.Namespace,
		},
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, rb, scheme); err != nil {
			return err
		}

		rb.RoleRef = rbacv1.RoleRef{
			Kind:     "Role",
			Name:     orchestratorObjectName(cr),
			APIGroup: "rbac.authorization.k8s.io",
		}
		rb.Subjects = []rbacv1.Subject{
			rbacv1.Subject{
				Kind: "ServiceAccount",
				Name: orchestratorObjectName(cr),
			},
		}

		return nil
	}

	return rb, f
}

func orchestratorObjectName(cr *miqv1alpha1.ManageIQ) string {
	return cr.Spec.AppName + "-orchestrator"
}

func addMessagingEnv(cr *miqv1alpha1.ManageIQ, c *corev1.Container) {
	if !*cr.Spec.DeployMessagingService {
		return
	}

	messagingEnv := []corev1.EnvVar{
		corev1.EnvVar{
			Name: "MESSAGING_HOSTNAME",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: kafkaSecretName(cr)},
					Key:                  "hostname",
				},
			},
		},
		corev1.EnvVar{
			Name: "MESSAGING_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: kafkaSecretName(cr)},
					Key:                  "password",
				},
			},
		},
		corev1.EnvVar{
			Name:  "MESSAGING_PORT",
			Value: "9092",
		},
		corev1.EnvVar{
			Name:  "MESSAGING_TYPE",
			Value: "kafka",
		},
		corev1.EnvVar{
			Name: "MESSAGING_USERNAME",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: kafkaSecretName(cr)},
					Key:                  "username",
				},
			},
		},
	}

	for _, env := range messagingEnv {
		c.Env = append(c.Env, env)
	}

	return
}

func addWorkerImageEnv(cr *miqv1alpha1.ManageIQ, c *corev1.Container) {
	// If any of the images were not provided, add the orchestrator namespace and tag
	if cr.Spec.BaseWorkerImage == "" || cr.Spec.WebserverWorkerImage == "" || cr.Spec.UIWorkerImage == "" {
		c.Env = append(c.Env, corev1.EnvVar{Name: "CONTAINER_IMAGE_NAMESPACE", Value: cr.Spec.OrchestratorImageNamespace})
		c.Env = append(c.Env, corev1.EnvVar{Name: "CONTAINER_IMAGE_TAG", Value: cr.Spec.OrchestratorImageTag})
	}

	if cr.Spec.BaseWorkerImage != "" {
		c.Env = append(c.Env, corev1.EnvVar{Name: "BASE_WORKER_IMAGE", Value: cr.Spec.BaseWorkerImage})
	}
	if cr.Spec.WebserverWorkerImage != "" {
		c.Env = append(c.Env, corev1.EnvVar{Name: "WEBSERVER_WORKER_IMAGE", Value: cr.Spec.WebserverWorkerImage})
	}
	if cr.Spec.UIWorkerImage != "" {
		c.Env = append(c.Env, corev1.EnvVar{Name: "UI_WORKER_IMAGE", Value: cr.Spec.UIWorkerImage})
	}
}

func OrchestratorDeployment(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*appsv1.Deployment, controllerutil.MutateFn, error) {
	delaySecs, err := strconv.Atoi(cr.Spec.OrchestratorInitialDelay)
	if err != nil {
		return nil, nil, err
	}
	pullPolicy := corev1.PullIfNotPresent
	if strings.Contains(cr.Spec.OrchestratorImageTag, "latest") {
		pullPolicy = corev1.PullAlways
	}

	deploymentLabels := map[string]string{
		"name": "orchestrator",
		"app":  cr.Spec.AppName,
	}

	container := corev1.Container{
		Name:            "orchestrator",
		Image:           cr.Spec.OrchestratorImageNamespace + "/" + cr.Spec.OrchestratorImageName + ":" + cr.Spec.OrchestratorImageTag,
		ImagePullPolicy: pullPolicy,
		LivenessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{"pidof", "MIQ Server"},
				},
			},
			InitialDelaySeconds: int32(delaySecs),
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
				Name:  "AUTH_TYPE",
				Value: cr.Spec.HttpdAuthenticationType,
			},
			corev1.EnvVar{
				Name:  "AUTH_SSO",
				Value: strconv.FormatBool(*cr.Spec.EnableSSO),
			},
			corev1.EnvVar{
				Name:  "LOCAL_LOGIN_ENABLED",
				Value: strconv.FormatBool(*cr.Spec.EnableApplicationLocalLogin),
			},
			corev1.EnvVar{
				Name:  "GUID",
				Value: cr.Spec.ServerGuid,
			},
			corev1.EnvVar{
				Name:  "DATABASE_REGION",
				Value: cr.Spec.DatabaseRegion,
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
				Name: "POD_NAME",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
				},
			},
			corev1.EnvVar{
				Name: "POD_UID",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.uid"},
				},
			},
			corev1.EnvVar{
				Name:  "ADMIN_GROUP",
				Value: cr.Spec.InitialAdminGroupName,
			},
			corev1.EnvVar{
				Name:  "WORKER_RESOURCES",
				Value: strconv.FormatBool(*cr.Spec.EnforceWorkerResourceConstraints),
			},
			corev1.EnvVar{
				Name:  "WORKER_SERVICE_ACCOUNT",
				Value: defaultServiceAccountName(cr.Spec.AppName),
			},
		},
	}

	addMessagingEnv(cr, &container)
	addWorkerImageEnv(cr, &container)
	err = addResourceReqs(cr.Spec.OrchestratorMemoryLimit, cr.Spec.OrchestratorMemoryRequest, cr.Spec.OrchestratorCpuLimit, cr.Spec.OrchestratorCpuRequest, &container)
	if err != nil {
		return nil, nil, err
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "orchestrator",
			Namespace: cr.ObjectMeta.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: deploymentLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: deploymentLabels,
					Name:   "orchestrator",
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
		var termSecs int64 = 90
		deployment.Spec.Template.Spec.ServiceAccountName = cr.Spec.AppName + "-orchestrator"
		deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = &termSecs

		return nil
	}

	return deployment, f, nil
}
