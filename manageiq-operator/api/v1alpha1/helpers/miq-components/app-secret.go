package miqtools

import (
	"context"

	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1"
	miqutils "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1/helpers/miq-components/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const appSecretName = "app-secrets"

func ManageAppSecret(cr *miqv1alpha1.ManageIQ, client client.Client, scheme *runtime.Scheme) (*corev1.Secret, controllerutil.MutateFn) {
	secretKey := types.NamespacedName{Namespace: cr.ObjectMeta.Namespace, Name: appSecretName}
	secret := &corev1.Secret{}
	secretErr := client.Get(context.TODO(), secretKey, secret)
	if secretErr != nil {
		secret = defaultAppSecret(cr)
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, secret, scheme); err != nil {
			return err
		}

		miqutils.AddAppLabel(cr.Spec.AppName, &secret.ObjectMeta)
		miqutils.AddBackupLabel(cr.Spec.BackupLabelName, &secret.ObjectMeta)

		return nil
	}

	return secret, f
}

func defaultAppSecret(cr *miqv1alpha1.ManageIQ) *corev1.Secret {
	secretData := map[string]string{
		"encryption-key": generateEncryptionKey(),
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appSecretName,
			Namespace: cr.ObjectMeta.Namespace,
		},
		StringData: secretData,
	}

	return secret
}
