package miqtools

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	tlstools "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/helpers/tlstools"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	"net/http"
)

func NewIngress(cr *miqv1alpha1.ManageIQ) *extensionsv1beta1.Ingress {

	labels := map[string]string{
		"app": cr.Spec.AppName,
	}

	return &extensionsv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httpd",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Spec: extensionsv1beta1.IngressSpec{
			TLS: []extensionsv1beta1.IngressTLS{
				extensionsv1beta1.IngressTLS{
					Hosts: []string{
						cr.Spec.ApplicationDomain,
					},
					SecretName: tlsSecretName(cr),
				},
			},
			Rules: []extensionsv1beta1.IngressRule{
				extensionsv1beta1.IngressRule{
					Host: cr.Spec.ApplicationDomain,
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "httpd",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
			},
		},
	}

}

func NewHttpdConfigMap(cr *miqv1alpha1.ManageIQ) (*corev1.ConfigMap, error) {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}

	if cr.Spec.HttpdAuthenticationType == "openid-connect" && cr.Spec.OIDCProviderURL != "" && cr.Spec.OIDCOAuthIntrospectionURL == "" {
		introspectionURL, err := fetchIntrospectionUrl(cr.Spec.OIDCProviderURL)
		if err != nil {
			return nil, err
		}
		cr.Spec.OIDCOAuthIntrospectionURL = introspectionURL
	}

	data := map[string]string{
		"application.conf":    httpdApplicationConf(),
		"authentication.conf": httpdAuthenticationConf(&cr.Spec),
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httpd-configs",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Data: data,
	}

	return configMap, nil
}

func NewHttpdAuthConfigMap(cr *miqv1alpha1.ManageIQ) *corev1.ConfigMap {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}
	data := map[string]string{
		"auth-configuration.conf": httpdAuthConfigurationConf(),
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httpd-auth-configs",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Data: data,
	}
}

func PrivilegedHttpd(authType string) (bool, error) {
	switch authType {
	case "internal", "openid-connect":
		return false, nil
	case "external", "active-directory", "saml":
		return true, nil
	default:
		return false, fmt.Errorf("unknown authenticaion type %s", authType)
	}
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
	podSpec.Containers[0].Env = append(podSpec.Containers[0].Env, clientId)
	podSpec.Containers[0].Env = append(podSpec.Containers[0].Env, clientSecret)
}

func addAuthConfigVolume(podSpec *corev1.PodSpec) {
	vol := corev1.Volume{
		Name: "httpd-auth-config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: "httpd-auth-configs"},
			},
		},
	}
	podSpec.Volumes = append(podSpec.Volumes, vol)

	mount := corev1.VolumeMount{Name: "httpd-auth-config", MountPath: "/etc/httpd/auth-conf.d"}
	podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, mount)
}

func addUserAuthVolume(secretName string, podSpec *corev1.PodSpec) {
	vol := corev1.Volume{
		Name: "user-auth-config",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	}
	podSpec.Volumes = append(podSpec.Volumes, vol)

	mount := corev1.VolumeMount{Name: "user-auth-config", MountPath: "/etc/httpd/user-conf.d"}
	podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, mount)
}

func addOIDCCACertVolume(secretName string, podSpec *corev1.PodSpec) {
	vol := corev1.Volume{
		Name: "oidc-ca-cert",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	}
	podSpec.Volumes = append(podSpec.Volumes, vol)

	mount := corev1.VolumeMount{Name: "oidc-ca-cert", MountPath: "/etc/pki/ca-trust/source/anchors"}
	podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, mount)
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

func httpdImage(namespace, tag string, privileged bool) string {
	var image string
	if privileged {
		image = "httpd-init"
	} else {
		image = "httpd"
	}

	return fmt.Sprintf("%s/%s:%s", namespace, image, tag)
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
	c.Image = httpdImage(spec.HttpdImageNamespace, spec.HttpdImageTag, privileged)
	c.ImagePullPolicy = corev1.PullIfNotPresent
	if privileged {
		c.LivenessProbe = &corev1.Probe{
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{"pidof", "httpd"},
				},
			},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      3,
		}
	}
	c.ReadinessProbe = &corev1.Probe{
		Handler: corev1.Handler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromInt(8080),
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
			PostStart: &corev1.Handler{
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

func NewHttpdDeployment(cr *miqv1alpha1.ManageIQ) (*appsv1.Deployment, error) {
	privileged, err := PrivilegedHttpd(cr.Spec.HttpdAuthenticationType)
	if err != nil {
		return nil, err
	}

	container := corev1.Container{}
	err = initializeHttpdContainer(&cr.Spec, privileged, &container)
	if err != nil {
		return nil, err
	}

	deploymentLabels := map[string]string{
		"app": cr.Spec.AppName,
	}
	podLabels := map[string]string{
		"name": "httpd",
		"app":  cr.Spec.AppName,
	}
	var repNum int32 = 1

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httpd",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    deploymentLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &repNum,
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "httpd",
					Labels: podLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{container},
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: "httpd-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "httpd-configs"},
								},
							},
						},
					},
				},
			},
		},
	}

	configureHttpdAuth(&cr.Spec, &deployment.Spec.Template.Spec)

	// Only assign the service account if we need additional privileges
	if privileged {
		deployment.Spec.Template.Spec.ServiceAccountName = cr.Spec.AppName + "-httpd"
	} else {
		deployment.Spec.Template.Spec.ServiceAccountName = defaultServiceAccountName(cr.Spec.AppName)
	}

	return deployment, nil
}

func NewUIService(cr *miqv1alpha1.ManageIQ) *corev1.Service {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}
	selector := map[string]string{
		"service": "ui",
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ui",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name: "web-service-3000",
					Port: 3000,
				},
			},
			Selector: selector,
		},
	}
}

func NewWebService(cr *miqv1alpha1.ManageIQ) *corev1.Service {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}
	selector := map[string]string{
		"service": "web-service",
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web-service",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name: "web-service-3000",
					Port: 3000,
				},
			},
			Selector: selector,
		},
	}
}

func NewRemoteConsoleService(cr *miqv1alpha1.ManageIQ) *corev1.Service {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}
	selector := map[string]string{
		"service": "remote-console",
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "remote-console",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name: "remote-console-3000",
					Port: 80,
				},
			},
			Selector: selector,
		},
	}
}

func NewHttpdService(cr *miqv1alpha1.ManageIQ) *corev1.Service {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}

	selector := map[string]string{
		"name": "httpd",
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httpd",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name: "http",
					Port: 8080,
				},
			},
			Selector: selector,
		},
	}
}

func NewHttpdDbusAPIService(cr *miqv1alpha1.ManageIQ) *corev1.Service {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}

	selector := map[string]string{
		"name": "httpd",
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httpd-dbus-api",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name: "http-dbus-api",
					Port: 8081,
				},
			},
			Selector: selector,
		},
	}
}

func TLSSecret(cr *miqv1alpha1.ManageIQ) (*corev1.Secret, error) {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}

	crt, key, err := tlstools.GenerateCrt("server")
	if err != nil {
		return nil, err
	}

	data := map[string]string{
		"tls.crt": string(crt),
		"tls.key": string(key),
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tlsSecretName(cr),
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		StringData: data,
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

func fetchIntrospectionUrl(providerUrl string) (string, error) {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
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

	if result["token_introspection_endpoint"] == nil {
		return "", fmt.Errorf("%s - token_introspection_endpoint is missing from the Provider metadata", errMsg)
	}

	return result["token_introspection_endpoint"].(string), nil
}
