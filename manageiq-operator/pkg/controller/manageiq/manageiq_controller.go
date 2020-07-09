package manageiq

import (
	"context"

	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"

	miqtool "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/helpers/miq-components"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extenv1beta1 "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_manageiq")

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
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ManageIQ")

	// Fetch the ManageIQ instance
	miqInstance := &miqv1alpha1.ManageIQ{}

	err := r.client.Get(context.TODO(), request.NamespacedName, miqInstance)
	if errors.IsNotFound(err) {
		return reconcile.Result{}, nil
	}

	miqInstance.Initialize()

	if e := miqInstance.Validate(); e != nil {
		return reconcile.Result{}, e
	}

	if e := r.generateSecrets(miqInstance); e != nil {
		return reconcile.Result{}, e
	}
	if e := r.generatePostgresqlResources(miqInstance); e != nil {
		return reconcile.Result{}, e
	}
	if e := r.generateHttpdResources(miqInstance); e != nil {
		return reconcile.Result{}, e
	}
	if e := r.generateMemcachedResources(miqInstance); e != nil {
		return reconcile.Result{}, e
	}
	if *miqInstance.Spec.DeployMessagingService {
		if e := r.generateKafkaResources(miqInstance); e != nil {
			return reconcile.Result{}, e
		}
	}
	if e := r.generateOrchestratorResources(miqInstance); e != nil {
		return reconcile.Result{}, e
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileManageIQ) generateHttpdResources(cr *miqv1alpha1.ManageIQ) error {
	privileged, err := miqtool.PrivilegedHttpd(cr.Spec.HttpdAuthenticationType)
	if err != nil {
		return err
	}

	if privileged {
		httpdServiceAccount := miqtool.HttpdServiceAccount(cr)
		if err := r.createk8sResIfNotExist(cr, httpdServiceAccount, &corev1.ServiceAccount{}); err != nil {
			return err
		}
	}

	httpdConfigMap, err := miqtool.NewHttpdConfigMap(cr)
	if err != nil {
		return err
	}
	if err := r.createk8sResIfNotExist(cr, httpdConfigMap, &corev1.ConfigMap{}); err != nil {
		return err
	}

	if cr.Spec.HttpdAuthenticationType != "internal" && cr.Spec.HttpdAuthenticationType != "openid-connect" {
		httpdAuthConfigMap := miqtool.NewHttpdAuthConfigMap(cr)
		if err := r.createk8sResIfNotExist(cr, httpdAuthConfigMap, &corev1.ConfigMap{}); err != nil {
			return err
		}
	}

	uiService := miqtool.NewUIService(cr)
	if err := r.createk8sResIfNotExist(cr, uiService, &corev1.Service{}); err != nil {
		return err
	}

	webService := miqtool.NewWebService(cr)
	if err := r.createk8sResIfNotExist(cr, webService, &corev1.Service{}); err != nil {
		return err
	}

	remoteConsoleService := miqtool.NewRemoteConsoleService(cr)
	if err := r.createk8sResIfNotExist(cr, remoteConsoleService, &corev1.Service{}); err != nil {
		return err
	}

	httpdService := miqtool.NewHttpdService(cr)
	if err := r.createk8sResIfNotExist(cr, httpdService, &corev1.Service{}); err != nil {
		return err
	}

	if privileged {
		httpdDbusAPIService := miqtool.NewHttpdDbusAPIService(cr)
		if err := r.createk8sResIfNotExist(cr, httpdDbusAPIService, &corev1.Service{}); err != nil {
			return err
		}
	}

	httpdDeployment, err := miqtool.NewHttpdDeployment(cr)
	if err != nil {
		return err
	}
	if err := r.createk8sResIfNotExist(cr, httpdDeployment, &appsv1.Deployment{}); err != nil {
		return err
	}

	httpdIngress := miqtool.NewIngress(cr)
	if err := r.createk8sResIfNotExist(cr, httpdIngress, &extenv1beta1.Ingress{}); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileManageIQ) generateMemcachedResources(cr *miqv1alpha1.ManageIQ) error {
	memcachedDeployment, err := miqtool.NewMemcachedDeployment(cr)
	if err != nil {
		return err
	}
	if err := r.createk8sResIfNotExist(cr, memcachedDeployment, &appsv1.Deployment{}); err != nil {
		return err
	}

	memcachedService := miqtool.NewMemcachedService(cr)
	if err := r.createk8sResIfNotExist(cr, memcachedService, &corev1.Service{}); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileManageIQ) generatePostgresqlResources(cr *miqv1alpha1.ManageIQ) error {
	postgresqlSecret := miqtool.DefaultPostgresqlSecret(cr)
	if err := r.createk8sResIfNotExist(cr, postgresqlSecret, &corev1.Secret{}); err != nil {
		return err
	}

	postgresqlConfigsConfigMap := miqtool.NewPostgresqlConfigsConfigMap(cr)
	if err := r.createk8sResIfNotExist(cr, postgresqlConfigsConfigMap, &corev1.ConfigMap{}); err != nil {
		return err
	}

	postgresqlPVC := miqtool.NewPostgresqlPVC(cr)
	if err := r.createk8sResIfNotExist(cr, postgresqlPVC, &corev1.PersistentVolumeClaim{}); err != nil {
		return err
	}

	postgresqlService := miqtool.NewPostgresqlService(cr)
	if err := r.createk8sResIfNotExist(cr, postgresqlService, &corev1.Service{}); err != nil {
		return err
	}

	postgresqlDeployment, err := miqtool.NewPostgresqlDeployment(cr)
	if err != nil {
		return err
	}
	if err := r.createk8sResIfNotExist(cr, postgresqlDeployment, &appsv1.Deployment{}); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileManageIQ) generateKafkaResources(cr *miqv1alpha1.ManageIQ) error {
	secret := miqtool.DefaultKafkaSecret(cr)
	if err := r.createk8sResIfNotExist(cr, secret, &corev1.Secret{}); err != nil {
		return err
	}

	kafkaPVC := miqtool.KafkaPVC(cr)
	if err := r.createk8sResIfNotExist(cr, kafkaPVC, &corev1.PersistentVolumeClaim{}); err != nil {
		return err
	}

	zookeeperPVC := miqtool.ZookeeperPVC(cr)
	if err := r.createk8sResIfNotExist(cr, zookeeperPVC, &corev1.PersistentVolumeClaim{}); err != nil {
		return err
	}

	kafkaService := miqtool.KafkaService(cr)
	if err := r.createk8sResIfNotExist(cr, kafkaService, &corev1.Service{}); err != nil {
		return err
	}

	zookeeperService := miqtool.ZookeeperService(cr)
	if err := r.createk8sResIfNotExist(cr, zookeeperService, &corev1.Service{}); err != nil {
		return err
	}

	kafkaDeployment, err := miqtool.KafkaDeployment(cr)
	if err != nil {
		return err
	}
	if err := r.createk8sResIfNotExist(cr, kafkaDeployment, &appsv1.Deployment{}); err != nil {
		return err
	}

	zookeeperDeployment, err := miqtool.ZookeeperDeployment(cr)
	if err != nil {
		return err
	}
	if err := r.createk8sResIfNotExist(cr, zookeeperDeployment, &appsv1.Deployment{}); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileManageIQ) generateOrchestratorResources(cr *miqv1alpha1.ManageIQ) error {
	orchestratorServiceAccount := miqtool.OrchestratorServiceAccount(cr)
	if err := r.createk8sResIfNotExist(cr, orchestratorServiceAccount, &corev1.ServiceAccount{}); err != nil {
		return err
	}

	orchestratorRole := miqtool.OrchestratorRole(cr)
	if err := r.createk8sResIfNotExist(cr, orchestratorRole, &rbacv1.Role{}); err != nil {
		return err
	}

	orchestratorRoleBinding := miqtool.OrchestratorRoleBinding(cr)
	if err := r.createk8sResIfNotExist(cr, orchestratorRoleBinding, &rbacv1.RoleBinding{}); err != nil {
		return err
	}

	orchestratorDeployment, err := miqtool.NewOrchestratorDeployment(cr)
	if err != nil {
		return err
	}
	if err := r.createk8sResIfNotExist(cr, orchestratorDeployment, &appsv1.Deployment{}); err != nil {
		return err
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

	return nil
}

func (r *ReconcileManageIQ) createk8sResIfNotExist(cr *miqv1alpha1.ManageIQ, res, restype metav1.Object) error {
	reqLogger := log.WithValues("task: ", "create resource")
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
