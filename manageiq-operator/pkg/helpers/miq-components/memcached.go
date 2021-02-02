package miqtools

import (
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func NewMemcachedDeployment(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*appsv1.Deployment, controllerutil.MutateFn, error) {
	memcachedImage := os.Getenv("MEMCACHED_IMAGE")
	if memcachedImage == "" {
		memcachedImage = cr.Spec.MemcachedImageName + ":" + cr.Spec.MemcachedImageTag
	}

	container := corev1.Container{
		Name:            "memcached",
		Image:           memcachedImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Ports: []corev1.ContainerPort{
			corev1.ContainerPort{
				ContainerPort: 11211,
				Protocol:      "TCP",
			},
		},
		LivenessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(11211),
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(11211),
				},
			},
		},
		Env: []corev1.EnvVar{
			corev1.EnvVar{
				Name:  "MEMCACHED_MAX_MEMORY",
				Value: cr.Spec.MemcachedMaxMemory,
			},
			corev1.EnvVar{
				Name:  "MEMCACHED_MAX_CONNECTIONS",
				Value: cr.Spec.MemcachedMaxConnection,
			},
			corev1.EnvVar{
				Name:  "MEMCACHED_SLAB_PAGE_SIZE",
				Value: cr.Spec.MemcachedSlabPageSize,
			},
		},
	}

	err := addResourceReqs(cr.Spec.MemcachedMemoryLimit, cr.Spec.MemcachedMemoryRequest, cr.Spec.MemcachedCpuLimit, cr.Spec.MemcachedCpuRequest, &container)
	if err != nil {
		return nil, nil, err
	}

	podLabels := map[string]string{
		"name": "memcached",
		"app":  cr.Spec.AppName,
	}

	// Values in this deployment are either immutable or used for lookup
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "memcached",
			Namespace: cr.ObjectMeta.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "memcached",
					Labels: podLabels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: defaultServiceAccountName(cr.Spec.AppName),
				},
			},
		},
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, deployment, scheme); err != nil {
			return err
		}
		addAppLabel(cr.Spec.AppName, &deployment.ObjectMeta)
		var repNum int32 = 1
		deployment.Spec.Replicas = &repNum
		deployment.Spec.Strategy = appsv1.DeploymentStrategy{
			Type: "Recreate",
		}
		deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
		deployment.Spec.Template.Spec.ServiceAccountName = defaultServiceAccountName(cr.Spec.AppName)
		return nil
	}

	return deployment, f, nil
}

func NewMemcachedService(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*corev1.Service, controllerutil.MutateFn) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "memcached",
			Namespace: cr.ObjectMeta.Namespace,
		},
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, service, scheme); err != nil {
			return err
		}
		addAppLabel(cr.Spec.AppName, &service.ObjectMeta)
		if len(service.Spec.Ports) == 0 {
			service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{})
		}
		service.Spec.Ports[0].Name = "memcached"
		service.Spec.Ports[0].Port = 11211
		service.Spec.Selector = map[string]string{"name": "memcached"}
		return nil
	}

	return service, f
}
