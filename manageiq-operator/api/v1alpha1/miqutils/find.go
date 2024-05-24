package miqutils

import (
	"context"
	"strings"

	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

func FindSecretByName(client client.Client, namespace string, name string) *corev1.Secret {
	secretKey := types.NamespacedName{Namespace: namespace, Name: name}
	secret := &corev1.Secret{}
	client.Get(context.TODO(), secretKey, secret)

	return secret
}

func FindKafka(client client.Client, scheme *runtime.Scheme, namespace string, name string) *unstructured.Unstructured {
	kafkaKey := types.NamespacedName{Namespace: namespace, Name: name}
	kafka := &unstructured.Unstructured{}
	kafka.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "kafka.strimzi.io",
		Kind:    "Kafka",
		Version: "v1beta2",
	})
	client.Get(context.TODO(), kafkaKey, kafka)

	return kafka
}

func FindKafkaTopic(client client.Client, scheme *runtime.Scheme, namespace string, name string, group string) *unstructured.Unstructured {
	kafkaTopicKey := types.NamespacedName{Namespace: namespace, Name: name}
	kafkaTopic := &unstructured.Unstructured{}
	kafkaTopic.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   group,
		Kind:    "KafkaTopic",
		Version: "v1beta2",
	})
	client.Get(context.TODO(), kafkaTopicKey, kafkaTopic)

	return kafkaTopic
}

func FindCatalogSourceByName(client client.Client, namespace string, name string) *olmv1alpha1.CatalogSource {
	catalogSourceKey := types.NamespacedName{Namespace: namespace, Name: name}
	catalogSource := &olmv1alpha1.CatalogSource{}
	if err := client.Get(context.TODO(), catalogSourceKey, catalogSource); err != nil {
		if strings.Contains(err.Error(), "no matches for kind") || errors.IsNotFound(err) {
			return nil
		}
	}

	return catalogSource
}

func FindVolume(volumeName string, volumes []corev1.Volume) corev1.Volume {
	volume := corev1.Volume{}
	for i := 0; i < len(volumes); i++ {
		if volumes[i].Name == volumeName {
			volume = volumes[i]
		}
	}

	return volume
}

func FindVolumeProjection(volumeProjectionName string, volumeProjections []corev1.VolumeProjection) corev1.VolumeProjection {
	volumeProjection := corev1.VolumeProjection{}
	for i := 0; i < len(volumeProjections); i++ {
		if volumeProjections[i].Secret != nil && volumeProjections[i].Secret.LocalObjectReference.Name == volumeProjectionName {
			volumeProjection = volumeProjections[i]
		} else if volumeProjections[i].ConfigMap != nil && volumeProjections[i].ConfigMap.LocalObjectReference.Name == volumeProjectionName {
			volumeProjection = volumeProjections[i]
		}
	}

	return volumeProjection
}
