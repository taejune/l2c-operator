package l2crun

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
)

var log = logf.Log.WithName("controller_l2crun")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new L2CRun Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileL2CRun{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("l2crun-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource L2CRun
	err = c.Watch(&source.Kind{Type: &l2cv1.L2CRun{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner L2cRunSH
	err = c.Watch(&source.Kind{Type: &tektonv1.PipelineRun{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &l2cv1.L2CRun{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileL2CRun implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileL2CRun{}

// ReconcileL2CRun reconciles a L2CRun object
type ReconcileL2CRun struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a L2CRun object and makes changes based on the state read
// and what is in the L2CRun.Spec
func (r *ReconcileL2CRun) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling L2CRun")

	// Fetch the L2CRun l2crun
	l2crun := &l2cv1.L2CRun{}
	err := r.client.Get(context.TODO(), request.NamespacedName, l2crun)
	if err != nil {
		if kerrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Fill unfilled status fields
	if initStatusField(l2crun) {
		if err = r.client.Status().Update(context.TODO(), l2crun); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Get L2c object referred by l2crun
	l2c := &l2cv1.L2C{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: l2crun.Spec.L2cName, Namespace: l2crun.Namespace}, l2c)
	if err != nil {
		if kerrors.IsNotFound(err) {
			reqLogger.Error(err, "L2c not found")
			if err = r.setStatus(l2crun, l2cv1.PhaseNull, l2cv1.StatusFailed, "L2c ["+l2crun.Spec.L2cName+"] not found"); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
		reqLogger.Error(err, "Unknown error getting L2c")
		return reconcile.Result{}, err
	}

	status := l2crun.Status.Status

	// Succeeded / Failed --> Do nothing!
	if status == l2cv1.StatusSucceeded || status == l2cv1.StatusFailed {
		reqLogger.Info("Status is already " + string(status))
		return reconcile.Result{}, nil
	}

	// Get PRs
	analyzePr, cicdPr := schemes.PipelineRun(l2crun, l2c)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: analyzePr.Name, Namespace: analyzePr.Namespace}, analyzePr)
	if err != nil {
		if kerrors.IsNotFound(err) {
			analyzePr = nil
		} else {
			if err = r.setStatus(l2crun, l2cv1.PhaseNull, l2cv1.StatusFailed, "Error getting PipelineRun status: "+err.Error()); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, err
		}
	}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: cicdPr.Name, Namespace: cicdPr.Namespace}, cicdPr)
	if err != nil {
		if kerrors.IsNotFound(err) {
			cicdPr = nil
		} else {
			if err = r.setStatus(l2crun, l2cv1.PhaseNull, l2cv1.StatusFailed, "Error getting PipelineRun status: "+err.Error()); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, err
		}
	}

	// Pending
	if status == l2cv1.StatusPending {
		// If there exists PRs, unknown status
		if analyzePr != nil || cicdPr != nil {
			reqLogger.Error(errors.New("unknown status"), "Unknown status: pending but PipelineRun exists")
			if err = r.setStatus(l2crun, l2cv1.PhaseNull, l2cv1.StatusFailed, "Unknown status: pending but PipelineRun exists"); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}

		// Launch first pipelinerun
		// Set status as Running first
		time := metav1.Now()
		l2crun.Status.StartTime = &time
		if err = r.setStatus(l2crun, l2cv1.PhaseNull, l2cv1.StatusRunning, "Launched PipelineRun"); err != nil {
			return reconcile.Result{}, err
		}
		// Create pipelinerun
		analyzePr, _ = schemes.PipelineRun(l2crun, l2c)
		if err := controllerutil.SetControllerReference(l2crun, analyzePr, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		if err = utils.CheckAndCreateObject(r.client, types.NamespacedName{Name: analyzePr.Name, Namespace: analyzePr.Namespace}, analyzePr); err != nil {
			if err = r.setStatus(l2crun, l2cv1.PhaseNull, l2cv1.StatusFailed, "Error creating children: "+err.Error()); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, nil
	}

	// Running --> need to inspect pipelinerun (taskruns actually) status as well
	if status != l2cv1.StatusRunning || (analyzePr == nil && cicdPr == nil) {
		reqLogger.Error(errors.New("unknown status "+string(l2crun.Status.Status)), "Status is unknown")
		return reconcile.Result{}, nil
	}

	taskRuns := map[string]*tektonv1.PipelineRunTaskRunStatus{}
	if analyzePr != nil {
		for taskRunName, status := range analyzePr.Status.TaskRuns {
			taskRuns[taskRunName] = status
		}
	}
	if cicdPr != nil {
		for taskRunName, status := range cicdPr.Status.TaskRuns {
			taskRuns[taskRunName] = status
		}
	}

	// Update each phase status
	failOccurred := false
	allSucceeded := true

	failedPhase := l2cv1.PhaseNull
	currentRunning := l2cv1.PhaseNull

	currentMessage := ""

	for i, status := range l2crun.Status.Conditions {
		phase := status.Type
		taskRunName := string(phase)
		oldStatus := status.Status
		oldMessage := status.Message

		// Find taskrun
		var tr *tektonv1.PipelineRunTaskRunStatus
		for _, _tr := range taskRuns {
			if _tr.PipelineTaskName == taskRunName {
				tr = _tr
			}
		}
		// If TaskRun found --> do jobs
		if tr != nil && len(tr.Status.Conditions) > 0 {
			trCond := tr.Status.Conditions[0]

			// Set Message same as taskrun message
			l2crun.Status.Conditions[i].Message = trCond.Message

			switch trCond.Status {
			// Status: True --> succeeded
			case corev1.ConditionTrue:
				l2crun.Status.Conditions[i].Status = l2cv1.StatusSucceeded
				break
			// Status: False --> failed
			case corev1.ConditionFalse:
				l2crun.Status.Conditions[i].Status = l2cv1.StatusFailed
				failedPhase = phase
				failOccurred = true
				allSucceeded = false
				break
			// Status: Unknown --> running/pending - see Reason
			case corev1.ConditionUnknown:
				if trCond.Reason == "Running" {
					l2crun.Status.Conditions[i].Status = l2cv1.StatusRunning
				} else {
					l2crun.Status.Conditions[i].Status = l2cv1.StatusPending
					l2crun.Status.Conditions[i].Message = trCond.Reason + " : " + trCond.Message
				}
				currentRunning = phase
				currentMessage = l2crun.Status.Conditions[i].Message
				allSucceeded = false
				break
			default:
				reqLogger.Error(errors.New("unknown taskrun status"), "Unknown taskrun status "+string(trCond.Status))
			}

			// Set lastTransitionTime
			if oldStatus != l2crun.Status.Conditions[i].Status || oldMessage != l2crun.Status.Conditions[i].Message {
				time := metav1.Now()
				l2crun.Status.Conditions[i].LastTransitionTime = &time
			}
		} else {
			// No desired taskrun found -> not launched yet
			allSucceeded = false
		}
	}

	// Update overall status
	// failOccurred --> Failed
	if failOccurred {
		l2crun.Status.Phase = failedPhase
		l2crun.Status.Status = l2cv1.StatusFailed
		l2crun.Status.Message = "Phase [" + string(failedPhase) + "] failed"
	}

	// allSucceeded --> Succeeded
	if allSucceeded {
		l2crun.Status.Phase = l2cv1.PhaseNull
		l2crun.Status.Status = l2cv1.StatusSucceeded
		l2crun.Status.Message = "All phase completed successfully!"
		time := metav1.Now()
		l2crun.Status.CompletionTime = &time
	}

	// failOccurred is not true but if one or more PR is failed, set l2crun failed
	if !failOccurred && analyzePr != nil && len(analyzePr.Status.Conditions) > 0 && analyzePr.Status.Conditions[0].Status == corev1.ConditionFalse {
		failOccurred = true
		allSucceeded = false
		l2crun.Status.Phase = l2cv1.PhaseNull
		l2crun.Status.Status = l2cv1.StatusFailed
		l2crun.Status.Message = analyzePr.Status.Conditions[0].Message
	}
	if !failOccurred && cicdPr != nil && len(cicdPr.Status.Conditions) > 0 && cicdPr.Status.Conditions[0].Status == corev1.ConditionFalse {
		failOccurred = true
		allSucceeded = false
		l2crun.Status.Phase = l2cv1.PhaseNull
		l2crun.Status.Status = l2cv1.StatusFailed
		l2crun.Status.Message = cicdPr.Status.Conditions[0].Message
	}

	// Still running ?
	if !failOccurred && !allSucceeded {
		l2crun.Status.Phase = currentRunning
		l2crun.Status.Status = l2cv1.StatusRunning
		l2crun.Status.Message = currentMessage

		// Launch second pr when first one is done but second pr does not exist
		if cicdPr == nil && analyzePr != nil && len(analyzePr.Status.Conditions) > 0 && analyzePr.Status.Conditions[0].Status == corev1.ConditionTrue {
			_, cicdPr = schemes.PipelineRun(l2crun, l2c)
			if err := controllerutil.SetControllerReference(l2crun, cicdPr, r.scheme); err != nil {
				return reconcile.Result{}, err
			}
			if err = utils.CheckAndCreateObject(r.client, types.NamespacedName{Name: cicdPr.Name, Namespace: cicdPr.Namespace}, cicdPr); err != nil {
				if err = r.setStatus(l2crun, l2cv1.PhaseNull, l2cv1.StatusFailed, "Error creating children: "+err.Error()); err != nil {
					return reconcile.Result{}, err
				}
				return reconcile.Result{}, nil
			}
		}
	}

	// Save status!
	if err := r.client.Status().Update(context.TODO(), l2crun); err != nil {
		reqLogger.Error(err, "Unknown error updating L2cRun status")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// Make sure the necessary fields are filled
func initStatusField(cr *l2cv1.L2CRun) bool {
	updated := false

	// Status
	if cr.Status.Status == "" {
		cr.Status.Status = l2cv1.StatusPending
		updated = true
	}

	// Conditions - should be sorted in order!
	phases := append([]l2cv1.Phase(nil), l2cv1.Phases...)
	condUpdated := false
	if len(cr.Status.Conditions) == len(phases) {
		for curIdx, curCond := range cr.Status.Conditions {
			if curCond.Type != phases[curIdx] {
				condUpdated = true
				break
			}
		}
	} else {
		condUpdated = true
	}
	if condUpdated {
		updated = true
		tobeInserted := phases

		// Delete unnecessary fields
		for curIdx, curCond := range cr.Status.Conditions {
			found := -1
			for desIdx, desCond := range tobeInserted {
				if curCond.Type == desCond {
					found = desIdx
					tobeInserted = append(tobeInserted[:desIdx], tobeInserted[desIdx+1:]...)
					desIdx = desIdx - 1
				}
			}
			if found < 0 {
				cr.Status.Conditions = append(cr.Status.Conditions[:curIdx], cr.Status.Conditions[curIdx+1:]...)
				curIdx = curIdx - 1
			}
		}

		// Fill necessary fields
		for _, desCond := range tobeInserted {
			updated = true

			cr.Status.Conditions = append(cr.Status.Conditions, l2cv1.L2cRunSHCondition{
				Type:   desCond,
				Status: l2cv1.StatusPending,
			})
		}

		// Sort in order TODO
	}

	return updated
}

func (r *ReconcileL2CRun) setStatus(cr *l2cv1.L2CRun, phase l2cv1.Phase, status l2cv1.Status, message string) error {
	reqLogger := log.WithValues("Request.Namespace", cr.Namespace, "Request.Name", cr.Name)
	cr.Status.Phase = phase
	cr.Status.Status = status
	cr.Status.Message = message
	if err := r.client.Status().Update(context.TODO(), cr); err != nil {
		reqLogger.Error(err, "Unknown error updating L2cRun status")
		return err
	}
	return nil
}
