package miqtools

import (
	"context"
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func ManageOperator(cr *miqv1alpha1.ManageIQ, client client.Client) (*appsv1.Deployment, controllerutil.MutateFn) {
	deployment := operatorDeployment(cr, client)

	f := func() error {
		addBackupLabel(cr.Spec.BackupLabelName, &deployment.ObjectMeta)

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

func operatorPod(cr *miqv1alpha1.ManageIQ, client client.Client) *corev1.Pod {
	operatorPodName := os.Getenv("POD_NAME")
	podKey := types.NamespacedName{Namespace: cr.Namespace, Name: operatorPodName}
	pod := &corev1.Pod{}
	client.Get(context.TODO(), podKey, pod)

	return pod
}

func operatorReplicaSet(cr *miqv1alpha1.ManageIQ, client client.Client) *appsv1.ReplicaSet {
	pod := operatorPod(cr, client)
	operatorReplicaSetName := pod.ObjectMeta.OwnerReferences[0].Name
	replicaSetKey := types.NamespacedName{Namespace: cr.Namespace, Name: operatorReplicaSetName}
	replicaSet := &appsv1.ReplicaSet{}
	client.Get(context.TODO(), replicaSetKey, replicaSet)

	return replicaSet
}

func operatorDeployment(cr *miqv1alpha1.ManageIQ, client client.Client) *appsv1.Deployment {
	replicaSet := operatorReplicaSet(cr, client)
	operatorDeploymentName := replicaSet.ObjectMeta.OwnerReferences[0].Name
	deploymentKey := types.NamespacedName{Namespace: cr.Namespace, Name: operatorDeploymentName}
	deployment := &appsv1.Deployment{}
	client.Get(context.TODO(), deploymentKey, deployment)

	return deployment
}
