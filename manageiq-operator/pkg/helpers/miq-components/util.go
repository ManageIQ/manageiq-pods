package miqtools

import (
	"context"
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
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
