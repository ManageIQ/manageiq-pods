package manageiq

import (
	"context"

	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"

	miqtool "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/helpers/miq-components"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var logger = log.Log.WithName("controller_manageiq")

// Add creates a new ManageIQ Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileManageIQ{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("manageiq-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ManageIQ
	err = c.Watch(&source.Kind{Type: &miqv1alpha1.ManageIQ{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	ownerHandler := &handler.EnqueueRequestForOwner{IsController: true, OwnerType: &miqv1alpha1.ManageIQ{}}
	// Watch for changes to secondary resource Deployments and requeue the owner ManageIQ
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, ownerHandler)

	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, ownerHandler)
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, ownerHandler)
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, ownerHandler)
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, ownerHandler)
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileManageIQ implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileManageIQ{}

// ReconcileManageIQ reconciles a ManageIQ object
type ReconcileManageIQ struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ManageIQ object and makes changes based on the state read
// and what is in the ManageIQ.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileManageIQ) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := logger.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ManageIQ")

	// Fetch the ManageIQ instance
	miqInstance := &miqv1alpha1.ManageIQ{}

	err := r.client.Get(context.TODO(), request.NamespacedName, miqInstance)
	if errors.IsNotFound(err) {
		return reconcile.Result{}, nil
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

	return reconcile.Result{}, nil
}

func (r *ReconcileManageIQ) generateDefaultServiceAccount(cr *miqv1alpha1.ManageIQ) error {
	serviceAccount, mutateFunc := miqtool.DefaultServiceAccount(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, serviceAccount, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service Account has been reconciled", "component", "app", "result", result)
	}

	return nil
}

func (r *ReconcileManageIQ) generateHttpdResources(cr *miqv1alpha1.ManageIQ) error {
	privileged := miqtool.PrivilegedHttpd(cr.Spec.HttpdAuthenticationType)

	if privileged {
		httpdServiceAccount, mutateFunc := miqtool.HttpdServiceAccount(cr, r.scheme)
		if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, httpdServiceAccount, mutateFunc); err != nil {
			return err
		} else if result != controllerutil.OperationResultNone {
			logger.Info("ServiceAccount has been reconciled", "component", "httpd", "result", result)
		}
	}

	httpdConfigMap, mutateFunc, err := miqtool.HttpdConfigMap(cr, r.scheme)
	if err != nil {
		return err
	}
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, httpdConfigMap, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("ConfigMap has been reconciled", "component", "httpd", "result", result)
	}

	if cr.Spec.HttpdAuthenticationType != "internal" && cr.Spec.HttpdAuthenticationType != "openid-connect" {
		httpdAuthConfigMap, mutateFunc := miqtool.HttpdAuthConfigMap(cr, r.scheme)
		if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, httpdAuthConfigMap, mutateFunc); err != nil {
			return err
		} else if result != controllerutil.OperationResultNone {
			logger.Info("ConfigMap has been reconciled", "component", "httpd-auth", "result", result)
		}
	}

	uiService, mutateFunc := miqtool.UIService(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, uiService, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service has been reconciled", "component", "httpd", "service", "ui", "result", result)
	}

	webService, mutateFunc := miqtool.WebService(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, webService, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service has been reconciled", "component", "httpd", "service", "web_service", "result", result)
	}

	remoteConsoleService, mutateFunc := miqtool.RemoteConsoleService(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, remoteConsoleService, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service has been reconciled", "component", "httpd", "service", "remote_console_service", "result", result)
	}

	httpdService, mutateFunc := miqtool.HttpdService(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, httpdService, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service has been reconciled", "component", "httpd", "result", result)
	}

	if privileged {
		httpdDbusAPIService, mutateFunc := miqtool.HttpdDbusAPIService(cr, r.scheme)
		if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, httpdDbusAPIService, mutateFunc); err != nil {
			return err
		} else if result != controllerutil.OperationResultNone {
			logger.Info("Service has been reconciled", "component", "httpd", "service", "dbus_api_service", "result", result)
		}
	}

	httpdDeployment, mutateFunc, err := miqtool.HttpdDeployment(cr, r.scheme)
	if err != nil {
		return err
	}

	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, httpdDeployment, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Deployment has been reconciled", "component", "httpd", "result", result)
	}

	httpdIngress, mutateFunc := miqtool.Ingress(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, httpdIngress, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Ingress has been reconciled", "component", "httpd", "result", result)
	}

	return nil
}

func (r *ReconcileManageIQ) generateMemcachedResources(cr *miqv1alpha1.ManageIQ) error {
	deployment, mutateFunc, err := miqtool.NewMemcachedDeployment(cr, r.scheme)
	if err != nil {
		return err
	}

	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, deployment, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Deployment has been reconciled", "component", "memcached", "result", result)
	}

	service, mutateFunc := miqtool.NewMemcachedService(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, service, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service has been reconciled", "component", "memcached", "result", result)
	}

	return nil
}

