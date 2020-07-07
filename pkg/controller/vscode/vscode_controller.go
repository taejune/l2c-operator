package vscode

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	tmaxv1alpha1 "tmax.io/l2c-operator/pkg/apis/tmax/v1alpha1"

	sonarapis "github.com/taejune/sonar-client-go/apis"
	sonarerrors "github.com/taejune/sonar-client-go/errors"
	"github.com/taejune/sonar-client-go/schemes"
	sonarschemes "github.com/taejune/sonar-client-go/schemes"
)

var log = logf.Log.WithName("controller_vscode")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new VSCode Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileVSCode{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("vscode-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource VSCode
	err = c.Watch(&source.Kind{Type: &tmaxv1alpha1.VSCode{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner VSCode
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1alpha1.VSCode{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1alpha1.VSCode{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1alpha1.VSCode{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1alpha1.VSCode{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileVSCode implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileVSCode{}

// ReconcileVSCode reconciles a VSCode object
type ReconcileVSCode struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a VSCode object and makes changes based on the state read
// and what is in the VSCode.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileVSCode) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling VSCode")

	// Fetch the VSCode instance
	instance := &tmaxv1alpha1.VSCode{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue

			token := schemes.Token{Login: "admin", Name: request.NamespacedName.Name}
			err = sonarapis.RevokeToken(token)
			if err != nil {
				reqLogger.Error(err, fmt.Sprintf("Failed to delete token for %s", request.NamespacedName.Name))
			}

			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Create Service
	svc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			vscodeSvc := r.NewServiceMeta(instance)
			reqLogger.Info("Creating a new Service", "Service.Namespace", vscodeSvc.Namespace, "Service.Name", vscodeSvc.Name)
			err = r.client.Create(context.TODO(), vscodeSvc)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		return reconcile.Result{}, err
	}

	// Create ConfigMap
	configmap := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, configmap)
	if err != nil {
		if errors.IsNotFound(err) {
			// Issue sonarqube token for VSCode's settings.json
			issue := sonarschemes.TokenIssue{Login: "admin", Name: instance.Name}
			token, err := sonarapis.GenerateToken(issue)
			if err != nil {
				if !sonarerrors.IsAlreadyExistsToken(err) {
					reqLogger.Error(err, "Faild to generate master token.")
					return reconcile.Result{}, err
				}
			}

			// XXX: How to inject sonarqube's service domain name?
			vscodeCm := r.NewConfigmapMeta(instance, token.Token)
			reqLogger.Info("Creating a new ConfigMap", "ConfigMap.Namespace", vscodeCm.Namespace, "ConfigMap.Name", vscodeCm.Name)
			err = r.client.Create(context.TODO(), vscodeCm)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		return reconcile.Result{}, err
	}

	// Create Secret
	secret := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			vscodeSecret := r.NewSecretMeta(instance)
			reqLogger.Info("Creating a new Secret", "Secret.Namespace", vscodeSecret.Namespace, "Secret.Name", vscodeSecret.Name)
			err = r.client.Create(context.TODO(), vscodeSecret)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, err
	}

	// Create Deployment
	deployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			vscodeDep := r.NewDepMeta(instance)
			reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", vscodeDep.Namespace, "Deployment.Name", vscodeDep.Name)
			err = r.client.Create(context.TODO(), vscodeDep)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, err
	}

	reqLogger.Info("Skip reconcile: All resource already exists", "Namespace", request.Namespace, "Name", request.Name)
	return reconcile.Result{}, nil
}
