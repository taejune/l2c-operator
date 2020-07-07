package l2c

import (
	"context"

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

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"tmax.io/l2c-operator/internal/schemes"
	"tmax.io/l2c-operator/internal/utils"
	l2cv1 "tmax.io/l2c-operator/pkg/apis/tmax/v1"
	tmaxv1alpha1 "tmax.io/l2c-operator/pkg/apis/tmax/v1alpha1"

	sonarapis "github.com/taejune/sonar-client-go/apis"
	sonarerrors "github.com/taejune/sonar-client-go/errors"
	sonarschemes "github.com/taejune/sonar-client-go/schemes"
)

var log = logf.Log.WithName("controller_l2c")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new L2C Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileL2C{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("l2c-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource L2C
	err = c.Watch(&source.Kind{Type: &l2cv1.L2C{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner L2C
	err = c.Watch(&source.Kind{Type: &tmaxv1alpha1.VSCode{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &l2cv1.L2C{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &l2cv1.L2C{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &l2cv1.L2C{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &l2cv1.L2C{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &tektonv1.Pipeline{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &l2cv1.L2C{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &tektonv1.PipelineResource{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &l2cv1.L2C{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileL2C implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileL2C{}

// ReconcileL2C reconciles a L2C object
type ReconcileL2C struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a L2C object and makes changes based on the state read
// and what is in the L2C.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileL2C) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling L2C")

	// Fetch the L2C instance
	l2c := &l2cv1.L2C{}
	err := r.client.Get(context.TODO(), request.NamespacedName, l2c)
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

	// TODO: SonarQube Project
	// TODO: If sonar project is ready (status.sonar-project-id exists) --> create all children objects

	proj := sonarschemes.NewProject(l2c.Spec.ProjectName, l2c.Spec.ProjectName, l2c.Spec.GitUrl, sonarschemes.Java, sonarschemes.MAVEN)
	err = sonarapis.CreateProject(proj)
	if err != nil {
		if !sonarerrors.IsAlreadyExists(err) {
			reqLogger.Error(err, "Failed to create project not caused by already exist")
			return reconcile.Result{}, err
		}
	}

	vscode := &tmaxv1alpha1.VSCode{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      l2c.Name + "-vscode",
		Namespace: l2c.Namespace,
	}, vscode)
	if err != nil {
		if errors.IsNotFound(err) {
			vscode = r.vscodeCr(l2c)
			err = r.client.Create(context.TODO(), vscode)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, err
	}

	// L2c's children
	// 1. Service (for WAS)
	// 2. ConfigMap (for both analyze-migrate & ci-cd)
	// 3. PipelineResource (git)
	// 4. PipelineResource (image)
	// 5. Pipeline (analyze-migrate)
	// 6. Pipeline (ci-cd)
	// 7. ServiceAccount
	// 8. RoleBinding

	// 1. Service (for WAS)
	wasService := schemes.Service(l2c)
	if err := controllerutil.SetControllerReference(l2c, wasService, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	if err = utils.CheckAndCreateObject(r.client, types.NamespacedName{Name: wasService.Name, Namespace: wasService.Namespace}, wasService); err != nil {
		if err = r.setStatus(l2c, l2cv1.StatusFailed, "Error creating children: "+err.Error()); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// 2. ConfigMap (for both analyze-migrate & ci-cd)
	configMap, err := schemes.ConfigMap(l2c)
	if err != nil {
		reqLogger.Error(err, "Error generating ConfigMap data")
		return reconcile.Result{}, err
	}
	if err := controllerutil.SetControllerReference(l2c, configMap, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	if err = utils.CheckAndCreateObject(r.client, types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, configMap); err != nil {
		if err = r.setStatus(l2c, l2cv1.StatusFailed, "Error creating children: "+err.Error()); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// 3.4. PipelineResource
	gitResource, imgResource := schemes.PipelineResource(l2c)

	// 3. PipelineResource (git)
	if err := controllerutil.SetControllerReference(l2c, gitResource, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	if err = utils.CheckAndCreateObject(r.client, types.NamespacedName{Name: gitResource.Name, Namespace: gitResource.Namespace}, gitResource); err != nil {
		if err = r.setStatus(l2c, l2cv1.StatusFailed, "Error creating children: "+err.Error()); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// 4. PipelineResource (image)
	if err := controllerutil.SetControllerReference(l2c, imgResource, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	if err = utils.CheckAndCreateObject(r.client, types.NamespacedName{Name: imgResource.Name, Namespace: imgResource.Namespace}, imgResource); err != nil {
		if err = r.setStatus(l2c, l2cv1.StatusFailed, "Error creating children: "+err.Error()); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// 5.6. Pipeline
	analyzeP, cicdP := schemes.Pipeline(l2c)

	// 5. Pipeline (analyze-migrate)
	if err := controllerutil.SetControllerReference(l2c, analyzeP, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	if err = utils.CheckAndCreateObject(r.client, types.NamespacedName{Name: analyzeP.Name, Namespace: analyzeP.Namespace}, analyzeP); err != nil {
		if err = r.setStatus(l2c, l2cv1.StatusFailed, "Error creating children: "+err.Error()); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// 6. Pipeline (ci-cd)
	if err := controllerutil.SetControllerReference(l2c, cicdP, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	if err = utils.CheckAndCreateObject(r.client, types.NamespacedName{Name: cicdP.Name, Namespace: cicdP.Namespace}, cicdP); err != nil {
		if err = r.setStatus(l2c, l2cv1.StatusFailed, "Error creating children: "+err.Error()); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// 7. ServiceAccount
	sa := schemes.ServiceAccount(l2c)
	if err := controllerutil.SetControllerReference(l2c, sa, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	if err = utils.CheckAndCreateObject(r.client, types.NamespacedName{Name: sa.Name, Namespace: sa.Namespace}, sa); err != nil {
		if err = r.setStatus(l2c, l2cv1.StatusFailed, "Error creating children: "+err.Error()); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// 8. RoleBinding
	rb := schemes.RoleBinding(l2c)
	if err := controllerutil.SetControllerReference(l2c, rb, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	if err = utils.CheckAndCreateObject(r.client, types.NamespacedName{Name: rb.Name, Namespace: rb.Namespace}, rb); err != nil {
		if err = r.setStatus(l2c, l2cv1.StatusFailed, "Error creating children: "+err.Error()); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// Set status as ready
	if err = r.setStatus(l2c, l2cv1.StatusReady, "All child objects are ready"); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileL2C) setStatus(cr *l2cv1.L2C, status l2cv1.Status, message string) error {
	reqLogger := log.WithValues("Request.Namespace", cr.Namespace, "Request.Name", cr.Name)
	cr.Status.Status = status
	cr.Status.Message = message
	if err := r.client.Status().Update(context.TODO(), cr); err != nil {
		reqLogger.Error(err, "Unknown error updating status")
		return err
	}
	return nil
}
