package manageiq

import (
	"context"

	miqv1alpha1 "github.com/manageiq-operator/pkg/apis/manageiq/v1alpha1"

	miqtool "github.com/manageiq-operator/pkg/helpers/miq-components"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extenv1beta1 "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
var currentAppName string

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Manageiq Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ManageiqReconciler{client: mgr.GetClient(), scheme: mgr.GetScheme()}
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

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Manageiq
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &miqv1alpha1.Manageiq{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ManageiqReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &ManageiqReconciler{}

// ManageiqReconciler reconciles a Manageiq object
type ManageiqReconciler struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Manageiq object and makes changes based on the state read
// and what is in the Manageiq.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ManageiqReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Manageiq")

	// Fetch the Manageiq instance
	miqInstance := &miqv1alpha1.Manageiq{}
	err := r.client.Get(context.TODO(), request.NamespacedName, miqInstance)

	if errors.IsNotFound(err) {
		err = CleanUpOrchestratedDeployments(miqInstance, r)
		return reconcile.Result{}, nil
	}

	currentAppName = miqInstance.Spec.AppName

	err = GenerateRbacResources(miqInstance, r)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = GenerateSecrets(miqInstance, r)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = GeneratePostgresqlResources(miqInstance, r)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = GenerateHttpdResources(miqInstance, r)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = GenerateMemcachedResources(miqInstance, r)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = GenerateOrchestratorResources(miqInstance, r)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func GenerateHttpdResources(cr *miqv1alpha1.Manageiq, r *ManageiqReconciler) error {
	HttpdIngress := miqtool.NewIngress(cr)
	HttpdConfigMap := miqtool.NewHttpdConfigMap(cr)
	HttpdAuthConfigMap := miqtool.NewHttpdAuthConfigMap(cr)
	HttpdService := miqtool.NewHttpdService(cr)
	HttpdDbusAPIService := miqtool.NewHttpdDbusAPIService(cr)

	UIService := miqtool.NewUIService(cr)
	WebService := miqtool.NewWebService(cr)
	RemoteConsoleService := miqtool.NewRemoteConsoleService(cr)

	HttpdDeployment := miqtool.NewHttpdDeployment(cr)

	err := Createk8sResIfNotExist(cr, HttpdDeployment, &appsv1.Deployment{}, r)
	if err != nil {
		return err
	}
	err = Createk8sResIfNotExist(HttpdDeployment, HttpdConfigMap, &corev1.ConfigMap{}, r)
	if err != nil {
		return err
	}
	err = Createk8sResIfNotExist(HttpdDeployment, HttpdAuthConfigMap, &corev1.ConfigMap{}, r)
	if err != nil {
		return err
	}
	err = Createk8sResIfNotExist(HttpdDeployment, UIService, &corev1.Service{}, r)
	if err != nil {
		return err
	}
	err = Createk8sResIfNotExist(HttpdDeployment, WebService, &corev1.Service{}, r)
	if err != nil {
		return err
	}
	err = Createk8sResIfNotExist(HttpdDeployment, RemoteConsoleService, &corev1.Service{}, r)
	if err != nil {
		return err
	}
	err = Createk8sResIfNotExist(HttpdDeployment, HttpdService, &corev1.Service{}, r)
	if err != nil {
		return err
	}
	err = Createk8sResIfNotExist(HttpdDeployment, HttpdDbusAPIService, &corev1.Service{}, r)
	if err != nil {
		return err
	}
	err = Createk8sResIfNotExist(HttpdService, HttpdIngress, &extenv1beta1.Ingress{}, r)
	if err != nil {
		return err
	}

	return err
}

func GenerateMemcachedResources(cr *miqv1alpha1.Manageiq, r *ManageiqReconciler) error {
	MemcachedService := miqtool.NewMemcachedService(cr)
	MemcachedDeployment := miqtool.NewMemcachedDeployment(cr)

	err := Createk8sResIfNotExist(cr, MemcachedDeployment, &appsv1.Deployment{}, r)
	if err != nil {
		return err
	}
	err = Createk8sResIfNotExist(MemcachedDeployment, MemcachedService, &corev1.Service{}, r)
	if err != nil {
		return err
	}
	return err
}

func GeneratePostgresqlResources(cr *miqv1alpha1.Manageiq, r *ManageiqReconciler) error {
	PostgresqlService := miqtool.NewPostgresqlService(cr)
	PostgresqlPVC := miqtool.NewPostgresqlPVC(cr)
	PostgresqlConfigsConfigMap := miqtool.NewPostgresqlConfigsConfigMap(cr)
	PostgresqlDeployment := miqtool.NewPostgresqlDeployment(cr)

	err := Createk8sResIfNotExist(cr, PostgresqlConfigsConfigMap, &corev1.ConfigMap{}, r)
	if err != nil {
		return err
	}

	err = Createk8sResIfNotExist(cr, PostgresqlPVC, &corev1.PersistentVolumeClaim{}, r)
	if err != nil {
		return err
	}

	err = Createk8sResIfNotExist(cr, PostgresqlService, &corev1.Service{}, r)
	if err != nil {
		return err
	}

	err = Createk8sResIfNotExist(cr, PostgresqlDeployment, &appsv1.Deployment{}, r)
	if err != nil {
		return err
	}

	return err
}

func GenerateOrchestratorResources(cr *miqv1alpha1.Manageiq, r *ManageiqReconciler) error {
	OrchestratorDeployment := miqtool.NewOrchestratorDeployment(cr)
	err := Createk8sResIfNotExist(cr, OrchestratorDeployment, &appsv1.Deployment{}, r)
	if err != nil {
		return err
	}

	return err
}

func GenerateSecrets(cr *miqv1alpha1.Manageiq, r *ManageiqReconciler) error {

	PostgresqlSecret := miqtool.NewPostgresqlSecret(cr)
	AppSecret := miqtool.AppSecret(cr)
	TLSSecret := miqtool.TLSSecret(cr)

	err := Createk8sResIfNotExist(cr, AppSecret, &corev1.Secret{}, r)
	if err != nil {
		return err
	}

	err = Createk8sResIfNotExist(cr, TLSSecret, &corev1.Secret{}, r)
	if err != nil {
		return err
	}

	err = Createk8sResIfNotExist(cr, PostgresqlSecret, &corev1.Secret{}, r)
	if err != nil {
		return err
	}

	return nil
}

func GenerateRbacResources(cr *miqv1alpha1.Manageiq, r *ManageiqReconciler) error {

	HttpdServiceAccount := miqtool.HttpdServiceAccount(cr)
	OrchestratorServiceAccount := miqtool.OrchestratorServiceAccount(cr)
	AnyuidServiceAccount := miqtool.AnyuidServiceAccount(cr)
	OrchestratorViewRoleBinding := miqtool.OrchestratorViewRoleBinding(cr)
	OrchestratorEditRoleBinding := miqtool.OrchestratorEditRoleBinding(cr)

	err := Createk8sResIfNotExist(cr, HttpdServiceAccount, &corev1.ServiceAccount{}, r)
	if err != nil {
		return err
	}
	err = Createk8sResIfNotExist(cr, OrchestratorServiceAccount, &corev1.ServiceAccount{}, r)
	if err != nil {
		return err
	}
	err = Createk8sResIfNotExist(cr, AnyuidServiceAccount, &corev1.ServiceAccount{}, r)
	if err != nil {
		return err
	}
	err = Createk8sResIfNotExist(cr, OrchestratorViewRoleBinding, &rbacv1.RoleBinding{}, r)
	if err != nil {
		return err
	}
	err = Createk8sResIfNotExist(cr, OrchestratorEditRoleBinding, &rbacv1.RoleBinding{}, r)
	if err != nil {
		return err
	}

	return nil
}

func Createk8sResIfNotExist(owner, res, restype metav1.Object, r *ManageiqReconciler) error {

	reqLogger := log.WithValues("task: ", "create resource")
	if err := controllerutil.SetControllerReference(owner, res, r.scheme); err != nil {
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

func CleanUpOrchestratedDeployments(cr *miqv1alpha1.Manageiq, r *ManageiqReconciler) error {
	reqLogger := log.WithValues("task: ", "Clean up resources")
	reqLogger.Info("Cleaning up orchestrated resources")
	gracePeriod := int64(0)
	deleteOpFunc := client.GracePeriodSeconds(gracePeriod)

	label := ManageIQAppLabel(cr)
	DepList := &appsv1.DeploymentList{}

	labelSelector := labels.SelectorFromSet(label)
	listOps := &client.ListOptions{Namespace: cr.Namespace, LabelSelector: labelSelector}

	err := r.client.List(context.TODO(), listOps, DepList)
	for _, item := range DepList.Items {
		err = r.client.Delete(context.TODO(), &item, deleteOpFunc)
		if err != nil {
			return err
		}
	}

	PodList := &corev1.PodList{}
	err = r.client.List(context.TODO(), listOps, PodList)
	for _, item := range PodList.Items {
		err = r.client.Delete(context.TODO(), &item, deleteOpFunc)
		if err != nil {
			return err
		}
	}

	return err
}

func ManageIQAppLabel(cr *miqv1alpha1.Manageiq) map[string]string {
	return map[string]string{
		"app": currentAppName,
	}
}