func (r *ReconcileManageIQ) generatePostgresqlResources(cr *miqv1alpha1.ManageIQ) error {
	hostName := getSecretKeyValue(r.client, cr.Namespace, cr.Spec.DatabaseSecret, "hostname")
	if hostName != "" {
		logger.Info("External PostgreSQL Database selected, skipping postgresql service reconciliation", "hostname", hostName)
		return nil
	}

	secret := miqtool.DefaultPostgresqlSecret(cr)
	if err := r.createk8sResIfNotExist(cr, secret, &corev1.Secret{}); err != nil {
		return err
	}

	configMap, mutateFunc := miqtool.PostgresqlConfigMap(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, configMap, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("ConfigMap has been reconciled", "component", "postgresql", "result", result)
	}

	pvc, mutateFunc := miqtool.PostgresqlPVC(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, pvc, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("PVC has been reconciled", "component", "postgresql", "result", result)
	}

	service, mutateFunc := miqtool.PostgresqlService(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, service, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service has been reconciled", "component", "postgresql", "result", result)
	}

	deployment, mutateFunc, err := miqtool.PostgresqlDeployment(cr, r.scheme)
	if err != nil {
		return err
	}

	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, deployment, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Deployment has been reconciled", "component", "postgresql", "result", result)
	}

	return nil
}

func (r *ReconcileManageIQ) generateKafkaResources(cr *miqv1alpha1.ManageIQ) error {
	hostName := getSecretKeyValue(r.client, cr.Namespace, cr.Spec.KafkaSecret, "hostname")
	if hostName != "" {
		logger.Info("External Kafka Messaging Service selected, skipping kafka and zookeeper service reconciliation", "hostname", hostName)
		return nil
	}

	secret := miqtool.DefaultKafkaSecret(cr)
	if err := r.createk8sResIfNotExist(cr, secret, &corev1.Secret{}); err != nil {
		return err
	}

	kafkaPVC, mutateFunc := miqtool.KafkaPVC(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, kafkaPVC, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("PVC has been reconciled", "component", "kafka", "result", result)
	}

	zookeeperPVC, mutateFunc := miqtool.ZookeeperPVC(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, zookeeperPVC, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("PVC has been reconciled", "component", "zookeeper", "result", result)
	}

	kafkaService, mutateFunc := miqtool.KafkaService(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, kafkaService, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service has been reconciled", "component", "kafka", "result", result)
	}

	zookeeperService, mutateFunc := miqtool.ZookeeperService(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, zookeeperService, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service has been reconciled", "component", "zookeeper", "result", result)
	}

	kafkaDeployment, mutateFunc, err := miqtool.KafkaDeployment(cr, r.scheme)
	if err != nil {
		return err
	}

	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, kafkaDeployment, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Deployment has been reconciled", "component", "kafka", "result", result)
	}

	zookeeperDeployment, mutateFunc, err := miqtool.ZookeeperDeployment(cr, r.scheme)
	if err != nil {
		return err
	}

	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, zookeeperDeployment, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Deployment has been reconciled", "component", "zookeeper", "result", result)
	}

	return nil
}

func (r *ReconcileManageIQ) generateOrchestratorResources(cr *miqv1alpha1.ManageIQ) error {
	serviceAccount, mutateFunc := miqtool.OrchestratorServiceAccount(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, serviceAccount, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service Account has been reconciled", "component", "orchestrator", "result", result)
	}

	role, mutateFunc := miqtool.OrchestratorRole(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, role, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Role has been reconciled", "component", "orchestrator", "result", result)
	}

	roleBinding, mutateFunc := miqtool.OrchestratorRoleBinding(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, roleBinding, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Role Binding has been reconciled", "component", "orchestrator", "result", result)
	}

	deployment, mutateFunc, err := miqtool.OrchestratorDeployment(cr, r.scheme, r.client)
	if err != nil {
		return err
	}

	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, deployment, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Deployment has been reconciled", "component", "orchestrator", "result", result)
	}

	return nil
}

