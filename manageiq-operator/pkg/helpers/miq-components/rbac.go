package miqtools

import (
	"fmt"
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DefaultServiceAccount(cr *miqv1alpha1.ManageIQ) *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultServiceAccountName(cr.Spec.AppName),
			Namespace: cr.ObjectMeta.Namespace,
		},
	}

	if cr.Spec.ImagePullSecret != "" {
		addSAPullSecret(sa, cr.Spec.ImagePullSecret)
	}

	return sa
}

func HttpdServiceAccount(cr *miqv1alpha1.ManageIQ) *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.AppName + "-httpd",
			Namespace: cr.ObjectMeta.Namespace,
		},
	}

	if cr.Spec.ImagePullSecret != "" {
		addSAPullSecret(sa, cr.Spec.ImagePullSecret)
	}

	return sa
}

func OrchestratorServiceAccount(cr *miqv1alpha1.ManageIQ) *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      orchestratorObjectName(cr),
			Namespace: cr.ObjectMeta.Namespace,
		},
	}

	if cr.Spec.ImagePullSecret != "" {
		addSAPullSecret(sa, cr.Spec.ImagePullSecret)
	}

	return sa
}

func OrchestratorRole(cr *miqv1alpha1.ManageIQ) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      orchestratorObjectName(cr),
			Namespace: cr.ObjectMeta.Namespace,
		},
		Rules: []rbacv1.PolicyRule{
			rbacv1.PolicyRule{
				APIGroups: []string{""},
				Resources: []string{"pods", "pods/finalizers"},
				Verbs:     []string{"*"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments", "deployments/scale"},
				Verbs:     []string{"*"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"extensions"},
				Resources: []string{"deployments", "deployments/scale"},
				Verbs:     []string{"*"},
			},
		},
	}
}

func OrchestratorRoleBinding(cr *miqv1alpha1.ManageIQ) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      orchestratorObjectName(cr),
			Namespace: cr.ObjectMeta.Namespace,
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     orchestratorObjectName(cr),
			APIGroup: "rbac.authorization.k8s.io",
		},
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind: "ServiceAccount",
				Name: orchestratorObjectName(cr),
			},
		},
	}
}

func orchestratorObjectName(cr *miqv1alpha1.ManageIQ) string {
	return cr.Spec.AppName + "-orchestrator"
}

func defaultServiceAccountName(appName string) string {
	return fmt.Sprintf("%s-default", appName)
}

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
