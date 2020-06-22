package schemes

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"tmax.io/l2c-operator/internal/utils"
	l2cv1 "tmax.io/l2c-operator/pkg/apis/tmax/v1"
)

func Service(cr *l2cv1.L2C) *corev1.Service {
	wasServiceName := utils.GetServiceName(cr)
	label := utils.GetL2cLabel(cr)
	label["app"] = cr.Name
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wasServiceName,
			Namespace: cr.Namespace,
			Labels:    label,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceType(cr.Spec.WasServiceType),
			Selector: map[string]string{
				"app":  cr.Name,
				"tier": "was",
			},
			Ports: []corev1.ServicePort{
				{
					Port: cr.Spec.WasPort,
				},
			},
		},
	}
}
