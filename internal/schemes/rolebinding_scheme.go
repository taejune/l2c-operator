package schemes

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"tmax.io/l2c-operator/internal/utils"
	tmaxv1 "tmax.io/l2c-operator/pkg/apis/tmax/v1"
)

func RoleBinding(cr *tmaxv1.L2C) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.GetBindingName(cr),
			Namespace: cr.Namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     utils.GetRoleName(),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      utils.GetServiceAccountName(cr),
				Namespace: cr.Namespace,
			},
		},
	}
}
