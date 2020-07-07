package l2c

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	l2cv1 "tmax.io/l2c-operator/pkg/apis/tmax/v1"
	tmaxv1alpha1 "tmax.io/l2c-operator/pkg/apis/tmax/v1alpha1"
)

func (r *ReconcileL2C) vscodeCr(cr *l2cv1.L2C) *tmaxv1alpha1.VSCode {
	label := labelsForL2C(cr.Name)
	vscode := &tmaxv1alpha1.VSCode{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-vscode",
			Namespace: cr.Namespace,
			Labels:    label,
		},
		Spec: tmaxv1alpha1.VSCodeSpec{
			ProjectName: cr.Spec.ProjectName,
			AccessCode:  cr.Spec.AccessCode,
			GitUrl:      cr.Spec.GitUrl,
		},
	}

	controllerutil.SetControllerReference(cr, vscode, r.scheme)
	return vscode
}

func labelsForL2C(name string) map[string]string {
	return map[string]string{"app": "l2c", "l2c": name}
}
