package miqtools

import (
	"context"

	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	traefikIngressController      = "traefik.io/ingress-controller"
	defaultIngressClassAnnotation = "ingressclass.kubernetes.io/is-default-class"

	traefikServersTransportName = "httpd"
	traefikRedirectMiddleware   = "httpd-redirect-https"
	traefikHTTPIngressRoute     = "httpd-http"
)

func TraefikGVK(kind string) schema.GroupVersionKind {
	return schema.GroupVersionKind{Group: "traefik.io", Version: "v1alpha1", Kind: kind}
}

// IsTraefikIngress is true only when the default IngressClass is served by Traefik.
// Use an uncached reader: IngressClass is cluster scoped.
func IsTraefikIngress(reader client.Reader) (bool, error) {
	ingressClasses := &networkingv1.IngressClassList{}
	if err := reader.List(context.TODO(), ingressClasses); err != nil {
		return false, err
	}

	for _, ingressClass := range ingressClasses.Items {
		if ingressClass.Annotations[defaultIngressClassAnnotation] == "true" && ingressClass.Spec.Controller == traefikIngressController {
			return true, nil
		}
	}

	return false, nil
}

func newTraefikObject(cr *miqv1alpha1.ManageIQ, kind string, name string) *unstructured.Unstructured {
	object := &unstructured.Unstructured{}
	object.SetGroupVersionKind(TraefikGVK(kind))
	object.SetName(name)
	object.SetNamespace(cr.ObjectMeta.Namespace)
	return object
}

func traefikMutate(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme, object *unstructured.Unstructured, spec map[string]interface{}) controllerutil.MutateFn {
	return func() error {
		if err := controllerutil.SetControllerReference(cr, object, scheme); err != nil {
			return err
		}

		objectMeta := metav1.ObjectMeta{Labels: object.GetLabels()}
		addAppLabel(cr.Spec.AppName, &objectMeta)
		object.SetLabels(objectMeta.Labels)

		object.Object["spec"] = spec
		return nil
	}
}

func TraefikServersTransport(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*unstructured.Unstructured, controllerutil.MutateFn) {
	object := newTraefikObject(cr, "ServersTransport", traefikServersTransportName)
	spec := map[string]interface{}{"insecureSkipVerify": true}

	return object, traefikMutate(cr, scheme, object, spec)
}

// TraefikRedirectMiddleware replaces the nginx ssl-redirect annotation.
func TraefikRedirectMiddleware(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*unstructured.Unstructured, controllerutil.MutateFn) {
	object := newTraefikObject(cr, "Middleware", traefikRedirectMiddleware)
	spec := map[string]interface{}{
		"redirectScheme": map[string]interface{}{
			"scheme":    "https",
			"permanent": true,
		},
	}

	return object, traefikMutate(cr, scheme, object, spec)
}

func httpdTraefikRoutes(cr *miqv1alpha1.ManageIQ, middlewares []interface{}) []interface{} {
	route := map[string]interface{}{
		"kind":  "Rule",
		"match": "Host(`" + cr.Spec.ApplicationDomain + "`)",
		"services": []interface{}{
			map[string]interface{}{
				"name":             "httpd",
				"port":             int64(8080),
				"scheme":           "https",
				"serversTransport": traefikServersTransportName,
			},
		},
	}
	if len(middlewares) > 0 {
		route["middlewares"] = middlewares
	}

	return []interface{}{route}
}

func TraefikIngressRoute(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*unstructured.Unstructured, controllerutil.MutateFn) {
	object := newTraefikObject(cr, "IngressRoute", "httpd")
	spec := map[string]interface{}{
		"entryPoints": []interface{}{"websecure"},
		"routes":      httpdTraefikRoutes(cr, nil),
		"tls": map[string]interface{}{
			"secretName": tlsSecretName(cr),
		},
	}

	return object, traefikMutate(cr, scheme, object, spec)
}

func TraefikHTTPIngressRoute(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*unstructured.Unstructured, controllerutil.MutateFn) {
	object := newTraefikObject(cr, "IngressRoute", traefikHTTPIngressRoute)
	middlewares := []interface{}{map[string]interface{}{"name": traefikRedirectMiddleware}}
	spec := map[string]interface{}{
		"entryPoints": []interface{}{"web"},
		"routes":      httpdTraefikRoutes(cr, middlewares),
	}

	return object, traefikMutate(cr, scheme, object, spec)
}
