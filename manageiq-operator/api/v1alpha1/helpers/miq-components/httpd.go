package miqtools

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"

	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1"
	tlstools "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1/helpers/tlstools"
	miqutilsv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1/miqutils"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func HttpdServiceAccount(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*corev1.ServiceAccount, controllerutil.MutateFn) {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.AppName + "-httpd",
			Namespace: cr.ObjectMeta.Namespace,
		},
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, serviceAccount, scheme); err != nil {
			return err
		}

		if cr.Spec.ImagePullSecret != "" {
			addSAPullSecret(serviceAccount, cr.Spec.ImagePullSecret)
		}

		return nil
	}

	return serviceAccount, f
}

func Route(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme, client client.Client) (*routev1.Route, controllerutil.MutateFn) {
	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httpd",
			Namespace: cr.ObjectMeta.Namespace,
		},
		Spec: routev1.RouteSpec{
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("http"),
			},
			TLS: &routev1.TLSConfig{
				InsecureEdgeTerminationPolicy: "Redirect",
				Termination:                   "reencrypt",
			},
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: "httpd",
			},
			WildcardPolicy: routev1.WildcardPolicyNone,
		},
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, route, scheme); err != nil {
			return err
		}

		route.Spec.Host = cr.Spec.ApplicationDomain
		// This removes the certificate that we previously set on the route and prevents anyone from setting their own.
		// Removing the certificate on our route will cause it to use the cluster default certificate.
		route.Spec.TLS.Certificate = ""
		route.Spec.TLS.Key = ""

		if internalCerts := InternalCertificatesSecret(cr, client); internalCerts.Data["httpd_crt"] != nil {
			route.Spec.TLS.DestinationCACertificate = string(internalCerts.Data["root_crt"])
			route.Spec.TLS.Termination = "reencrypt"
		} else {
			route.Spec.TLS.Termination = "edge"
		}

		return nil
	}

	return route, f
}

func tlsSecret(cr *miqv1alpha1.ManageIQ, client client.Client) *corev1.Secret {
	name := tlsSecretName(cr)

	secretKey := types.NamespacedName{Namespace: cr.Namespace, Name: name}
	secret := &corev1.Secret{}
	client.Get(context.TODO(), secretKey, secret)

	return secret
}

func Ingress(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*networkingv1.Ingress, controllerutil.MutateFn) {
	implementationSpecific := networkingv1.PathType("ImplementationSpecific")
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httpd",
			Namespace: cr.ObjectMeta.Namespace,
		},
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, ingress, scheme); err != nil {
			return err
		}
		if len(ingress.Spec.TLS) == 0 {
			ingress.Spec.TLS = append(ingress.Spec.TLS, networkingv1.IngressTLS{})
		}
		if len(ingress.Spec.TLS[0].Hosts) == 0 {
			ingress.Spec.TLS[0].Hosts = append(ingress.Spec.TLS[0].Hosts, cr.Spec.ApplicationDomain)
		}
		ingress.Spec.TLS[0].Hosts[0] = cr.Spec.ApplicationDomain
		ingress.Spec.TLS[0].SecretName = tlsSecretName(cr)
		ingress.Spec.Rules = []networkingv1.IngressRule{
			networkingv1.IngressRule{
				Host: cr.Spec.ApplicationDomain,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								Path:     "/",
								PathType: &implementationSpecific,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: "httpd",
										Port: networkingv1.ServiceBackendPort{
											Number: 8080,
										},
									},
								},
							},
						},
					},
				},
			},
		}
		addAppLabel(cr.Spec.AppName, &ingress.ObjectMeta)
		return nil
	}

	return ingress, f
}

