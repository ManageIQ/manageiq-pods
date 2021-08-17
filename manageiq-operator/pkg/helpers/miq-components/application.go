package miqtools

import (
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func ApplicationUiHttpdConfigMap(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme, client client.Client) (*corev1.ConfigMap, controllerutil.MutateFn) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ui-httpd-configs",
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

		protocol := "http"

		configMap.Data["manageiq-http.conf"] = uiHttpdConfig(protocol)

		return nil
	}

	return configMap, f
}

func ApplicationApiHttpdConfigMap(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme, client client.Client) (*corev1.ConfigMap, controllerutil.MutateFn) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-httpd-configs",
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

		protocol := "http"

		configMap.Data["manageiq-http.conf"] = apiHttpdConfig(protocol)

		return nil
	}

	return configMap, f
}
