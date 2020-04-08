package manageiq

import (
	"context"

	miqv1alpha1 "github.com/manageiq/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"

	miqtool "github.com/manageiq/manageiq-pods/manageiq-operator/pkg/helpers/miq-components"
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

// Add creates a new Manageiq Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileManageiq{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("manageiq-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Manageiq
	err = c.Watch(&source.Kind{Type: &miqv1alpha1.Manageiq{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	ownerHandler := &handler.EnqueueRequestForOwner{IsController: true, OwnerType: &miqv1alpha1.Manageiq{}}
	// Watch for changes to secondary resource Deployments and requeue the owner Manageiq
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

// blank assignment to verify that ReconcileManageiq implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileManageiq{}

// ReconcileManageiq reconciles a Manageiq object
type ReconcileManageiq struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Manageiq object and makes changes based on the state read
// and what is in the Manageiq.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileManageiq) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Manageiq")

	// Fetch the Manageiq instance
	miqInstance := &miqv1alpha1.Manageiq{}
	err := r.client.Get(context.TODO(), request.NamespacedName, miqInstance)

	if errors.IsNotFound(err) {
		return reconcile.Result{}, nil
	}

	if e := r.generateRbacResources(miqInstance); e != nil {
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
	if e := r.generateOrchestratorResources(miqInstance); e != nil {
		return reconcile.Result{}, e
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileManageiq) generateHttpdResources(cr *miqv1alpha1.Manageiq) error {
	httpdConfigMap := miqtool.NewHttpdConfigMap(cr)
	if err := r.createk8sResIfNotExist(cr, httpdConfigMap, &corev1.ConfigMap{}); err != nil {
		return err
	}

	httpdAuthConfigMap := miqtool.NewHttpdAuthConfigMap(cr)
	if err := r.createk8sResIfNotExist(cr, httpdAuthConfigMap, &corev1.ConfigMap{}); err != nil {
		return err
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

	httpdDbusAPIService := miqtool.NewHttpdDbusAPIService(cr)
	if err := r.createk8sResIfNotExist(cr, httpdDbusAPIService, &corev1.Service{}); err != nil {
		return err
	}

	httpdDeployment := miqtool.NewHttpdDeployment(cr)
	if err := r.createk8sResIfNotExist(cr, httpdDeployment, &appsv1.Deployment{}); err != nil {
		return err
	}

	httpdIngress := miqtool.NewIngress(cr)
	if err := r.createk8sResIfNotExist(cr, httpdIngress, &extenv1beta1.Ingress{}); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileManageiq) generateMemcachedResources(cr *miqv1alpha1.Manageiq) error {
	memcachedDeployment := miqtool.NewMemcachedDeployment(cr)
	if err := r.createk8sResIfNotExist(cr, memcachedDeployment, &appsv1.Deployment{}); err != nil {
		return err
	}

	memcachedService := miqtool.NewMemcachedService(cr)
	if err := r.createk8sResIfNotExist(cr, memcachedService, &corev1.Service{}); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileManageiq) generatePostgresqlResources(cr *miqv1alpha1.Manageiq) error {
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

	postgresqlDeployment := miqtool.NewPostgresqlDeployment(cr)
	if err := r.createk8sResIfNotExist(cr, postgresqlDeployment, &appsv1.Deployment{}); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileManageiq) generateOrchestratorResources(cr *miqv1alpha1.Manageiq) error {
	orchestratorDeployment := miqtool.NewOrchestratorDeployment(cr)
	if err := r.createk8sResIfNotExist(cr, orchestratorDeployment, &appsv1.Deployment{}); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileManageiq) generateSecrets(cr *miqv1alpha1.Manageiq) error {
	appSecret := miqtool.AppSecret(cr)
	if err := r.createk8sResIfNotExist(cr, appSecret, &corev1.Secret{}); err != nil {
		return err
	}

	tlsSecret := miqtool.TLSSecret(cr)
	if err := r.createk8sResIfNotExist(cr, tlsSecret, &corev1.Secret{}); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileManageiq) generateRbacResources(cr *miqv1alpha1.Manageiq) error {
	httpdServiceAccount := miqtool.HttpdServiceAccount(cr)
	if err := r.createk8sResIfNotExist(cr, httpdServiceAccount, &corev1.ServiceAccount{}); err != nil {
		return err
	}

	orchestratorServiceAccount := miqtool.OrchestratorServiceAccount(cr)
	if err := r.createk8sResIfNotExist(cr, orchestratorServiceAccount, &corev1.ServiceAccount{}); err != nil {
		return err
	}

	anyuidServiceAccount := miqtool.AnyuidServiceAccount(cr)
	if err := r.createk8sResIfNotExist(cr, anyuidServiceAccount, &corev1.ServiceAccount{}); err != nil {
		return err
	}

	orchestratorViewRoleBinding := miqtool.OrchestratorViewRoleBinding(cr)
	if err := r.createk8sResIfNotExist(cr, orchestratorViewRoleBinding, &rbacv1.RoleBinding{}); err != nil {
		return err
	}

	orchestratorEditRoleBinding := miqtool.OrchestratorEditRoleBinding(cr)
	if err := r.createk8sResIfNotExist(cr, orchestratorEditRoleBinding, &rbacv1.RoleBinding{}); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileManageiq) createk8sResIfNotExist(cr *miqv1alpha1.Manageiq, res, restype metav1.Object) error {
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
