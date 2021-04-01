package miqtools

import (
	"fmt"
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	controllertools "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/helpers/controllertools"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func addSAPullSecret(sa *corev1.ServiceAccount, secret string) {
	secretRef := corev1.LocalObjectReference{Name: secret}
	if sa.ImagePullSecrets == nil {
		sa.ImagePullSecrets = []corev1.LocalObjectReference{secretRef}
	} else {
		for _, ref := range sa.ImagePullSecrets {
			if ref.Name == secret {
				return
			}
		}
		sa.ImagePullSecrets = append(sa.ImagePullSecrets, secretRef)
	}
}

func defaultServiceAccountName(appName string) string {
	return fmt.Sprintf("%s-default", appName)
}

func DefaultServiceAccount(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*corev1.ServiceAccount, controllerutil.MutateFn) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultServiceAccountName(cr.Spec.AppName),
			Namespace: cr.ObjectMeta.Namespace,
		},
	}

	f := func() error {
		if err := controllertools.ReplaceControllerReference(cr, sa, scheme); err != nil {
			return err
		}

		if cr.Spec.ImagePullSecret != "" {
			addSAPullSecret(sa, cr.Spec.ImagePullSecret)
		}

		addBackupLabel(cr.Spec.BackupLabelName, &sa.ObjectMeta)

		return nil
	}

	return sa, f
}