func HttpdConfigMap(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme, client client.Client) (*corev1.ConfigMap, controllerutil.MutateFn, error) {
	if cr.Spec.HttpdAuthenticationType == "openid-connect" && cr.Spec.OIDCProviderURL != "" && cr.Spec.OIDCOAuthIntrospectionURL == "" {
		introspectionURL, err := fetchIntrospectionUrl(cr.Spec.OIDCProviderURL, *cr.Spec.OIDCOAuthIntrospectionSSLVerify)
		if err != nil {
			return nil, nil, err
		}
		cr.Spec.OIDCOAuthIntrospectionURL = introspectionURL
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httpd-configs",
			Namespace: cr.ObjectMeta.Namespace,
		},
		Data: make(map[string]string),
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, configMap, scheme); err != nil {
			return err
		}
		addAppLabel(cr.Spec.AppName, &configMap.ObjectMeta)

		uiHttpProtocol, uiWebSocketProtocol := "http", "ws"
		if certSecret := InternalCertificatesSecret(cr, client); certSecret.Data["ui_crt"] != nil && certSecret.Data["ui_key"] != nil {
			uiHttpProtocol, uiWebSocketProtocol = "https", "wss"
		}

		apiHttpProtocol := "http"
		if certSecret := InternalCertificatesSecret(cr, client); certSecret.Data["api_crt"] != nil && certSecret.Data["api_key"] != nil {
			apiHttpProtocol = "https"
		}

		configMap.Data["application.conf"] = httpdApplicationConf(cr.Spec.ApplicationDomain, uiHttpProtocol, uiWebSocketProtocol, apiHttpProtocol)
		configMap.Data["authentication.conf"] = httpdAuthenticationConf(&cr.Spec)
		configMap.Data["health.conf"] = httpdHealthConf()

		if certSecret := InternalCertificatesSecret(cr, client); certSecret.Data["httpd_crt"] != nil && certSecret.Data["httpd_key"] != nil {
			configMap.Data["ssl_config"] = httpdSslConfig()
		}

		if certSecret := InternalCertificatesSecret(cr, client); certSecret.Data["ui_crt"] != nil && certSecret.Data["ui_key"] != nil {
			configMap.Data["ssl_proxy_config"] = httpdSslProxyConfig()
		}

		return nil
	}

	return configMap, f, nil
}

func HttpdAuthConfigMap(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*corev1.ConfigMap, controllerutil.MutateFn) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httpd-auth-configs",
			Namespace: cr.ObjectMeta.Namespace,
		},
		Data: make(map[string]string),
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, configMap, scheme); err != nil {
			return err
		}
		addAppLabel(cr.Spec.AppName, &configMap.ObjectMeta)
		addBackupLabel(cr.Spec.BackupLabelName, &configMap.ObjectMeta)

		configMap.Data["auth-configuration.conf"] = httpdAuthConfigurationConf()

		return nil
	}

	return configMap, f
}

func HttpdAuthConfig(client client.Client, cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*corev1.Secret, controllerutil.MutateFn) {
	if cr.Spec.HttpdAuthConfig == "" {
		return nil, nil
	}

	secret := &corev1.Secret{}
	if secretErr := client.Get(context.TODO(), types.NamespacedName{Namespace: cr.Namespace, Name: cr.Spec.HttpdAuthConfig}, secret); secretErr != nil {
		return nil, nil
	}

	f := func() error {
		addBackupLabel(cr.Spec.BackupLabelName, &secret.ObjectMeta)
		return nil
	}

	return secret, f
}

func PrivilegedHttpd(authType string) bool {
	switch authType {
	case "internal", "openid-connect":
		return false
	case "external", "active-directory", "saml":
		return true
	}
	return false
}

func addOIDCEnv(secretName string, podSpec *corev1.PodSpec) {
	clientId := corev1.EnvVar{
		Name: "HTTPD_AUTH_OIDC_CLIENT_ID",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
				Key:                  "CLIENT_ID",
			},
		},
	}
	clientSecret := corev1.EnvVar{
		Name: "HTTPD_AUTH_OIDC_CLIENT_SECRET",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
				Key:                  "CLIENT_SECRET",
			},
		},
	}

	podSpec.Containers[0].Env = addOrUpdateEnvVar(podSpec.Containers[0].Env, clientId)
	podSpec.Containers[0].Env = addOrUpdateEnvVar(podSpec.Containers[0].Env, clientSecret)
}

