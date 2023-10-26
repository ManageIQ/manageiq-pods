package miqutils

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func FindPodByName(client client.Client, namespace string, name string) *corev1.Pod {
	podKey := types.NamespacedName{Namespace: namespace, Name: name}
	pod := &corev1.Pod{}
	client.Get(context.TODO(), podKey, pod)

	return pod
}

func FindReplicaSetByName(client client.Client, namespace string, name string) *appsv1.ReplicaSet {
	replicaSetKey := types.NamespacedName{Namespace: namespace, Name: name}
	replicaSet := &appsv1.ReplicaSet{}
	client.Get(context.TODO(), replicaSetKey, replicaSet)

	return replicaSet
}

func FindDeploymentByName(client client.Client, namespace string, name string) *appsv1.Deployment {
	deploymentKey := types.NamespacedName{Namespace: namespace, Name: name}
	deployment := &appsv1.Deployment{}
	client.Get(context.TODO(), deploymentKey, deployment)

	return deployment
}
