/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1"
	cr_migration "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1/helpers/cr_migration"
	miqtool "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1/helpers/miq-components"
)

// ManageIQReconciler reconciles a ManageIQ object
type ManageIQReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:namespace=changeme,groups="",resources=configmaps;events;persistentvolumeclaims;pods;pods/finalizers;secrets;serviceaccounts;services;services/finalizers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:namespace=changeme,groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:namespace=changeme,groups=apps,resources=deployments/finalizers,resourceNames=manageiq-operator,verbs=update
//+kubebuilder:rbac:namespace=changeme,groups=apps,resources=deployments/scale,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:namespace=changeme,groups=apps,resources=replicasets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:namespace=changeme,groups=coordination.k8s.io,resources=leases,verbs=get;list;create;update;delete
//+kubebuilder:rbac:namespace=changeme,groups=extensions,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:namespace=changeme,groups=extensions,resources=deployments/scale,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:namespace=changeme,groups=extensions,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:namespace=changeme,groups=manageiq.org,resources=manageiqs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:namespace=changeme,groups=manageiq.org,resources=manageiqs/finalizers,verbs=update
//+kubebuilder:rbac:namespace=changeme,groups=manageiq.org,resources=manageiqs/status,verbs=get;update;patch
//+kubebuilder:rbac:namespace=changeme,groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;create
//+kubebuilder:rbac:namespace=changeme,groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:namespace=changeme,groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:namespace=changeme,groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:namespace=changeme,groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:namespace=changeme,groups=route.openshift.io,resources=*,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ManageIQ object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ManageIQReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	reqLogger := logger.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ManageIQ")

	// Fetch the ManageIQ instance
	miqInstance := &miqv1alpha1.ManageIQ{}

	err := r.Client.Get(context.TODO(), request.NamespacedName, miqInstance)
	if errors.IsNotFound(err) {
		return reconcile.Result{}, nil
	}

	logger.Info("Migrating the CR...")
	if e := r.migrateCR(miqInstance); e != nil {
		return reconcile.Result{}, e
	}

	logger.Info("Reconciling the CR...")
	if e := r.manageCR(miqInstance); e != nil {
		return reconcile.Result{}, e
	}

	logger.Info("Validating the CR...")
	if e := miqInstance.Validate(); e != nil {
		return reconcile.Result{}, e
	}

	logger.Info("Reconciling the operator pod...")
	if os.Getenv("POD_NAME") != "" {
		if e := r.manageOperator(miqInstance); e != nil {
			return reconcile.Result{}, e
		}
	} else {
		logger.Info("Skipping reconcile of the operator pod; not running in a cluster.")
	}

	logger.Info("Reconciling the NetworkPolicies...")
	if e := r.generateNetworkPolicies(miqInstance); e != nil {
		return reconcile.Result{}, e
	}
	logger.Info("Reconciling the Secrets...")
	if e := r.generateSecrets(miqInstance); e != nil {
		return reconcile.Result{}, e
	}
	logger.Info("Reconciling the default Service Account...")
	if e := r.generateDefaultServiceAccount(miqInstance); e != nil {
		return reconcile.Result{}, e
	}
	logger.Info("Reconciling the Postgresql resources...")
	if e := r.generatePostgresqlResources(miqInstance); e != nil {
		return reconcile.Result{}, e
	}
	logger.Info("Reconciling the HTTPD resources...")
	if e := r.generateHttpdResources(miqInstance); e != nil {
		return reconcile.Result{}, e
	}
	logger.Info("Reconciling the Memcached resources...")
	if e := r.generateMemcachedResources(miqInstance); e != nil {
		return reconcile.Result{}, e
	}
	if *miqInstance.Spec.DeployMessagingService {
		logger.Info("Reconciling the Kafka resources...")
		if e := r.generateKafkaResources(miqInstance); e != nil {
			return reconcile.Result{}, e
		}
	}
	logger.Info("Reconciling the Orchestrator resources...")
	if e := r.generateOrchestratorResources(miqInstance); e != nil {
		return reconcile.Result{}, e
	}
	logger.Info("Reconciling the application resources...")
	if e := r.manageApplicationResources(miqInstance); e != nil {
		return reconcile.Result{}, e
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ManageIQReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&miqv1alpha1.ManageIQ{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Complete(r)
}

var logger = log.Log.WithName("controller_manageiq")

func (r *ManageIQReconciler) generateDefaultServiceAccount(cr *miqv1alpha1.ManageIQ) error {
	serviceAccount, mutateFunc := miqtool.DefaultServiceAccount(cr, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, serviceAccount, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service Account has been reconciled", "component", "app", "result", result)
	}

	return nil
}

func (r *ManageIQReconciler) generateHttpdResources(cr *miqv1alpha1.ManageIQ) error {
	privileged := miqtool.PrivilegedHttpd(cr.Spec.HttpdAuthenticationType)

	if privileged {
		httpdServiceAccount, mutateFunc := miqtool.HttpdServiceAccount(cr, r.Scheme)
		if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, httpdServiceAccount, mutateFunc); err != nil {
			return err
		} else if result != controllerutil.OperationResultNone {
			logger.Info("ServiceAccount has been reconciled", "component", "httpd", "result", result)
		}
	}

	httpdConfigMap, mutateFunc, err := miqtool.HttpdConfigMap(cr, r.Scheme, r.Client)
	if err != nil {
		return err
	}
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, httpdConfigMap, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("ConfigMap has been reconciled", "component", "httpd", "result", result)
	}

	if cr.Spec.HttpdAuthenticationType != "internal" && cr.Spec.HttpdAuthenticationType != "openid-connect" {
		httpdAuthConfigMap, mutateFunc := miqtool.HttpdAuthConfigMap(cr, r.Scheme)
		if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, httpdAuthConfigMap, mutateFunc); err != nil {
			return err
		} else if result != controllerutil.OperationResultNone {
			logger.Info("ConfigMap has been reconciled", "component", "httpd-auth", "result", result)
		}
	}

	if httpdAuthConfig, mutateFunc := miqtool.HttpdAuthConfig(r.Client, cr, r.Scheme); httpdAuthConfig != nil {
		if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, httpdAuthConfig, mutateFunc); err != nil {
			return err
		} else if result != controllerutil.OperationResultNone {
			logger.Info("Secret has been reconciled", "component", "httpd-auth", "result", result)
		}
	}

	uiService, mutateFunc := miqtool.UIService(cr, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, uiService, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service has been reconciled", "component", "httpd", "service", "ui", "result", result)
	}

	webService, mutateFunc := miqtool.WebService(cr, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, webService, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service has been reconciled", "component", "httpd", "service", "web_service", "result", result)
	}

	remoteConsoleService, mutateFunc := miqtool.RemoteConsoleService(cr, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, remoteConsoleService, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service has been reconciled", "component", "httpd", "service", "remote_console_service", "result", result)
	}

	httpdService, mutateFunc := miqtool.HttpdService(cr, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, httpdService, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service has been reconciled", "component", "httpd", "result", result)
	}

	if privileged {
		httpdDbusAPIService, mutateFunc := miqtool.HttpdDbusAPIService(cr, r.Scheme)
		if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, httpdDbusAPIService, mutateFunc); err != nil {
			return err
		} else if result != controllerutil.OperationResultNone {
			logger.Info("Service has been reconciled", "component", "httpd", "service", "dbus_api_service", "result", result)
		}
	}

	if err := r.reconcileHttpdDeployment(cr); err != nil {
		return err
	}

	if internalCerts := miqtool.InternalCertificatesSecret(cr, r.Client); internalCerts.Data["httpd_crt"] != nil {
		httpdRoute, mutateFunc := miqtool.Route(cr, r.Scheme, r.Client)
		if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, httpdRoute, mutateFunc); err != nil {
			return err
		} else if result != controllerutil.OperationResultNone {
			logger.Info("Route has been reconciled", "component", "httpd", "result", result)
		}
	} else {
		httpdIngress, mutateFunc := miqtool.Ingress(cr, r.Scheme)
		if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, httpdIngress, mutateFunc); err != nil {
			return err
		} else if result != controllerutil.OperationResultNone {
			logger.Info("Ingress has been reconciled", "component", "httpd", "result", result)
		}
	}

	return nil
}

