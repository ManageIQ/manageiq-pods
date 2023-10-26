package miqtools

import (
	"context"

	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1"
	miqutilsv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1/miqutils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
				Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments", "deployments/scale"},
				Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"extensions"},
				Resources: []string{"deployments", "deployments/scale"},
				Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
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

func addPostgresConfig(cr *miqv1alpha1.ManageIQ, d *appsv1.Deployment, client client.Client) {
	d.Spec.Template.Spec.Containers[0].Env = addOrUpdateEnvVar(d.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "DATABASE_REGION", Value: cr.Spec.DatabaseRegion})
}

func updateOrchestratorEnv(cr *miqv1alpha1.ManageIQ, c *corev1.Container) {
	c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "ADMIN_GROUP", Value: cr.Spec.InitialAdminGroupName})
	c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "APP_NAME", Value: cr.Spec.AppName})
	c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "APPLICATION_DOMAIN", Value: cr.Spec.ApplicationDomain})
	c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "AUTH_SSO", Value: strconv.FormatBool(*cr.Spec.EnableSSO)})
	c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "AUTH_TYPE", Value: cr.Spec.HttpdAuthenticationType})
	c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "GUID", Value: cr.Spec.ServerGuid})
	c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "LOCAL_LOGIN_ENABLED", Value: strconv.FormatBool(*cr.Spec.EnableApplicationLocalLogin)})
	c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "MEMCACHED_SERVER", Value: "memcached:11211"})
	c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "WORKER_RESOURCES", Value: strconv.FormatBool(*cr.Spec.EnforceWorkerResourceConstraints)})
	c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "WORKER_SERVICE_ACCOUNT", Value: defaultServiceAccountName(cr.Spec.AppName)})

	// If any of the images were not provided, add the orchestrator namespace and tag
	if cr.Spec.BaseWorkerImage == "" || cr.Spec.WebserverWorkerImage == "" || cr.Spec.UIWorkerImage == "" {
		string1 := strings.Split(cr.Spec.OrchestratorImage, ":")
		string2 := strings.Split(string1[0], "/")
		c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "CONTAINER_IMAGE_NAMESPACE", Value: string2[0]})
		c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "CONTAINER_IMAGE_TAG", Value: string1[1]})
	}

	if cr.Spec.BaseWorkerImage != "" {
		c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "BASE_WORKER_IMAGE", Value: cr.Spec.BaseWorkerImage})
	}
	if cr.Spec.WebserverWorkerImage != "" {
		c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "WEBSERVER_WORKER_IMAGE", Value: cr.Spec.WebserverWorkerImage})
	}
	if cr.Spec.UIWorkerImage != "" {
		c.Env = addOrUpdateEnvVar(c.Env, corev1.EnvVar{Name: "UI_WORKER_IMAGE", Value: cr.Spec.UIWorkerImage})
	}
}

