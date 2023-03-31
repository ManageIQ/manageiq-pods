package miqtools

import (
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func NetworkPolicyDefaultDeny(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*networkingv1.NetworkPolicy, controllerutil.MutateFn) {
	networkPolicy := newNetworkPolicy(cr, "default-deny")

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, networkPolicy, scheme); err != nil {
			return err
		}
		addAppLabel(cr.Spec.AppName, &networkPolicy.ObjectMeta)
		setIngressPolicyType(networkPolicy)

		networkPolicy.Spec.PodSelector.MatchLabels = map[string]string{"app": cr.Spec.AppName}

		return nil
	}

	return networkPolicy, f
}

func NetworkPolicyAllowInboundHttpd(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*networkingv1.NetworkPolicy, controllerutil.MutateFn) {
	networkPolicy := newNetworkPolicy(cr, "allow-inbound-httpd")

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, networkPolicy, scheme); err != nil {
			return err
		}
		addAppLabel(cr.Spec.AppName, &networkPolicy.ObjectMeta)
		setIngressPolicyType(networkPolicy)

		networkPolicy.Spec.PodSelector.MatchLabels = map[string]string{"name": "httpd"}

		ensureIngressRule(networkPolicy)
		setFirstIngressTCPPort(networkPolicy, 8080)
		if len(networkPolicy.Spec.Ingress[0].From) != 1 {
			networkPolicy.Spec.Ingress[0].From = []networkingv1.NetworkPolicyPeer{
				networkingv1.NetworkPolicyPeer{},
			}
		}
		networkPolicy.Spec.Ingress[0].From[0].IPBlock = &networkingv1.IPBlock{}
		networkPolicy.Spec.Ingress[0].From[0].IPBlock.CIDR = "0.0.0.0/0"

		return nil
	}

	return networkPolicy, f
}

func NetworkPolicyAllowHttpdApi(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*networkingv1.NetworkPolicy, controllerutil.MutateFn) {
	networkPolicy := newNetworkPolicy(cr, "allow-httpd-api")

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, networkPolicy, scheme); err != nil {
			return err
		}
		addAppLabel(cr.Spec.AppName, &networkPolicy.ObjectMeta)
		setIngressPolicyType(networkPolicy)

		networkPolicy.Spec.PodSelector.MatchLabels = map[string]string{"service": "web-service"}

		ensureIngressRule(networkPolicy)
		setFirstIngressTCPPort(networkPolicy, 3000)
		if len(networkPolicy.Spec.Ingress[0].From) != 1 {
			networkPolicy.Spec.Ingress[0].From = []networkingv1.NetworkPolicyPeer{
				networkingv1.NetworkPolicyPeer{},
			}
		}
		networkPolicy.Spec.Ingress[0].From[0].PodSelector = &metav1.LabelSelector{}
		networkPolicy.Spec.Ingress[0].From[0].PodSelector.MatchLabels = map[string]string{"name": "httpd"}

		return nil
	}

	return networkPolicy, f
}

func NetworkPolicyAllowHttpdUi(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*networkingv1.NetworkPolicy, controllerutil.MutateFn) {
	networkPolicy := newNetworkPolicy(cr, "allow-httpd-ui")

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, networkPolicy, scheme); err != nil {
			return err
		}
		addAppLabel(cr.Spec.AppName, &networkPolicy.ObjectMeta)
		setIngressPolicyType(networkPolicy)

		networkPolicy.Spec.PodSelector.MatchLabels = map[string]string{"service": "ui"}

		ensureIngressRule(networkPolicy)
		setFirstIngressTCPPort(networkPolicy, 3000)
		if len(networkPolicy.Spec.Ingress[0].From) != 1 {
			networkPolicy.Spec.Ingress[0].From = []networkingv1.NetworkPolicyPeer{
				networkingv1.NetworkPolicyPeer{},
			}
		}
		networkPolicy.Spec.Ingress[0].From[0].PodSelector = &metav1.LabelSelector{}
		networkPolicy.Spec.Ingress[0].From[0].PodSelector.MatchLabels = map[string]string{"name": "httpd"}

		return nil
	}

	return networkPolicy, f
}

func NetworkPolicyAllowHttpdRemoteConsole(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*networkingv1.NetworkPolicy, controllerutil.MutateFn) {
	networkPolicy := newNetworkPolicy(cr, "allow-httpd-remote-console")

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, networkPolicy, scheme); err != nil {
			return err
		}
		addAppLabel(cr.Spec.AppName, &networkPolicy.ObjectMeta)
		setIngressPolicyType(networkPolicy)

		networkPolicy.Spec.PodSelector.MatchLabels = map[string]string{"service": "remote-console"}

		ensureIngressRule(networkPolicy)
		setFirstIngressTCPPort(networkPolicy, 3000)
		if len(networkPolicy.Spec.Ingress[0].From) != 1 {
			networkPolicy.Spec.Ingress[0].From = []networkingv1.NetworkPolicyPeer{
				networkingv1.NetworkPolicyPeer{},
			}
		}
		networkPolicy.Spec.Ingress[0].From[0].PodSelector = &metav1.LabelSelector{}
		networkPolicy.Spec.Ingress[0].From[0].PodSelector.MatchLabels = map[string]string{"name": "httpd"}

		return nil
	}

	return networkPolicy, f
}

