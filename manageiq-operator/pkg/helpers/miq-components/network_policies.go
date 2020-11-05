package miqtools

import (
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func NetworkPolicyDefaultDeny(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*extensionsv1beta1.NetworkPolicy, controllerutil.MutateFn) {
	networkPolicy := newNetworkPolicy(cr, "-default-deny")

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, networkPolicy, scheme); err != nil {
			return err
		}
		addAppLabel(cr.Spec.AppName, &networkPolicy.ObjectMeta)
		setIngressPolicyType(networkPolicy)

		return nil
	}

	return networkPolicy, f
}

func NetworkPolicyAllowInboundHttpd(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*extensionsv1beta1.NetworkPolicy, controllerutil.MutateFn) {
	networkPolicy := newNetworkPolicy(cr, "-allow-inbound-httpd")

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, networkPolicy, scheme); err != nil {
			return err
		}
		addAppLabel(cr.Spec.AppName, &networkPolicy.ObjectMeta)
		setIngressPolicyType(networkPolicy)

		networkPolicy.Spec.PodSelector.MatchLabels = map[string]string{"name": "httpd"}

		ensureIngressRule(networkPolicy)
		if len(networkPolicy.Spec.Ingress[0].From) != 1 {
			networkPolicy.Spec.Ingress[0].From = []extensionsv1beta1.NetworkPolicyPeer{
				extensionsv1beta1.NetworkPolicyPeer{},
			}
		}
		if len(networkPolicy.Spec.Ingress[0].Ports) != 1 {
			networkPolicy.Spec.Ingress[0].Ports = []extensionsv1beta1.NetworkPolicyPort{
				extensionsv1beta1.NetworkPolicyPort{},
			}
		}
		tcp := corev1.ProtocolTCP
		networkPolicy.Spec.Ingress[0].From[0].IPBlock = &extensionsv1beta1.IPBlock{}
		networkPolicy.Spec.Ingress[0].From[0].IPBlock.CIDR = "0.0.0.0/0"
		networkPolicy.Spec.Ingress[0].Ports[0].Protocol = &tcp
		networkPolicy.Spec.Ingress[0].Ports[0].Port = &intstr.IntOrString{Type: intstr.Int, IntVal: 8080}

		return nil
	}

	return networkPolicy, f
}

func NetworkPolicyAllowHttpdApi(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*extensionsv1beta1.NetworkPolicy, controllerutil.MutateFn) {
	networkPolicy := newNetworkPolicy(cr, "-allow-httpd-api")

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, networkPolicy, scheme); err != nil {
			return err
		}
		addAppLabel(cr.Spec.AppName, &networkPolicy.ObjectMeta)
		setIngressPolicyType(networkPolicy)

		networkPolicy.Spec.PodSelector.MatchLabels = map[string]string{"service": "web-service"}

		ensureIngressRule(networkPolicy)
		if len(networkPolicy.Spec.Ingress[0].From) != 1 {
			networkPolicy.Spec.Ingress[0].From = []extensionsv1beta1.NetworkPolicyPeer{
				extensionsv1beta1.NetworkPolicyPeer{},
			}
		}
		if len(networkPolicy.Spec.Ingress[0].Ports) != 1 {
			networkPolicy.Spec.Ingress[0].Ports = []extensionsv1beta1.NetworkPolicyPort{
				extensionsv1beta1.NetworkPolicyPort{},
			}
		}
		tcp := corev1.ProtocolTCP
		networkPolicy.Spec.Ingress[0].From[0].PodSelector = &metav1.LabelSelector{}
		networkPolicy.Spec.Ingress[0].From[0].PodSelector.MatchLabels = map[string]string{"name": "httpd"}
		networkPolicy.Spec.Ingress[0].Ports[0].Protocol = &tcp
		networkPolicy.Spec.Ingress[0].Ports[0].Port = &intstr.IntOrString{Type: intstr.Int, IntVal: 3000}

		return nil
	}

	return networkPolicy, f
}

func NetworkPolicyAllowHttpdUi(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*extensionsv1beta1.NetworkPolicy, controllerutil.MutateFn) {
	networkPolicy := newNetworkPolicy(cr, "-allow-httpd-ui")

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, networkPolicy, scheme); err != nil {
			return err
		}
		addAppLabel(cr.Spec.AppName, &networkPolicy.ObjectMeta)
		setIngressPolicyType(networkPolicy)

		networkPolicy.Spec.PodSelector.MatchLabels = map[string]string{"service": "ui"}

		ensureIngressRule(networkPolicy)
		if len(networkPolicy.Spec.Ingress[0].From) != 1 {
			networkPolicy.Spec.Ingress[0].From = []extensionsv1beta1.NetworkPolicyPeer{
				extensionsv1beta1.NetworkPolicyPeer{},
			}
		}
		if len(networkPolicy.Spec.Ingress[0].Ports) != 1 {
			networkPolicy.Spec.Ingress[0].Ports = []extensionsv1beta1.NetworkPolicyPort{
				extensionsv1beta1.NetworkPolicyPort{},
			}
		}
		tcp := corev1.ProtocolTCP
		networkPolicy.Spec.Ingress[0].From[0].PodSelector = &metav1.LabelSelector{}
		networkPolicy.Spec.Ingress[0].From[0].PodSelector.MatchLabels = map[string]string{"name": "httpd"}
		networkPolicy.Spec.Ingress[0].Ports[0].Protocol = &tcp
		networkPolicy.Spec.Ingress[0].Ports[0].Port = &intstr.IntOrString{Type: intstr.Int, IntVal: 3000}

		return nil
	}

	return networkPolicy, f
}

