package miqtools

import (
	miqv1alpha1 "github.com/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func HttpdServiceAccount(cr *miqv1alpha1.Manageiq) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.AppName + "-httpd",
			Namespace: cr.ObjectMeta.Namespace,
		},
	}
}

func OrchestratorServiceAccount(cr *miqv1alpha1.Manageiq) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.AppName + "-orchestrator",
			Namespace: cr.ObjectMeta.Namespace,
		},
	}
}

func AnyuidServiceAccount(cr *miqv1alpha1.Manageiq) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.AppName + "-anyuid",
			Namespace: cr.ObjectMeta.Namespace,
		},
	}
}

func OrchestratorViewRoleBinding(cr *miqv1alpha1.Manageiq) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "view",
			Namespace: cr.ObjectMeta.Namespace,
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "view",
			APIGroup: "rbac.authorization.k8s.io",
		},
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind: "ServiceAccount",
				Name: cr.Spec.AppName + "-orchestrator",
			},
		},
	}
}

func OrchestratorEditRoleBinding(cr *miqv1alpha1.Manageiq) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "edit",
			Namespace: cr.ObjectMeta.Namespace,
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "edit",
			APIGroup: "rbac.authorization.k8s.io",
		},
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind: "ServiceAccount",
				Name: cr.Spec.AppName + "-orchestrator",
			},
		},
	}
}
