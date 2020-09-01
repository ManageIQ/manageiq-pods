package miqtools

import (
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AppSecret(cr *miqv1alpha1.ManageIQ) *corev1.Secret {
	secretData := map[string]string{
		"encryption-key": generateEncryptionKey(),
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-secrets",
			Namespace: cr.ObjectMeta.Namespace,
		},
		StringData: secretData,
	}

	addAppLabel(cr.Spec.AppName, &secret.ObjectMeta)
	addBackupLabel(cr.Spec.BackupLabelName, &secret.ObjectMeta)

	return secret
}
