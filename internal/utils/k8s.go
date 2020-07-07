package utils

import (
	"context"
	"errors"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
)

// Check first if the object exists
// if not, create one
func CheckAndCreateObject(client client.Client, namespacedName types.NamespacedName, obj interface{}) error {
	resourceType := reflect.TypeOf(obj).String()
	reqLogger := log.Log.WithValues(resourceType+".Namespace", namespacedName.Namespace, resourceType+".Name", namespacedName.Name)

	var typedObj runtime.Object
	switch o := obj.(type) {
	case *corev1.Service:
		typedObj = o
	case *corev1.ConfigMap:
		typedObj = o
	case *corev1.ServiceAccount:
		typedObj = o
	case *rbacv1.RoleBinding:
		typedObj = o
	case *tektonv1.PipelineResource:
		typedObj = o
	case *tektonv1.Pipeline:
		typedObj = o
	case *tektonv1.PipelineRun:
		typedObj = o
	default:
		err := errors.New("Unsupported type " + resourceType)
		reqLogger.Error(err, "Unsupported type attempted to be created")
		return err
	}

	err := client.Get(context.TODO(), namespacedName, typedObj)
	if err != nil && k8serrors.IsNotFound(err) {
		reqLogger.Info("Creating")
		if err = client.Create(context.TODO(), typedObj); err != nil {
			reqLogger.Error(err, "Error creating")
			return err
		}
	} else if err != nil {
		reqLogger.Error(err, "Error getting status")
		return err
	} else {
		reqLogger.Info("Already exists")
	}
	return nil
}
