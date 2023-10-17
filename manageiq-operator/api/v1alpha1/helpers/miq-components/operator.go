package miqtools

import (
	"context"

	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1"
	miqutilsv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1/miqutils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func ManageOperator(cr *miqv1alpha1.ManageIQ, client client.Client) (*appsv1.Deployment, controllerutil.MutateFn) {
	podName := os.Getenv("POD_NAME")
	pod := miqutilsv1alpha1.FindPodByName(client, cr.Namespace, podName)
	replicaSet := miqutilsv1alpha1.FindReplicaSetByName(client, cr.Namespace, pod.ObjectMeta.OwnerReferences[0].Name)
	deployment := miqutilsv1alpha1.FindDeploymentByName(client, cr.Namespace, replicaSet.ObjectMeta.OwnerReferences[0].Name)

	f := func() error {
		addAppLabel(cr.Spec.AppName, &deployment.ObjectMeta)
		addAppLabel(cr.Spec.AppName, &deployment.Spec.Template.ObjectMeta)
		addBackupLabel(cr.Spec.BackupLabelName, &deployment.ObjectMeta)
		deployment.Spec.Template.Spec.Containers[0].SecurityContext = DefaultSecurityContext()

		return nil
	}

	return deployment, f
}

func ImagePullSecret(cr *miqv1alpha1.ManageIQ, client client.Client) (*corev1.Secret, controllerutil.MutateFn) {
	secretKey := types.NamespacedName{Namespace: cr.Namespace, Name: cr.Spec.ImagePullSecret}
	secret := &corev1.Secret{}
	client.Get(context.TODO(), secretKey, secret)

	f := func() error {
		addBackupLabel(cr.Spec.BackupLabelName, &secret.ObjectMeta)

		return nil
	}

	return secret, f
}

func OidcClientSecret(cr *miqv1alpha1.ManageIQ, client client.Client) (*corev1.Secret, controllerutil.MutateFn) {
	secretKey := types.NamespacedName{Namespace: cr.Namespace, Name: cr.Spec.OIDCClientSecret}
	secret := &corev1.Secret{}
	client.Get(context.TODO(), secretKey, secret)

	f := func() error {
		addBackupLabel(cr.Spec.BackupLabelName, &secret.ObjectMeta)

		return nil
	}

	return secret, f
}

func OidcCaCertSecret(cr *miqv1alpha1.ManageIQ, client client.Client) (*corev1.Secret, controllerutil.MutateFn) {
	secretKey := types.NamespacedName{Namespace: cr.Namespace, Name: cr.Spec.OIDCCACertSecret}
	secret := &corev1.Secret{}
	client.Get(context.TODO(), secretKey, secret)

	f := func() error {
		addBackupLabel(cr.Spec.BackupLabelName, &secret.ObjectMeta)

		return nil
	}

	return secret, f
}

func ManageOperatorServiceAccount(cr *miqv1alpha1.ManageIQ, client client.Client) (*corev1.ServiceAccount, controllerutil.MutateFn) {
	serviceAccount := operatorServiceAccount(cr, client)

	f := func() error {
		addBackupLabel(cr.Spec.BackupLabelName, &serviceAccount.ObjectMeta)

		return nil
	}

	return serviceAccount, f
}

func ManageOperatorRole(cr *miqv1alpha1.ManageIQ, client client.Client) (*rbacv1.Role, controllerutil.MutateFn) {
	operatorRole := operatorRole(cr, client)

	f := func() error {
		addBackupLabel(cr.Spec.BackupLabelName, &operatorRole.ObjectMeta)

		return nil
	}

	return operatorRole, f
}

func ManageOperatorRoleBinding(cr *miqv1alpha1.ManageIQ, client client.Client) (*rbacv1.RoleBinding, controllerutil.MutateFn) {
	operatorRoleBinding := operatorRoleBinding(cr, client)

	f := func() error {
		addBackupLabel(cr.Spec.BackupLabelName, &operatorRoleBinding.ObjectMeta)

		return nil
	}

	return operatorRoleBinding, f
}

func operatorServiceAccount(cr *miqv1alpha1.ManageIQ, client client.Client) *corev1.ServiceAccount {
	podName := os.Getenv("POD_NAME")
	pod := miqutilsv1alpha1.FindPodByName(client, cr.Namespace, podName)
	operatorServiceAccountName := pod.Spec.ServiceAccountName
	serviceAccountKey := types.NamespacedName{Namespace: cr.Namespace, Name: operatorServiceAccountName}
	serviceAccount := &corev1.ServiceAccount{}
	client.Get(context.TODO(), serviceAccountKey, serviceAccount)

	return serviceAccount
}

func operatorRoleBinding(cr *miqv1alpha1.ManageIQ, c client.Client) *rbacv1.RoleBinding {
	roleBindingList := &rbacv1.RoleBindingList{}
	c.List(context.TODO(), roleBindingList)

	operatorServiceAccount := operatorServiceAccount(cr, c)

	for _, roleBinding := range roleBindingList.Items {
		for _, subject := range roleBinding.Subjects {
			if subject.Name == operatorServiceAccount.Name {
				return &roleBinding
			}
		}
	}

	return nil
}

func operatorRole(cr *miqv1alpha1.ManageIQ, c client.Client) *rbacv1.Role {
	operatorRoleBinding := operatorRoleBinding(cr, c)

	roleKey := types.NamespacedName{Namespace: cr.Namespace, Name: operatorRoleBinding.RoleRef.Name}
	role := &rbacv1.Role{}
	c.Get(context.TODO(), roleKey, role)

	return role
}