func (r *ManageIQReconciler) reconcileHttpdDeployment(cr *miqv1alpha1.ManageIQ) error {
	httpdDeployment, mutateFunc, err := miqtool.HttpdDeployment(r.Client, cr, r.Scheme)
	if err != nil {
		return err
	}

	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, httpdDeployment, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Deployment has been reconciled", "component", "httpd", "result", result)
	}
	return nil
}

func (r *ManageIQReconciler) generateMemcachedResources(cr *miqv1alpha1.ManageIQ) error {
	deployment, mutateFunc, err := miqtool.NewMemcachedDeployment(cr, r.Scheme, r.Client)
	if err != nil {
		return err
	}

	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, deployment, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Deployment has been reconciled", "component", "memcached", "result", result)
	}

	service, mutateFunc := miqtool.NewMemcachedService(cr, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, service, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service has been reconciled", "component", "memcached", "result", result)
	}

	return nil
}

func (r *ManageIQReconciler) generatePostgresqlResources(cr *miqv1alpha1.ManageIQ) error {
	secret, mutateFunc := miqtool.ManagePostgresqlSecret(cr, r.Client, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, secret, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Secret has been reconciled", "component", "postgresql", "result", result)
	}

	hostName := string(secret.Data["hostname"])
	if hostName != "postgresql" {
		logger.Info("External PostgreSQL Database selected, skipping postgresql service reconciliation", "hostname", hostName)
		return nil
	}

	configMap, mutateFunc := miqtool.PostgresqlConfigMap(cr, r.Client, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, configMap, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("ConfigMap has been reconciled", "component", "postgresql", "result", result)
	}

	pvc, mutateFunc := miqtool.PostgresqlPVC(cr, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, pvc, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("PVC has been reconciled", "component", "postgresql", "result", result)
	}

	service, mutateFunc := miqtool.PostgresqlService(cr, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, service, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service has been reconciled", "component", "postgresql", "result", result)
	}

	deployment, mutateFunc, err := miqtool.PostgresqlDeployment(cr, r.Client, r.Scheme)
	if err != nil {
		return err
	}

	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, deployment, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Deployment has been reconciled", "component", "postgresql", "result", result)
	}

	return nil
}

