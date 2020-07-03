package sonarqube

import (
	"context"
	"fmt"

	sonarclient "github.com/taejune/sonar-client-go/client"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	tmaxv1alpha1 "tmax.io/l2c-operator/pkg/apis/tmax/v1alpha1"
)

var log = logf.Log.WithName("controller_sonarqube")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Sonarqube Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSonarqube{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("sonarqube-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Sonarqube
	err = c.Watch(&source.Kind{Type: &tmaxv1alpha1.Sonarqube{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Sonarqube
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1alpha1.Sonarqube{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileSonarqube implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSonarqube{}

// ReconcileSonarqube reconciles a Sonarqube object
type ReconcileSonarqube struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Sonarqube object and makes changes based on the state read
// and what is in the Sonarqube.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSonarqube) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Sonarqube")

	// Fetch the Sonarqube instance
	cr := &tmaxv1alpha1.Sonarqube{}
	err := r.client.Get(context.TODO(), request.NamespacedName, cr)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	svc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Create sonarqube Service for L2C")
			svc = newServiceForSonar(cr)
			// Set Sonarqube instance as the owner and controller
			if err := controllerutil.SetControllerReference(cr, svc, r.scheme); err != nil {
				return reconcile.Result{}, err
			}
			err := r.client.Create(context.TODO(), svc)
			if err != nil {
				reqLogger.Error(err, "Failed to create service for L2C sonarqube")
				return reconcile.Result{}, err
			}
		}
		reqLogger.Error(err, "Failed to get service for L2C sonarqube")
		return reconcile.Result{}, err
	}

	if len(svc.Status.LoadBalancer.Ingress) < 1 {
		reqLogger.Info("Cannot get service IP")
		return reconcile.Result{Requeue: true}, nil
	}

	addr := fmt.Sprintf("http://%s:%d", svc.Status.LoadBalancer.Ingress[0].IP, 9000)
	sonarclient.SetConfig(addr, "admin", "admin")

	dep := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, dep)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep = newDeploymentForSonar(cr)
		// Set Sonarqube instance as the owner and controller
		if err := controllerutil.SetControllerReference(cr, dep, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	}

	reqLogger.Info("Sonarqube for L2C is ready")
	return reconcile.Result{}, nil
}