func getHttpdAuthConfigVersion(client client.Client, namespace string, spec *miqv1alpha1.ManageIQSpec) string {
	httpd_auth_config_version := ""
	if spec.HttpdAuthConfig != "" {
		secret := &corev1.Secret{}
		secretErr := client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: spec.HttpdAuthConfig}, secret)
		if secretErr == nil {
			httpd_auth_config_version = string(secret.GetObjectMeta().GetResourceVersion())
		}
	}
	return httpd_auth_config_version
}

func addAuthConfigVolume(podSpec *corev1.PodSpec) {
	volumeMount := corev1.VolumeMount{Name: "httpd-auth-config", MountPath: "/etc/httpd/auth-conf.d"}
	podSpec.Containers[0].VolumeMounts = addOrUpdateVolumeMount(podSpec.Containers[0].VolumeMounts, volumeMount)

	configMapVolumeSource := corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: "httpd-auth-configs"}}
	podSpec.Volumes = addOrUpdateVolume(podSpec.Volumes, corev1.Volume{Name: "httpd-auth-config", VolumeSource: corev1.VolumeSource{ConfigMap: &configMapVolumeSource}})
}

func addUserAuthVolume(secretName string, podSpec *corev1.PodSpec) {
	volumeMount := corev1.VolumeMount{Name: "user-auth-config", MountPath: "/etc/httpd/user-conf.d"}
	podSpec.Containers[0].VolumeMounts = addOrUpdateVolumeMount(podSpec.Containers[0].VolumeMounts, volumeMount)

	secretVolumeSource := corev1.SecretVolumeSource{SecretName: secretName}
	podSpec.Volumes = addOrUpdateVolume(podSpec.Volumes, corev1.Volume{Name: "user-auth-config", VolumeSource: corev1.VolumeSource{Secret: &secretVolumeSource}})
}

func addOIDCCACertVolume(secretName string, podSpec *corev1.PodSpec) {
	volumeMount := corev1.VolumeMount{Name: "oidc-ca-cert", MountPath: "/etc/pki/ca-trust/source/anchors"}
	podSpec.Containers[0].VolumeMounts = addOrUpdateVolumeMount(podSpec.Containers[0].VolumeMounts, volumeMount)

	secretVolumeSource := corev1.SecretVolumeSource{SecretName: secretName}
	podSpec.Volumes = addOrUpdateVolume(podSpec.Volumes, corev1.Volume{Name: "oidc-ca-cert", VolumeSource: corev1.VolumeSource{Secret: &secretVolumeSource}})
}

func configureHttpdAuth(spec *miqv1alpha1.ManageIQSpec, podSpec *corev1.PodSpec) {
	authType := spec.HttpdAuthenticationType

	if authType == "internal" {
		return
	}

	if spec.HttpdAuthConfig != "" {
		addUserAuthVolume(spec.HttpdAuthConfig, podSpec)
	}

	if spec.OIDCCACertSecret != "" {
		addOIDCCACertVolume(spec.OIDCCACertSecret, podSpec)
	}

	if authType == "openid-connect" && spec.OIDCClientSecret != "" {
		addOIDCEnv(spec.OIDCClientSecret, podSpec)
	} else if authType != "openid-connect" {
		addAuthConfigVolume(podSpec)
	}
}

func assignHttpdPorts(privileged bool, c *corev1.Container) {
	httpdPort := corev1.ContainerPort{
		ContainerPort: 8080,
		Protocol:      "TCP",
	}
	c.Ports = append(c.Ports, httpdPort)

	if privileged {
		dbusApiPort := corev1.ContainerPort{
			ContainerPort: 8081,
			Protocol:      "TCP",
		}
		c.Ports = append(c.Ports, dbusApiPort)
	}
}