func (r *ManageIQReconciler) generateKafkaResources(cr *miqv1alpha1.ManageIQ) error {
	secret, mutateFunc := miqtool.ManageKafkaSecret(cr, r.Client, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, secret, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Secret has been reconciled", "component", "kafka", "result", result)
	}

	hostName := getSecretKeyValue(r.Client, cr.Namespace, cr.Spec.KafkaSecret, "hostname")
	if hostName != "" {
		logger.Info("External Kafka Messaging Service selected, skipping kafka and zookeeper service reconciliation", "hostname", hostName)
		return nil
	}

	kafkaPVC, mutateFunc := miqtool.KafkaPVC(cr, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, kafkaPVC, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("PVC has been reconciled", "component", "kafka", "result", result)
	}

	zookeeperPVC, mutateFunc := miqtool.ZookeeperPVC(cr, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, zookeeperPVC, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("PVC has been reconciled", "component", "zookeeper", "result", result)
	}

	kafkaService, mutateFunc := miqtool.KafkaService(cr, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, kafkaService, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service has been reconciled", "component", "kafka", "result", result)
	}

	zookeeperService, mutateFunc := miqtool.ZookeeperService(cr, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, zookeeperService, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service has been reconciled", "component", "zookeeper", "result", result)
	}

	kafkaDeployment, mutateFunc, err := miqtool.KafkaDeployment(cr, r.Scheme)
	if err != nil {
		return err
	}

	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, kafkaDeployment, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Deployment has been reconciled", "component", "kafka", "result", result)
	}

	zookeeperDeployment, mutateFunc, err := miqtool.ZookeeperDeployment(cr, r.Scheme)
	if err != nil {
		return err
	}

	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, zookeeperDeployment, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Deployment has been reconciled", "component", "zookeeper", "result", result)
	}

	return nil
}