func NetworkPolicyAllowMemcached(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme, c *client.Client) (*networkingv1.NetworkPolicy, controllerutil.MutateFn) {
	networkPolicy := newNetworkPolicy(cr, "allow-memcached")

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
		setFirstIngressTCPPort(networkPolicy, 11211)
		if len(networkPolicy.Spec.Ingress[0].From) != 2 {
			networkPolicy.Spec.Ingress[0].From = []networkingv1.NetworkPolicyPeer{
				networkingv1.NetworkPolicyPeer{},
				networkingv1.NetworkPolicyPeer{},
			}
		}
		orchestratedByLabelKey := cr.Spec.AppName + "-orchestrated-by"
		orchestratedByLabelValue := pod.Name
		networkPolicy.Spec.Ingress[0].From[0].PodSelector = &metav1.LabelSelector{}
		networkPolicy.Spec.Ingress[0].From[0].PodSelector.MatchLabels = map[string]string{"name": "orchestrator"}
		networkPolicy.Spec.Ingress[0].From[1].PodSelector = &metav1.LabelSelector{}
		networkPolicy.Spec.Ingress[0].From[1].PodSelector.MatchLabels = map[string]string{orchestratedByLabelKey: orchestratedByLabelValue}

		return nil
	}

	return networkPolicy, f
}

func NetworkPolicyAllowPostgres(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme, c *client.Client) (*networkingv1.NetworkPolicy, controllerutil.MutateFn) {
	networkPolicy := newNetworkPolicy(cr, "allow-postgres")

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
		setFirstIngressTCPPort(networkPolicy, 5432)
		if len(networkPolicy.Spec.Ingress[0].From) != 2 {
			networkPolicy.Spec.Ingress[0].From = []networkingv1.NetworkPolicyPeer{
				networkingv1.NetworkPolicyPeer{},
				networkingv1.NetworkPolicyPeer{},
			}
		}
		orchestratedByLabelKey := cr.Spec.AppName + "-orchestrated-by"
		orchestratedByLabelValue := pod.Name
		networkPolicy.Spec.Ingress[0].From[0].PodSelector = &metav1.LabelSelector{}
		networkPolicy.Spec.Ingress[0].From[0].PodSelector.MatchLabels = map[string]string{"name": "orchestrator"}
		networkPolicy.Spec.Ingress[0].From[1].PodSelector = &metav1.LabelSelector{}
		networkPolicy.Spec.Ingress[0].From[1].PodSelector.MatchLabels = map[string]string{orchestratedByLabelKey: orchestratedByLabelValue}

		return nil
	}

	return networkPolicy, f
}

func NetworkPolicyAllowKafka(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme, c *client.Client) (*networkingv1.NetworkPolicy, controllerutil.MutateFn) {
	networkPolicy := newNetworkPolicy(cr, "allow-kafka")

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, networkPolicy, scheme); err != nil {
			return err
		}
		addAppLabel(cr.Spec.AppName, &networkPolicy.ObjectMeta)
		setIngressPolicyType(networkPolicy)

		networkPolicy.Spec.PodSelector.MatchLabels = map[string]string{"name": "kafka"}

		pod := orchestratorPod(*c)
		if pod == nil {
			return nil
		}

		ensureIngressRule(networkPolicy)
		setFirstIngressTCPPort(networkPolicy, 9092)
		if len(networkPolicy.Spec.Ingress[0].From) != 2 {
			networkPolicy.Spec.Ingress[0].From = []networkingv1.NetworkPolicyPeer{
				networkingv1.NetworkPolicyPeer{},
				networkingv1.NetworkPolicyPeer{},
			}
		}
		orchestratedByLabelKey := cr.Spec.AppName + "-orchestrated-by"
		orchestratedByLabelValue := pod.Name
		networkPolicy.Spec.Ingress[0].From[0].PodSelector = &metav1.LabelSelector{}
		networkPolicy.Spec.Ingress[0].From[0].PodSelector.MatchLabels = map[string]string{"name": "orchestrator"}
		networkPolicy.Spec.Ingress[0].From[1].PodSelector = &metav1.LabelSelector{}
		networkPolicy.Spec.Ingress[0].From[1].PodSelector.MatchLabels = map[string]string{orchestratedByLabelKey: orchestratedByLabelValue}

		return nil
	}

	return networkPolicy, f
}

func newNetworkPolicy(cr *miqv1alpha1.ManageIQ, name string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.AppName + "-" + name,
			Namespace: cr.ObjectMeta.Namespace,
		},
	}
}

func setIngressPolicyType(networkPolicy *networkingv1.NetworkPolicy) {
	if len(networkPolicy.Spec.PolicyTypes) != 1 {
		networkPolicy.Spec.PolicyTypes = append(networkPolicy.Spec.PolicyTypes, "Ingress")
	}
	networkPolicy.Spec.PolicyTypes[0] = "Ingress"
}

func ensureIngressRule(networkPolicy *networkingv1.NetworkPolicy) {
	if len(networkPolicy.Spec.Ingress) != 1 {
		networkPolicy.Spec.Ingress = []networkingv1.NetworkPolicyIngressRule{
			networkingv1.NetworkPolicyIngressRule{},
		}
	}
}

func setFirstIngressTCPPort(networkPolicy *networkingv1.NetworkPolicy, port int32) {
	if len(networkPolicy.Spec.Ingress[0].Ports) != 1 {
		networkPolicy.Spec.Ingress[0].Ports = []networkingv1.NetworkPolicyPort{
			networkingv1.NetworkPolicyPort{},
		}
	}
	tcp := corev1.ProtocolTCP
	ports := &networkPolicy.Spec.Ingress[0].Ports[0]
	ports.Protocol = &tcp
	ports.Port = &intstr.IntOrString{Type: intstr.Int, IntVal: port}
}
