package miqtools

import (
	"context"
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

func GetSecretKeyValue(client client.Client, nameSpace string, secretName string, keyName string) string {
	secretKey := types.NamespacedName{Namespace: nameSpace, Name: secretName}
	secret := &corev1.Secret{}
	secretErr := client.Get(context.TODO(), secretKey, secret)
	if secretErr != nil {
		return ""
	}
	return string(secret.Data[keyName])
}
