package miqtools

import (
	miqv1alpha1 "github.com/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewPostgresqlSecret(cr *miqv1alpha1.Manageiq) *corev1.Secret {
	labels := map[string]string{
		"app": cr.Spec.AppName,
	}
	secret := map[string]string{
		"dbname":   cr.Spec.DatabaseName,
		"username": cr.Spec.DatabaseUser,
		"password": cr.Spec.DatabasePassword,
		"hostname": "postgresql",
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql-secrets",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		StringData: secret,
	}
}