func (r *ManageIQReconciler) generateOrchestratorResources(cr *miqv1alpha1.ManageIQ) error {
	serviceAccount, mutateFunc := miqtool.OrchestratorServiceAccount(cr, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, serviceAccount, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service Account has been reconciled", "component", "orchestrator", "result", result)
	}

	role, mutateFunc := miqtool.OrchestratorRole(cr, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, role, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Role has been reconciled", "component", "orchestrator", "result", result)
	}

	roleBinding, mutateFunc := miqtool.OrchestratorRoleBinding(cr, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, roleBinding, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Role Binding has been reconciled", "component", "orchestrator", "result", result)
	}

	deployment, mutateFunc, err := miqtool.OrchestratorDeployment(cr, r.Scheme, r.Client)
	if err != nil {
		return err
	}

	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, deployment, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Deployment has been reconciled", "component", "orchestrator", "result", result)
	}

	return nil
}

func (r *ManageIQReconciler) generateNetworkPolicies(cr *miqv1alpha1.ManageIQ) error {
	networkPolicyDefaultDeny, mutateFunc := miqtool.NetworkPolicyDefaultDeny(cr, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, networkPolicyDefaultDeny, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("NetworkPolicy default-deny has been reconciled", "component", "network_policy", "result", result)
	}

	networkPolicyAllowInboundHttpd, mutateFunc := miqtool.NetworkPolicyAllowInboundHttpd(cr, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, networkPolicyAllowInboundHttpd, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("NetworkPolicy allow inbound-httpd has been reconciled", "component", "network_policy", "result", result)
	}

	networkPolicyAllowHttpdApi, mutateFunc := miqtool.NetworkPolicyAllowHttpdApi(cr, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, networkPolicyAllowHttpdApi, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("NetworkPolicy allow httpd-api has been reconciled", "component", "network_policy", "result", result)
	}

	networkPolicyAllowHttpdRemoteConsole, mutateFunc := miqtool.NetworkPolicyAllowHttpdRemoteConsole(cr, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, networkPolicyAllowHttpdRemoteConsole, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("NetworkPolicy allow httpd-remote-console has been reconciled", "component", "network_policy", "result", result)
	}

	networkPolicyAllowHttpdUi, mutateFunc := miqtool.NetworkPolicyAllowHttpdUi(cr, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, networkPolicyAllowHttpdUi, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("NetworkPolicy allow httpd-ui has been reconciled", "component", "network_policy", "result", result)
	}

	networkPolicyAllowMemcached, mutateFunc := miqtool.NetworkPolicyAllowMemcached(cr, r.Scheme, &r.Client)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, networkPolicyAllowMemcached, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("NetworkPolicy allow memcached has been reconciled", "component", "network_policy", "result", result)
	}

	networkPolicyAllowPostgres, mutateFunc := miqtool.NetworkPolicyAllowPostgres(cr, r.Scheme, &r.Client)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, networkPolicyAllowPostgres, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("NetworkPolicy allow postgres has been reconciled", "component", "network_policy", "result", result)
	}

	networkPolicyAllowZookeeper, mutateFunc := miqtool.NetworkPolicyAllowZookeeper(cr, r.Scheme, &r.Client)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, networkPolicyAllowZookeeper, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("NetworkPolicy allow zookeeper has been reconciled", "component", "network_policy", "result", result)
	}

	return nil
}

