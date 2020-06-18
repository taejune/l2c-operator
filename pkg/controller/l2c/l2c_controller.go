package l2c

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"tmax.io/l2c-operator/internal/schemes"
	"tmax.io/l2c-operator/internal/utils"
	l2cv1 "tmax.io/l2c-operator/pkg/apis/tmax/v1"
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

	// ------------------------------------------------------------------------
	svc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: l2c.Name, Namespace: l2c.Namespace}, svc)
	if err != nil && errors.IsNotFound(err) {
		vscodeSvc := r.serviceForL2C(l2c)
		reqLogger.Info("Creating a new Service", "Service.Namespace", vscodeSvc.Namespace, "Service.Name", vscodeSvc.Name)
		err = r.client.Create(context.TODO(), vscodeSvc)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	// ------------------------------------------------------------------------
	configmap := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: l2c.Name, Namespace: l2c.Namespace}, configmap)
	if err != nil && errors.IsNotFound(err) {
		vscodeCm := r.configMapForL2C(l2c)
		reqLogger.Info("Creating a new ConfigMap", "ConfigMap.Namespace", vscodeCm.Namespace, "ConfigMap.Name", vscodeCm.Name)
		err = r.client.Create(context.TODO(), vscodeCm)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	// ------------------------------------------------------------------------
	secret := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: l2c.Name, Namespace: l2c.Namespace}, secret)
	if err != nil && errors.IsNotFound(err) {
		vscodeSecret := r.secretForL2C(l2c)
		reqLogger.Info("Creating a new Secret", "Secret.Namespace", vscodeSecret.Namespace, "Secret.Name", vscodeSecret.Name)
		err = r.client.Create(context.TODO(), vscodeSecret)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	// ------------------------------------------------------------------------
	deployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: l2c.Name, Namespace: l2c.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		vscodeDep := r.deploymentForL2C(l2c)
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", vscodeDep.Namespace, "Deployment.Name", vscodeDep.Name)
		err = r.client.Create(context.TODO(), vscodeDep)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Set status as ready
	if err = r.setStatus(l2c, l2cv1.StatusReady, "All child objects are ready"); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// labelsForMemcached returns the labels for selecting the resources
// belonging to the given memcached CR name.
func labelsForL2C(name string) map[string]string {
	return map[string]string{"app": "l2c", "l2c_cr": name}
}

// deploymentForL2C returns a memcached Deployment object
func (r *ReconcileL2C) deploymentForL2C(cr *l2cv1.L2C) *appsv1.Deployment {
	ls := labelsForL2C(cr.Name)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "vscode",
							Image: "192.168.6.110:5000/tmax/code-server:3.3.1",
							Env: []corev1.EnvVar{
								{
									Name:  "GIT_URL",
									Value: cr.Spec.GitUrl,
								},
								{
									Name:  "PROJECT_NAME",
									Value: "TEST0",
								},
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"/bin/bash",
											"-c",
											"git clone ${GIT_URL} ~/project/${PROJECT_NAME}; cp /tmp/settings.json /home/coder/.local/share/code-server/User/settings.json",
										},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "settings-json-config",
									MountPath: "/tmp/settings.json",
									SubPath:   "settings.json",
								},
								{
									Name:      "config-yaml-secret",
									MountPath: "/home/coder/.config/code-server/config.yaml",
									SubPath:   "config.yaml",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "settings-json-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name,
									},
								},
							},
						},
						{
							Name: "config-yaml-secret",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: cr.Name,
								},
							},
						},
					},
				},
			},
		},
	}
	// Set L2C instance as the owner and controller
	controllerutil.SetControllerReference(cr, dep, r.scheme)
	return dep
}

func (r *ReconcileL2C) serviceForL2C(cr *l2cv1.L2C) *corev1.Service {
	labels := labelsForL2C(cr.Name)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     "LoadBalancer",
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port: 8080,
					TargetPort: intstr.IntOrString{
						IntVal: 8080,
					},
				},
			},
		},
	}
	// Set L2C instance as the owner and controller
	controllerutil.SetControllerReference(cr, svc, r.scheme)
	return svc
}

func (r *ReconcileL2C) configMapForL2C(cr *l2cv1.L2C) *corev1.ConfigMap {
	labels := labelsForL2C(cr.Name)

	svc := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"settings.json": fmt.Sprintf("{\n    \"sonarlint.connectedMode.connections.sonarqube\": [\n        {\n            \"serverUrl\": \"http://l2c-sonar\",\n            \"token\": \"e51f629418eab9c5e205a4caa3714854fff763c1\"\n         }\n    ],\n    \"sonarlint.connectedMode.project\": {\n        \"projectKey\": \"%s\"\n    },\n    \"java.semanticHighlighting.enabled\": true,\n    \"sonarlint.ls.javaHome\": \"/usr/lib/jvm/java-11-openjdk-amd64\",\n    \"java.home\": \"/usr/lib/jvm/java-11-openjdk-amd64\"\n}\n", cr.Spec.ProjectName),
		},
	}
	// Set L2C instance as the owner and controller
	controllerutil.SetControllerReference(cr, svc, r.scheme)
	return svc
}

func (r *ReconcileL2C) secretForL2C(cr *l2cv1.L2C) *corev1.Secret {
	labels := labelsForL2C(cr.Name)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		StringData: map[string]string{
			"config.yaml": fmt.Sprintf("bind-addr: 127.0.0.1:8080\nauth: password\npassword: %s\ncert: false", cr.Spec.AccessCode),
		},
		Type: "Opaque",
	}
	// Set L2C instance as the owner and controller
	controllerutil.SetControllerReference(cr, secret, r.scheme)
	return secret
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
