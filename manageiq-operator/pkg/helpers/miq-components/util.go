package miqtools

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func addResourceReqs(memLimit string, memReq string, cpuReq string, c *corev1.Container) error {
	if memLimit == "" && memReq == "" && cpuReq == "" {
		return nil
	}

	if memLimit != "" {
		limit, err := resource.ParseQuantity(memLimit)
		if err != nil {
			return err
		}
		c.Resources.Limits = corev1.ResourceList{"memory": limit}
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