func (r *ReconcileManageIQ) generateNetworkPolicies(cr *miqv1alpha1.ManageIQ) error {
	networkPolicyDefaultDeny, mutateFunc := miqtool.NetworkPolicyDefaultDeny(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, networkPolicyDefaultDeny, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("NetworkPolicy default-deny has been reconciled", "component", "network_policy", "result", result)
	}

	networkPolicyAllowInboundHttpd, mutateFunc := miqtool.NetworkPolicyAllowInboundHttpd(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, networkPolicyAllowInboundHttpd, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("NetworkPolicy allow inbound-httpd has been reconciled", "component", "network_policy", "result", result)
	}

	networkPolicyAllowHttpdApi, mutateFunc := miqtool.NetworkPolicyAllowHttpdApi(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, networkPolicyAllowHttpdApi, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("NetworkPolicy allow httpd-api has been reconciled", "component", "network_policy", "result", result)
	}

	networkPolicyAllowHttpdUi, mutateFunc := miqtool.NetworkPolicyAllowHttpdUi(cr, r.scheme)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, networkPolicyAllowHttpdUi, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("NetworkPolicy allow httpd-ui has been reconciled", "component", "network_policy", "result", result)
	}

	networkPolicyAllowMemcached, mutateFunc := miqtool.NetworkPolicyAllowMemcached(cr, r.scheme, &r.client)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, networkPolicyAllowMemcached, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("NetworkPolicy allow memcached has been reconciled", "component", "network_policy", "result", result)
	}

	networkPolicyAllowPostgres, mutateFunc := miqtool.NetworkPolicyAllowPostgres(cr, r.scheme, &r.client)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, networkPolicyAllowPostgres, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("NetworkPolicy allow postgres has been reconciled", "component", "network_policy", "result", result)
	}

	return nil
}

func (r *ReconcileManageIQ) generateSecrets(cr *miqv1alpha1.ManageIQ) error {
	appSecret := miqtool.AppSecret(cr)
	if err := r.createk8sResIfNotExist(cr, appSecret, &corev1.Secret{}); err != nil {
		return err
	}

	tlsSecret, err := miqtool.TLSSecret(cr)
	if err != nil {
		return err
	}

	if err := r.createk8sResIfNotExist(cr, tlsSecret, &corev1.Secret{}); err != nil {
		return err
	}

	if cr.Spec.ImagePullSecret != "" {
		imagePullSecret, mutateFunc := miqtool.ImagePullSecret(cr, r.client)
		if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, imagePullSecret, mutateFunc); err != nil {
			return err
		} else if result != controllerutil.OperationResultNone {
			logger.Info("Image Pull Secret has been reconciled", "component", "operator", "result", result)
		}
	}

	if cr.Spec.HttpdAuthenticationType == "openid-connect" {
		if cr.Spec.OIDCClientSecret != "" {
			oidcClientSecret, mutateFunc := miqtool.OidcClientSecret(cr, r.client)
			if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, oidcClientSecret, mutateFunc); err != nil {
				return err
			} else if result != controllerutil.OperationResultNone {
				logger.Info("OIDC Client Secret has been reconciled", "component", "operator", "result", result)
			}
		}

		if cr.Spec.OIDCCACertSecret != "" {
			oidcCaCertSecret, mutateFunc := miqtool.OidcCaCertSecret(cr, r.client)
			if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, oidcCaCertSecret, mutateFunc); err != nil {
				return err
			} else if result != controllerutil.OperationResultNone {
				logger.Info("OIDC CA Secret has been reconciled", "component", "operator", "result", result)
			}
		}
	}

	return nil
}

func (r *ReconcileManageIQ) manageCR(cr *miqv1alpha1.ManageIQ) error {
	manageiq, mutateFunc := miqtool.ManageCR(cr, &r.client)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, manageiq, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("CR has been reconciled", "component", "app", "result", result)
	}

	return nil
}

func (r *ReconcileManageIQ) manageOperator(cr *miqv1alpha1.ManageIQ) error {
	operator, mutateFunc := miqtool.ManageOperator(cr, r.client)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, operator, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Operator has been reconciled", "component", "app", "result", result)
	}

	serviceAccount, mutateFunc := miqtool.ManageOperatorServiceAccount(cr, r.client)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, serviceAccount, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Service Account has been reconciled", "component", "operator", "result", result)
	}

	role, mutateFunc := miqtool.ManageOperatorRole(cr, r.client)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, role, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Role has been reconciled", "component", "operator", "result", result)
	}

	roleBinding, mutateFunc := miqtool.ManageOperatorRoleBinding(cr, r.client)
	if result, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, roleBinding, mutateFunc); err != nil {
		return err
	} else if result != controllerutil.OperationResultNone {
		logger.Info("Role Binding has been reconciled", "component", "operator", "result", result)
	}

	return nil
}

func (r *ReconcileManageIQ) createk8sResIfNotExist(cr *miqv1alpha1.ManageIQ, res, restype metav1.Object) error {
	reqLogger := logger.WithValues("task: ", "create resource")
	if err := controllerutil.SetControllerReference(cr, res, r.scheme); err != nil {
		return err
	}
	resClient := res.(runtime.Object)
	resTypeClient := restype.(runtime.Object)
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: res.GetName(), Namespace: res.GetNamespace()}, resTypeClient.(runtime.Object))
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating ", "Resource.Namespace", res.GetNamespace(), "Resource.Name", res.GetName())
		if err = r.client.Create(context.TODO(), resClient); err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
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