func initializeHttpdContainer(spec *miqv1alpha1.ManageIQSpec, privileged bool, c *corev1.Container) error {
	c.Name = "httpd"
	c.Image = spec.HttpdImage
	c.ImagePullPolicy = corev1.PullIfNotPresent
	c.LivenessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/health/healthz",
				Port: intstr.FromInt(8081),
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      3,
	}
	c.ReadinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/health/healthz",
				Port: intstr.FromInt(8081),
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      3,
	}
	c.VolumeMounts = []corev1.VolumeMount{
		corev1.VolumeMount{Name: "httpd-config", MountPath: "/etc/httpd/conf.d"},
	}

	// Add Lifecycle object for saving the environment if we're running with init
	if privileged {
		c.Lifecycle = &corev1.Lifecycle{
			PostStart: &corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/usr/bin/save-container-environment"},
				},
			},
		}
	}

	assignHttpdPorts(privileged, c)

	err := addResourceReqs(spec.HttpdMemoryLimit, spec.HttpdMemoryRequest, spec.HttpdCpuLimit, spec.HttpdCpuRequest, c)
	if err != nil {
		return err
	}

	return nil
}

func HttpdDeployment(client client.Client, cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*appsv1.Deployment, controllerutil.MutateFn, error) {
	privileged := PrivilegedHttpd(cr.Spec.HttpdAuthenticationType)

	container := corev1.Container{}
	err := initializeHttpdContainer(&cr.Spec, privileged, &container)
	if err != nil {
		return nil, nil, err
	}

	deploymentLabels := map[string]string{
		"app":  cr.Spec.AppName,
		"name": "httpd",
	}
	deploymentSelectorLabels := map[string]string{}
	maps.Copy(deploymentSelectorLabels, deploymentLabels)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httpd",
			Namespace: cr.ObjectMeta.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: deploymentSelectorLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: deploymentLabels,
					Name:   "httpd",
				},
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
		deployment.Spec.Strategy = appsv1.DeploymentStrategy{
			Type: "Recreate",
		}
		addAnnotations(cr.Spec.AppAnnotations, &deployment.Spec.Template.ObjectMeta)
		deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
		deployment.Spec.Template.Spec.Containers[0].SecurityContext = DefaultSecurityContext()

		configMapVolumeSource := corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: "httpd-configs"}}
		deployment.Spec.Template.Spec.Volumes = addOrUpdateVolume(deployment.Spec.Template.Spec.Volumes, corev1.Volume{Name: "httpd-config", VolumeSource: corev1.VolumeSource{ConfigMap: &configMapVolumeSource}})

		// Only assign the service account if we need additional privileges
		if privileged {
			deployment.Spec.Template.Spec.ServiceAccountName = cr.Spec.AppName + "-httpd"
		} else {
			deployment.Spec.Template.Spec.ServiceAccountName = defaultServiceAccountName(cr.Spec.AppName)
		}

		configureHttpdAuth(&cr.Spec, &deployment.Spec.Template.Spec)

		// This is not used by the pod, it is defined to trigger a redeployment if the secret was updated
		httpdAuthConfigVersion := getHttpdAuthConfigVersion(client, cr.Namespace, &cr.Spec)
		deployment.Spec.Template.Spec.Containers[0].Env = addOrUpdateEnvVar(deployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "MANAGED_HTTPD_CFG_VERSION", Value: httpdAuthConfigVersion})

		addInternalCertificate(cr, deployment, client, "httpd", "/root")

		secret := InternalCertificatesSecret(cr, client)
		if secret.Data["root_crt"] != nil {
			volumeName := "internal-root-certificate"
			volumeMount := corev1.VolumeMount{Name: volumeName, MountPath: "/etc/pki/ca-trust/source/anchors", ReadOnly: true}
			deployment.Spec.Template.Spec.Containers[0].VolumeMounts = addOrUpdateVolumeMount(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, volumeMount)

			secretVolumeSource := corev1.SecretVolumeSource{SecretName: secret.Name, Items: []corev1.KeyToPath{corev1.KeyToPath{Key: "root_crt", Path: "root.crt"}}}
			deployment.Spec.Template.Spec.Volumes = addOrUpdateVolume(deployment.Spec.Template.Spec.Volumes, corev1.Volume{Name: volumeName, VolumeSource: corev1.VolumeSource{Secret: &secretVolumeSource}})
		}

		miqutilsv1alpha1.SetDeploymentNodeAffinity(deployment, client)

		return nil
	}

	return deployment, f, nil
}

