package miqtools

import (
	"bytes"
	miqv1alpha1 "github.com/manageiq/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	tlstools "github.com/manageiq/manageiq-pods/manageiq-operator/pkg/helpers/tlstools"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	"os"
)

func NewIngress(cr *miqv1alpha1.Manageiq) *extensionsv1beta1.Ingress {

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
					SecretName: "tls-secret",
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
										ServicePort: intstr.FromInt(80),
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

func readContentFromFile(filename string) string {
	filerc, _ := os.Open(filename)

	defer filerc.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(filerc)
	contents := buf.String()
	return contents
}

func NewHttpdConfigMap(cr *miqv1alpha1.Manageiq) *corev1.ConfigMap {

	labels := map[string]string{
		"app": cr.Spec.AppName,
	}

	//pwd, _ := os.Getwd()
	appconf := readContentFromFile("pkg/resources/httpd_conf/application.conf")
	authconf := readContentFromFile("pkg/resources/httpd_conf/authentication.conf")
	configurationInternalAuth := readContentFromFile("pkg/resources/httpd_conf/configuration-internal-auth.conf")
	configurationExternalAuth := readContentFromFile("pkg/resources/httpd_conf/configuration-external-auth.conf")
	configurationActiveDirectoryAuth := readContentFromFile("pkg/resources/httpd_conf/configuration-active-directory-auth")
	configurationSamlAuth := readContentFromFile("pkg/resources/httpd_conf/configuration-saml-auth")
	configurationOpenidConnectAuth := readContentFromFile("pkg/resources/httpd_conf/configuration-openid-connect-auth")

	externalAuthLoadModulesConf := readContentFromFile("pkg/resources/httpd_conf/external-auth-load-modules-conf")
	externalAuthLoginFormConf := readContentFromFile("pkg/resources/httpd_conf/external-auth-login-form-conf")
	externalAuthApplicationApiConf := readContentFromFile("pkg/resources/httpd_conf/external-auth-application-api-conf")
	externalAuthLookupUserDetailsConf := readContentFromFile("pkg/resources/httpd_conf/external-auth-lookup-user-details-conf")
	externalAuthRemoteUserConf := readContentFromFile("pkg/resources/httpd_conf/external-auth-remote-user-conf")
	externalAuthOpenidConnectRemoteUserConf := readContentFromFile("pkg/resources/httpd_conf/external-auth-openid-connect-remote-user-conf")

	data := map[string]string{
		"application.conf":                              string(appconf),
		"authentication.conf":                           string(authconf),
		"configuration-internal-auth":                   string(configurationInternalAuth),
		"configuration-external-auth":                   string(configurationExternalAuth),
		"configuration-active-directory-auth":           string(configurationActiveDirectoryAuth),
		"configuration-saml-auth":                       string(configurationSamlAuth),
		"configuration-openid-connect-auth":             string(configurationOpenidConnectAuth),
		"external-auth-load-modules-conf":               string(externalAuthLoadModulesConf),
		"external-auth-login-form-conf":                 string(externalAuthLoginFormConf),
		"external-auth-application-api-conf":            string(externalAuthApplicationApiConf),
		"external-auth-lookup-user-details-conf":        string(externalAuthLookupUserDetailsConf),
		"external-auth-remote-user-conf":                string(externalAuthRemoteUserConf),
		"external-auth-openid-connect-remote-user-conf": string(externalAuthOpenidConnectRemoteUserConf),
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httpd-configs",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Data: data,
	}
}

func NewHttpdAuthConfigMap(cr *miqv1alpha1.Manageiq) *corev1.ConfigMap {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}
	authconf := readContentFromFile("pkg/resources/httpd_conf/authentication.conf")
	data := map[string]string{
		"auth-type":                       "internal",
		"auth-kerberos-realms":            "undefined",
		"auth-oidc-provider-metadata-url": "undefined",
		"auth-oidc-client-id":             "undefined",
		"auth-oidc-client-secret":         "undefined",
		"auth-configuration.conf":         string(authconf),
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

func NewHttpdDeployment(cr *miqv1alpha1.Manageiq) *appsv1.Deployment {
	DeploymentLabels := map[string]string{
		"app": cr.Spec.AppName,
	}

	PodLabels := map[string]string{
		"name": "httpd",
		"app":  cr.Spec.AppName,
	}

	var RepNum int32 = 1

	memLimit, _ := resource.ParseQuantity(cr.Spec.HttpdMemLimit)
	memReq, _ := resource.ParseQuantity(cr.Spec.HttpdMemReq)
	cpuReq, _ := resource.ParseQuantity(cr.Spec.HttpdCPUReq)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httpd",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    DeploymentLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &RepNum,
			Selector: &metav1.LabelSelector{
				MatchLabels: PodLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "httpd",
					Labels: PodLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:  "httpd",
							Image: cr.Spec.HttpdImageName + ":" + cr.Spec.HttpdImageTag,
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									ContainerPort: 80,
									Protocol:      "TCP",
								},
								corev1.ContainerPort{
									ContainerPort: 8080,
									Protocol:      "TCP",
								},
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{"pidof", "httpd"},
									},
								},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      3,
							},
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name:  "APPLICATION_DOMAIN",
									Value: cr.Spec.ApplicationDomain,
								},
								corev1.EnvVar{
									Name: "HTTPD_AUTH_TYPE",

									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: "httpd-auth-configs"},
											Key:                  "auth-type",
										},
									},
								},

								corev1.EnvVar{
									Name:  "APPLICATION_DOMAIN",
									Value: cr.Spec.ApplicationDomain,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{Name: "httpd-config", MountPath: "/etc/httpd/conf.d"},
								corev1.VolumeMount{Name: "httpd-auth-config", MountPath: "/etc/httpd/auth-conf.d"},
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
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{"/usr/bin/save-container-environment"},
									},
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: "httpd-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "httpd-configs"},
								},
							},
						},
						corev1.Volume{
							Name: "httpd-auth-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "httpd-auth-configs"},
								},
							},
						},
					},
					ServiceAccountName: cr.Spec.AppName + "-httpd",
				},
			},
		},
	}
}

func NewUIService(cr *miqv1alpha1.Manageiq) *corev1.Service {
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

func NewWebService(cr *miqv1alpha1.Manageiq) *corev1.Service {
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

func NewRemoteConsoleService(cr *miqv1alpha1.Manageiq) *corev1.Service {
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

func NewHttpdService(cr *miqv1alpha1.Manageiq) *corev1.Service {
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
					Port: 80,
				},
			},
			Selector: selector,
		},
	}
}

func NewHttpdDbusAPIService(cr *miqv1alpha1.Manageiq) *corev1.Service {
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
					Port: 8080,
				},
			},
			Selector: selector,
		},
	}
}

func TLSSecret(cr *miqv1alpha1.Manageiq) *corev1.Secret {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}

	crt, key := tlstools.GenerateCrt("server")
	secret := map[string]string{
		"tls.crt": string(crt),
		"tls.key": string(key),
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tls-secret",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		StringData: secret,
		Type:       "kubernetes.io/tls",
	}
}
