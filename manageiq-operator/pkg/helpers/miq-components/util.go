package miqtools

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

func serviceMutateFn(service *corev1.Service, labels map[string]string, ports []corev1.ServicePort, selector map[string]string) controllerutil.MutateFn {

	mutateFn := func() error {
		service.ObjectMeta.Labels = labels
		service.Spec.Ports = ports
		service.Spec.Selector = selector
		return nil
	}

	return mutateFn
}

func configMapMutateFn(configMap *corev1.ConfigMap, labels map[string]string, data map[string]string) controllerutil.MutateFn {

	mutateFn := func() error {
		configMap.ObjectMeta.Labels = labels
		if len(data) > 0 {
			configMap.Data = data
		}
		return nil
	}

	return mutateFn
}

func pvcMutateFn(pvc *corev1.PersistentVolumeClaim, labels map[string]string, accessModes []corev1.PersistentVolumeAccessMode, resources corev1.ResourceRequirements) controllerutil.MutateFn {

	mutateFn := func() error {
		pvc.ObjectMeta.Labels = labels
		pvc.Spec.AccessModes = accessModes
		pvc.Spec.Resources = resources
		return nil
	}

	return mutateFn
}