func OrchestratorDeployment(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme, client client.Client) (*appsv1.Deployment, controllerutil.MutateFn, error) {
	delaySecs, err := strconv.Atoi(cr.Spec.OrchestratorInitialDelay)
	if err != nil {
		return nil, nil, err
	}
	pullPolicy := corev1.PullIfNotPresent

	if s := strings.Split(cr.Spec.OrchestratorImage, ":"); strings.Contains(s[1], "latest") {
		pullPolicy = corev1.PullAlways
	}

	deploymentLabels := map[string]string{
		"name": "orchestrator",
		"app":  cr.Spec.AppName,
	}

	container := corev1.Container{
		Name:            "orchestrator",
		Image:           cr.Spec.OrchestratorImage,
		ImagePullPolicy: pullPolicy,
		LivenessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
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
		},
	}

	addMessagingEnv(cr, &container)
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

	deployment.Spec.Template.Spec.Containers = []corev1.Container{container}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, deployment, scheme); err != nil {
			return err
		}
		addAppLabel(cr.Spec.AppName, &deployment.ObjectMeta)
		var repNum int32 = 1
		deployment.Spec.Replicas = &repNum
		deployment.Spec.Strategy = appsv1.DeploymentStrategy{
			Type: "Recreate",
		}
		addAnnotations(cr.Spec.AppAnnotations, &deployment.Spec.Template.ObjectMeta)
		var termSecs int64 = 90
		deployment.Spec.Template.Spec.ServiceAccountName = cr.Spec.AppName + "-orchestrator"
		deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = &termSecs

		addPostgresConfig(cr, deployment, client)
		updateOrchestratorEnv(cr, &deployment.Spec.Template.Spec.Containers[0])
		deployment.Spec.Template.Spec.Containers[0].Image = cr.Spec.OrchestratorImage
		deployment.Spec.Template.Spec.Containers[0].SecurityContext = DefaultSecurityContext()

		addInternalRootCertificate(cr, deployment, client)

		certSecret := InternalCertificatesSecret(cr, client)
		if certSecret.Data["api_crt"] != nil && certSecret.Data["api_key"] != nil {
			deployment.Spec.Template.Spec.Containers[0].Env = addOrUpdateEnvVar(deployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "API_SSL_SECRET_NAME", Value: cr.Spec.InternalCertificatesSecret})
		}
		if certSecret.Data["remote_console_crt"] != nil && certSecret.Data["remote_console_key"] != nil {
			deployment.Spec.Template.Spec.Containers[0].Env = addOrUpdateEnvVar(deployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "REMOTE_CONSOLE_SSL_SECRET_NAME", Value: cr.Spec.InternalCertificatesSecret})
		}
		if certSecret.Data["ui_crt"] != nil && certSecret.Data["ui_key"] != nil {
			deployment.Spec.Template.Spec.Containers[0].Env = addOrUpdateEnvVar(deployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "UI_SSL_SECRET_NAME", Value: cr.Spec.InternalCertificatesSecret})
		}

		volumeMount := corev1.VolumeMount{Name: "encryption-key", MountPath: "/run/secrets/manageiq/application", ReadOnly: true}
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = addOrUpdateVolumeMount(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, volumeMount)

		secretVolumeSource := corev1.SecretVolumeSource{SecretName: "app-secrets", Items: []corev1.KeyToPath{corev1.KeyToPath{Key: "encryption-key", Path: "encryption_key"}}}
		deployment.Spec.Template.Spec.Volumes = addOrUpdateVolume(deployment.Spec.Template.Spec.Volumes, corev1.Volume{Name: "encryption-key", VolumeSource: corev1.VolumeSource{Secret: &secretVolumeSource}})

		databaseVolumeMount := corev1.VolumeMount{Name: "database-secret", MountPath: "/run/secrets/postgresql", ReadOnly: true}
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = addOrUpdateVolumeMount(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, databaseVolumeMount)

		databaseSecretVolumeSource := corev1.SecretVolumeSource{SecretName: cr.Spec.DatabaseSecret, Items: []corev1.KeyToPath{
			corev1.KeyToPath{Key: "dbname", Path: "POSTGRESQL_DATABASE"},
			corev1.KeyToPath{Key: "hostname", Path: "POSTGRESQL_HOSTNAME"},
			corev1.KeyToPath{Key: "password", Path: "POSTGRESQL_PASSWORD"},
			corev1.KeyToPath{Key: "port", Path: "POSTGRESQL_PORT"},
			corev1.KeyToPath{Key: "username", Path: "POSTGRESQL_USER"},
		}}
		deployment.Spec.Template.Spec.Volumes = addOrUpdateVolume(deployment.Spec.Template.Spec.Volumes, corev1.Volume{Name: "database-secret", VolumeSource: corev1.VolumeSource{Secret: &databaseSecretVolumeSource}})

		miqutilsv1alpha1.SetDeploymentNodeAffinity(deployment, client)

		return nil
	}

	return deployment, f, nil
}

func orchestratorPod(c client.Client) *corev1.Pod {
	podList := &corev1.PodList{}
	c.List(context.TODO(), podList)

	for _, pod := range podList.Items {
		if pod.ObjectMeta.Labels["name"] == "orchestrator" {
			return &pod
		}
	}

	return nil
}

func addInternalRootCertificate(cr *miqv1alpha1.ManageIQ, d *appsv1.Deployment, client client.Client) {
	secret := InternalCertificatesSecret(cr, client)
	if secret.Data["root_crt"] != nil {
		volumeMount := corev1.VolumeMount{Name: "internal-root-certificate", MountPath: "/etc/pki/ca-trust/source/anchors", ReadOnly: true}
		d.Spec.Template.Spec.Containers[0].VolumeMounts = addOrUpdateVolumeMount(d.Spec.Template.Spec.Containers[0].VolumeMounts, volumeMount)

		secretVolumeSource := corev1.SecretVolumeSource{SecretName: secret.Name, Items: []corev1.KeyToPath{corev1.KeyToPath{Key: "root_crt", Path: "root.crt"}}}
		d.Spec.Template.Spec.Volumes = addOrUpdateVolume(d.Spec.Template.Spec.Volumes, corev1.Volume{Name: "internal-root-certificate", VolumeSource: corev1.VolumeSource{Secret: &secretVolumeSource}})

		d.Spec.Template.Spec.Containers[0].Env = addOrUpdateEnvVar(d.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "SSL_SECRET_NAME", Value: cr.Spec.InternalCertificatesSecret})

		if secret.Data["memcached_crt"] != nil && secret.Data["memcached_key"] != nil {
			d.Spec.Template.Spec.Containers[0].Env = addOrUpdateEnvVar(d.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "MEMCACHED_ENABLE_SSL", Value: "true"})
			d.Spec.Template.Spec.Containers[0].Env = addOrUpdateEnvVar(d.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "MEMCACHED_SSL_CA", Value: "/etc/pki/ca-trust/source/anchors/root.crt"})
		}
	}
}
