package miqtools

import (
	"fmt"

	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
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
		if err := controllerutil.SetControllerReference(cr, sa, scheme); err != nil {
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

func AutomationRole(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*rbacv1.Role, controllerutil.MutateFn) {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "manageiq-automation",
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
				Resources: []string{"pods", "secrets"},
				Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
			},
		}

		return nil
	}

	return role, f
}

func AutomationRoleBinding(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*rbacv1.RoleBinding, controllerutil.MutateFn) {
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "manageiq-automation",
			Namespace: cr.ObjectMeta.Namespace,
		},
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, rb, scheme); err != nil {
			return err
		}

		rb.RoleRef = rbacv1.RoleRef{
			Kind:     "Role",
			Name:     "manageiq-automation",
			APIGroup: "rbac.authorization.k8s.io",
		}
		rb.Subjects = []rbacv1.Subject{
			rbacv1.Subject{
				Kind: "ServiceAccount",
				Name: "manageiq-automation",
			},
		}

		return nil
	}

	return rb, f
}

func AutomationServiceAccount(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*corev1.ServiceAccount, controllerutil.MutateFn) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "manageiq-automation",
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
