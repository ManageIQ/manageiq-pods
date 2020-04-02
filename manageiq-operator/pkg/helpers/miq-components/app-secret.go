package miqtools

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"

	miqv1alpha1 "github.com/manageiq/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AppSecret(cr *miqv1alpha1.Manageiq) *corev1.Secret {

	labels := map[string]string{
		"app": cr.Spec.AppName,
	}
	secret := map[string]string{
		"admin-password": cr.Spec.ApplicationAdminPassword,
		"encryption-key": generateEncryptionKey(),
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-secrets",
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		StringData: secret,
	}
}

func generateEncryptionKey() string {
	buf := make([]byte, 32)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err) // out of randomness, should never happen
	}

	h := sha256.New()
	h.Write(buf)

	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}