func NetworkPolicyAllowMemcached(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme, c *client.Client) (*extensionsv1beta1.NetworkPolicy, controllerutil.MutateFn) {
	networkPolicy := newNetworkPolicy(cr, "-allow-memcached")

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, networkPolicy, scheme); err != nil {
			return err
		}
		addAppLabel(cr.Spec.AppName, &networkPolicy.ObjectMeta)
		setIngressPolicyType(networkPolicy)

		networkPolicy.Spec.PodSelector.MatchLabels = map[string]string{"name": "memcached"}

		pod := orchestratorPod(*c)
		if pod == nil {
			return nil
		}

		ensureIngressRule(networkPolicy)
		if len(networkPolicy.Spec.Ingress[0].From) != 2 {
			networkPolicy.Spec.Ingress[0].From = []extensionsv1beta1.NetworkPolicyPeer{
				extensionsv1beta1.NetworkPolicyPeer{},
				extensionsv1beta1.NetworkPolicyPeer{},
			}
		}
		if len(networkPolicy.Spec.Ingress[0].Ports) != 1 {
			networkPolicy.Spec.Ingress[0].Ports = []extensionsv1beta1.NetworkPolicyPort{
				extensionsv1beta1.NetworkPolicyPort{},
			}
		}
		orchestratedByLabelKey := cr.Spec.AppName + "-orchestrated-by"
		orchestratedByLabelValue := pod.Name
		tcp := corev1.ProtocolTCP
		networkPolicy.Spec.Ingress[0].From[0].PodSelector = &metav1.LabelSelector{}
		networkPolicy.Spec.Ingress[0].From[0].PodSelector.MatchLabels = map[string]string{"name": "orchestrator"}
		networkPolicy.Spec.Ingress[0].From[1].PodSelector = &metav1.LabelSelector{}
		networkPolicy.Spec.Ingress[0].From[1].PodSelector.MatchLabels = map[string]string{orchestratedByLabelKey: orchestratedByLabelValue}
		networkPolicy.Spec.Ingress[0].Ports[0].Protocol = &tcp
		networkPolicy.Spec.Ingress[0].Ports[0].Port = &intstr.IntOrString{Type: intstr.Int, IntVal: 11211}

		return nil
	}

	return networkPolicy, f
}

func NetworkPolicyAllowPostgres(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme, c *client.Client) (*extensionsv1beta1.NetworkPolicy, controllerutil.MutateFn) {
	networkPolicy := newNetworkPolicy(cr, "-allow-postgres")

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, networkPolicy, scheme); err != nil {
			return err
		}
		addAppLabel(cr.Spec.AppName, &networkPolicy.ObjectMeta)
		setIngressPolicyType(networkPolicy)

		networkPolicy.Spec.PodSelector.MatchLabels = map[string]string{"name": "postgresql"}

		pod := orchestratorPod(*c)
		if pod == nil {
			return nil
		}

		ensureIngressRule(networkPolicy)
		if len(networkPolicy.Spec.Ingress[0].From) != 2 {
			networkPolicy.Spec.Ingress[0].From = []extensionsv1beta1.NetworkPolicyPeer{
				extensionsv1beta1.NetworkPolicyPeer{},
				extensionsv1beta1.NetworkPolicyPeer{},
			}
		}
		if len(networkPolicy.Spec.Ingress[0].Ports) != 1 {
			networkPolicy.Spec.Ingress[0].Ports = []extensionsv1beta1.NetworkPolicyPort{
				extensionsv1beta1.NetworkPolicyPort{},
			}
		}
		orchestratedByLabelKey := cr.Spec.AppName + "-orchestrated-by"
		orchestratedByLabelValue := pod.Name
		tcp := corev1.ProtocolTCP
		networkPolicy.Spec.Ingress[0].From[0].PodSelector = &metav1.LabelSelector{}
		networkPolicy.Spec.Ingress[0].From[0].PodSelector.MatchLabels = map[string]string{"name": "orchestrator"}
		networkPolicy.Spec.Ingress[0].From[1].PodSelector = &metav1.LabelSelector{}
		networkPolicy.Spec.Ingress[0].From[1].PodSelector.MatchLabels = map[string]string{orchestratedByLabelKey: orchestratedByLabelValue}
		networkPolicy.Spec.Ingress[0].Ports[0].Protocol = &tcp
		networkPolicy.Spec.Ingress[0].Ports[0].Port = &intstr.IntOrString{Type: intstr.Int, IntVal: 5432}

		return nil
	}

	return networkPolicy, f
}

func newNetworkPolicy(cr *miqv1alpha1.ManageIQ, name string) *extensionsv1beta1.NetworkPolicy {
	return &extensionsv1beta1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.AppName + name,
			Namespace: cr.ObjectMeta.Namespace,
		},
	}
}

func setIngressPolicyType(networkPolicy *extensionsv1beta1.NetworkPolicy) {
	if len(networkPolicy.Spec.PolicyTypes) != 1 {
		networkPolicy.Spec.PolicyTypes = append(networkPolicy.Spec.PolicyTypes, "Ingress")
	}
	networkPolicy.Spec.PolicyTypes[0] = "Ingress"
}

func ensureIngressRule(networkPolicy *extensionsv1beta1.NetworkPolicy) {
	if len(networkPolicy.Spec.Ingress) != 1 {
		networkPolicy.Spec.Ingress = []extensionsv1beta1.NetworkPolicyIngressRule{
			extensionsv1beta1.NetworkPolicyIngressRule{},
		}
	}
}
