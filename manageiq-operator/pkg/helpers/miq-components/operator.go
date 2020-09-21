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
	operatorPodName := os.Getenv("POD_NAME")
	podKey := types.NamespacedName{Namespace: cr.Namespace, Name: operatorPodName}
	pod := &corev1.Pod{}
	client.Get(context.TODO(), podKey, pod)

	operatorReplicaSetName := pod.ObjectMeta.OwnerReferences[0].Name
	replicaSetKey := types.NamespacedName{Namespace: cr.Namespace, Name: operatorReplicaSetName}
	replicaSet := &appsv1.ReplicaSet{}
	client.Get(context.TODO(), replicaSetKey, replicaSet)

	operatorDeploymentName := replicaSet.ObjectMeta.OwnerReferences[0].Name
	deploymentKey := types.NamespacedName{Namespace: cr.Namespace, Name: operatorDeploymentName}
	deployment := &appsv1.Deployment{}
	client.Get(context.TODO(), deploymentKey, deployment)

	f := func() error {
		addBackupLabel(cr.Spec.BackupLabelName, &deployment.ObjectMeta)

		return nil
	}

	return deployment, f
}