func UIService(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*corev1.Service, controllerutil.MutateFn) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ui",
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
		service.Spec.Ports[0].Name = "ui-service-3000"
		service.Spec.Ports[0].Port = 3000
		service.Spec.Selector = map[string]string{"service": "ui"}
		return nil
	}

	return service, f
}

func WebService(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*corev1.Service, controllerutil.MutateFn) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web-service",
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
		service.Spec.Ports[0].Name = "web-service-3000"
		service.Spec.Ports[0].Port = 3000
		service.Spec.Selector = map[string]string{"service": "web-service"}
		return nil
	}

	return service, f
}

func RemoteConsoleService(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*corev1.Service, controllerutil.MutateFn) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "remote-console",
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
		service.Spec.Ports[0].Name = "remote-console-3000"
		service.Spec.Ports[0].Port = 3000
		service.Spec.Selector = map[string]string{"service": "remote-console"}
		return nil
	}

	return service, f
}

func HttpdService(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*corev1.Service, controllerutil.MutateFn) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httpd",
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
		service.Spec.Ports[0].Name = "http"
		service.Spec.Ports[0].Port = 8080
		service.Spec.Selector = map[string]string{"name": "httpd"}
		return nil
	}

	return service, f
}

func HttpdDbusAPIService(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*corev1.Service, controllerutil.MutateFn) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httpd-dbus-api",
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
		service.Spec.Ports[0].Name = "httpd-dbus-api"
		service.Spec.Ports[0].Port = 8081
		service.Spec.Selector = map[string]string{"name": "httpd"}
		return nil
	}

	return service, f
}

func ManageTlsSecret(cr *miqv1alpha1.ManageIQ, client client.Client, scheme *runtime.Scheme) (*corev1.Secret, controllerutil.MutateFn, error) {
	secretKey := types.NamespacedName{Namespace: cr.ObjectMeta.Namespace, Name: tlsSecretName(cr)}
	secret := &corev1.Secret{}
	secretErr := client.Get(context.TODO(), secretKey, secret)
	var err error
	if secretErr != nil {
		secret, err = defaultTLSSecret(cr)
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, secret, scheme); err != nil {
			return err
		}

		addAppLabel(cr.Spec.AppName, &secret.ObjectMeta)
		addBackupLabel(cr.Spec.BackupLabelName, &secret.ObjectMeta)

		return nil
	}

	return secret, f, err
}

func defaultTLSSecret(cr *miqv1alpha1.ManageIQ) (*corev1.Secret, error) {
	crt, key, err := tlstools.GenerateCrt(cr.Spec.ApplicationDomain)
	if err != nil {
		return nil, err
	}

	secretData := map[string]string{
		"tls.crt": string(crt),
		"tls.key": string(key),
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tlsSecretName(cr),
			Namespace: cr.ObjectMeta.Namespace,
		},
		StringData: secretData,
		Type:       "kubernetes.io/tls",
	}

	return secret, nil
}

func tlsSecretName(cr *miqv1alpha1.ManageIQ) string {
	secretName := "tls-secret"
	if cr.Spec.TLSSecret != "" {
		secretName = cr.Spec.TLSSecret
	}

	return secretName
}

func fetchIntrospectionUrl(providerUrl string, sslVerify bool) (string, error) {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: !sslVerify}
	client := &http.Client{Transport: customTransport}
	errMsg := fmt.Sprintf("failed to get the OIDCOAuthIntrospectionURL from %s", providerUrl)

	resp, err := client.Get(providerUrl)
	if err != nil {
		return "", fmt.Errorf("%s - %s", errMsg, err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("%s - StatusCode: %d", errMsg, resp.StatusCode)
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("%s - %s", errMsg, err)
	}

	if result["introspection_endpoint"] == nil {
		return "", fmt.Errorf("%s - introspection_endpoint is missing from the Provider metadata", errMsg)
	}

	return result["introspection_endpoint"].(string), nil
}
