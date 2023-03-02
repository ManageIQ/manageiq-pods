package miqtools

import (
	"context"
	"fmt"

	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func addResourceReqs(memLimit, memReq, cpuLimit, cpuReq string, c *corev1.Container) error {
	if memLimit == "" && memReq == "" && cpuLimit == "" && cpuReq == "" {
		return nil
	}

	if memLimit != "" || cpuLimit != "" {
		c.Resources.Limits = make(map[corev1.ResourceName]resource.Quantity)
	}

	if memLimit != "" {
		limit, err := resource.ParseQuantity(memLimit)
		if err != nil {
			return err
		}
		c.Resources.Limits["memory"] = limit
	}

	if cpuLimit != "" {
		limit, err := resource.ParseQuantity(cpuLimit)
		if err != nil {
			return err
		}
		c.Resources.Limits["cpu"] = limit
	}

	if memReq != "" || cpuReq != "" {
		c.Resources.Requests = make(map[corev1.ResourceName]resource.Quantity)
	}

	if memReq != "" {
		req, err := resource.ParseQuantity(memReq)
		if err != nil {
			return err
		}
		c.Resources.Requests["memory"] = req
	}

	if cpuReq != "" {
		req, err := resource.ParseQuantity(cpuReq)
		if err != nil {
			return err
		}
		c.Resources.Requests["cpu"] = req
	}

	return nil
}

func addAppLabel(appName string, meta *metav1.ObjectMeta) {
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	meta.Labels["app"] = appName
}

func addBackupLabel(backupLabel string, meta *metav1.ObjectMeta) {
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	meta.Labels[backupLabel] = "t"
}

func addBackupAnnotation(volumesToBackup string, meta *metav1.ObjectMeta) {
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	meta.Annotations["backup.velero.io/backup-volumes"] = volumesToBackup
}

func addAnnotations(annotations map[string]string, meta *metav1.ObjectMeta) {
	if len(annotations) > 0 {
		if meta.Annotations == nil {
			meta.Annotations = make(map[string]string)
		}
		for key, value := range annotations {
			meta.Annotations[key] = value
		}
	}
}

func InternalCertificatesSecret(cr *miqv1alpha1.ManageIQ, client client.Client) *corev1.Secret {
	name := cr.Spec.InternalCertificatesSecret

	secretKey := types.NamespacedName{Namespace: cr.Namespace, Name: name}
	secret := &corev1.Secret{}
	client.Get(context.TODO(), secretKey, secret)

	return secret
}

func addInternalCertificate(cr *miqv1alpha1.ManageIQ, d *appsv1.Deployment, client client.Client, name string, mountPoint string) {
	secret := InternalCertificatesSecret(cr, client)
	if secret.Data[fmt.Sprintf("%s_crt", name)] != nil && secret.Data[fmt.Sprintf("%s_key", name)] != nil {
		volumeName := fmt.Sprintf("%s-certificate", name)

		volumeMount := corev1.VolumeMount{Name: volumeName, MountPath: mountPoint, ReadOnly: true}
		d.Spec.Template.Spec.Containers[0].VolumeMounts = addOrUpdateVolumeMount(d.Spec.Template.Spec.Containers[0].VolumeMounts, volumeMount)

		secretVolumeSource := corev1.SecretVolumeSource{SecretName: secret.Name, Items: []corev1.KeyToPath{corev1.KeyToPath{Key: fmt.Sprintf("%s_crt", name), Path: "server.crt"}, corev1.KeyToPath{Key: fmt.Sprintf("%s_key", name), Path: "server.key"}}}
		d.Spec.Template.Spec.Volumes = addOrUpdateVolume(d.Spec.Template.Spec.Volumes, corev1.Volume{Name: volumeName, VolumeSource: corev1.VolumeSource{Secret: &secretVolumeSource}})
	}
}

func addKafkaStores(cr *miqv1alpha1.ManageIQ, d *appsv1.Deployment, client client.Client, mountPoint string) {
	secret := InternalCertificatesSecret(cr, client)
	if secret.Data["kafka_truststore"] != nil && secret.Data["kafka_keystore"] != nil && secret.Data["zookeeper_keystore"] != nil {
		volumeName := "kafka-certificate"

		volumeMount := corev1.VolumeMount{Name: volumeName, MountPath: mountPoint, ReadOnly: true}
		d.Spec.Template.Spec.Containers[0].VolumeMounts = addOrUpdateVolumeMount(d.Spec.Template.Spec.Containers[0].VolumeMounts, volumeMount)

		secretVolumeSource := corev1.SecretVolumeSource{SecretName: secret.Name, Items: []corev1.KeyToPath{corev1.KeyToPath{Key: "kafka_truststore", Path: "kafka.truststore.jks"}, corev1.KeyToPath{Key: "kafka_keystore", Path: "kafka.keystore.jks"}, corev1.KeyToPath{Key: "zookeeper_keystore", Path: "zookeeper.keystore.jks"}}}
		d.Spec.Template.Spec.Volumes = addOrUpdateVolume(d.Spec.Template.Spec.Volumes, corev1.Volume{Name: volumeName, VolumeSource: corev1.VolumeSource{Secret: &secretVolumeSource}})
	}
}

func addZookeeperStores(cr *miqv1alpha1.ManageIQ, d *appsv1.Deployment, client client.Client, mountPoint string) {
	secret := InternalCertificatesSecret(cr, client)
	if secret.Data["kafka_truststore"] != nil && secret.Data["zookeeper_keystore"] != nil {
		volumeName := "zookeeper-certificate"

		volumeMount := corev1.VolumeMount{Name: volumeName, MountPath: mountPoint, ReadOnly: true}
		d.Spec.Template.Spec.Containers[0].VolumeMounts = addOrUpdateVolumeMount(d.Spec.Template.Spec.Containers[0].VolumeMounts, volumeMount)

		secretVolumeSource := corev1.SecretVolumeSource{SecretName: secret.Name, Items: []corev1.KeyToPath{corev1.KeyToPath{Key: "kafka_truststore", Path: "kafka.truststore.jks"}, corev1.KeyToPath{Key: "zookeeper_keystore", Path: "zookeeper.keystore.jks"}}}
		d.Spec.Template.Spec.Volumes = addOrUpdateVolume(d.Spec.Template.Spec.Volumes, corev1.Volume{Name: volumeName, VolumeSource: corev1.VolumeSource{Secret: &secretVolumeSource}})
	}
}

func addOrUpdateEnvVar(environment []corev1.EnvVar, variable corev1.EnvVar) []corev1.EnvVar {
	index := -1
	for i, env := range environment {
		if env.Name == variable.Name {
			index = i
		}
	}

	if index == -1 {
		environment = append(environment, variable)
	} else {
		environment[index] = variable
	}

	return environment
}

func addOrUpdateVolumeMount(volumeMounts []corev1.VolumeMount, volumeMount corev1.VolumeMount) []corev1.VolumeMount {
	if volumeMounts == nil {
		volumeMounts = []corev1.VolumeMount{}
	}

	index := -1
	for i, v := range volumeMounts {
		if v.Name == volumeMount.Name {
			index = i
		}
	}

	if index == -1 {
		volumeMounts = append(volumeMounts, volumeMount)
	} else {
		volumeMounts[index] = volumeMount
	}

	return volumeMounts
}

func addOrUpdateVolume(volumes []corev1.Volume, volume corev1.Volume) []corev1.Volume {
	if volumes == nil {
		volumes = []corev1.Volume{}
	}

	index := -1
	for i, v := range volumes {
		if v.Name == volume.Name {
			index = i
		}
	}

	if index == -1 {
		volumes = append(volumes, volume)
	} else {
		volumes[index] = volume
	}

	return volumes
}

func DefaultSecurityContext() *corev1.SecurityContext {
	dropCapability := []corev1.Capability{"ALL"}
	varFalse := false
	varTrue := true
	sc := &corev1.SecurityContext{
		AllowPrivilegeEscalation: &varFalse,
		Privileged:               &varFalse,
		Capabilities: &corev1.Capabilities{
			Drop: dropCapability,
		},
		RunAsNonRoot: &varTrue,
	}

	return sc
}