func (r *ManageIQReconciler) generateSecrets(cr *miqv1alpha1.ManageIQ) error {
	secret, mutateFunc := miqtool.ManageAppSecret(cr, r.Client, r.Scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, secret, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Secret has been reconciled", "component", "app-secret", "result", result)
	}

	secret, mutateFunc, err := miqtool.ManageTlsSecret(cr, r.Client, r.Scheme)
	if err != nil {
		return err
	}
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, secret, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Secret has been reconciled", "component", "tls-secret", "result", result)
	}

	if cr.Spec.ImagePullSecret != "" {
		imagePullSecret, mutateFunc := miqtool.ImagePullSecret(cr, r.Client)
		if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, imagePullSecret, mutateFunc); err != nil {
			return err
		} else if result != controllerutil.OperationResultNone {
			logger.Info("Image Pull Secret has been reconciled", "component", "operator", "result", result)
		}
	}

	if cr.Spec.HttpdAuthenticationType == "openid-connect" {
		if cr.Spec.OIDCClientSecret != "" {
			oidcClientSecret, mutateFunc := miqtool.OidcClientSecret(cr, r.Client)
			if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, oidcClientSecret, mutateFunc); err != nil {
				return err
			} else if result != controllerutil.OperationResultNone {
				logger.Info("OIDC Client Secret has been reconciled", "component", "operator", "result", result)
			}
		}

		if cr.Spec.OIDCCACertSecret != "" {
			oidcCaCertSecret, mutateFunc := miqtool.OidcCaCertSecret(cr, r.Client)
			if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, oidcCaCertSecret, mutateFunc); err != nil {
				return err
			} else if result != controllerutil.OperationResultNone {
				logger.Info("OIDC CA Secret has been reconciled", "component", "operator", "result", result)
			}
		}
	}

	return nil
}

func (r *ManageIQReconciler) migrateCR(cr *miqv1alpha1.ManageIQ) error {
	manageiq, mutateFunc := cr_migration.Migrate(cr)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, manageiq, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("CR has been migrated", "component", "app", "result", result)
	}

	return nil
}

func (r *ManageIQReconciler) manageCR(cr *miqv1alpha1.ManageIQ) error {
	manageiq, mutateFunc := miqtool.ManageCR(cr, &r.Client)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, manageiq, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("CR has been reconciled", "component", "app", "result", result)
	}

	return nil
}

func (r *ManageIQReconciler) manageOperator(cr *miqv1alpha1.ManageIQ) error {
	operator, mutateFunc := miqtool.ManageOperator(cr, r.Client)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, operator, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Operator has been reconciled", "component", "app", "result", result)
	}

	serviceAccount, mutateFunc := miqtool.ManageOperatorServiceAccount(cr, r.Client)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, serviceAccount, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service Account has been reconciled", "component", "operator", "result", result)
	}

	role, mutateFunc := miqtool.ManageOperatorRole(cr, r.Client)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, role, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Role has been reconciled", "component", "operator", "result", result)
	}

	roleBinding, mutateFunc := miqtool.ManageOperatorRoleBinding(cr, r.Client)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, roleBinding, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Role Binding has been reconciled", "component", "operator", "result", result)
	}

	return nil
}

func getSecretKeyValue(client client.Client, nameSpace string, secretName string, keyName string) string {
	secretKey := types.NamespacedName{Namespace: nameSpace, Name: secretName}
	secret := &corev1.Secret{}
	secretErr := client.Get(context.TODO(), secretKey, secret)
	if secretErr != nil {
		return ""
	}
	return string(secret.Data[keyName])
}

func (r *ManageIQReconciler) manageApplicationResources(cr *miqv1alpha1.ManageIQ) error {
	configMap, mutateFunc := miqtool.ApplicationUiHttpdConfigMap(cr, r.Scheme, r.Client)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, configMap, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("ConfigMap has been reconciled", "component", "application ui", "result", result)
	}

	configMap, mutateFunc = miqtool.ApplicationApiHttpdConfigMap(cr, r.Scheme, r.Client)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, configMap, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("ConfigMap has been reconciled", "component", "application api", "result", result)
	}

	configMap, mutateFunc = miqtool.ApplicationRemoteConsoleHttpdConfigMap(cr, r.Scheme, r.Client)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, configMap, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("ConfigMap has been reconciled", "component", "application remote console", "result", result)
	}

	return nil
}
