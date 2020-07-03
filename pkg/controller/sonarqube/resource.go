package sonarqube

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	tmaxv1alpha1 "tmax.io/l2c-operator/pkg/apis/tmax/v1alpha1"
)

func newServiceForSonar(cr *tmaxv1alpha1.Sonarqube) *corev1.Service {
	labels := labelsForSonar(cr.Name)

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
					Port: 9000,
					TargetPort: intstr.IntOrString{
						IntVal: 9000,
					},
				},
			},
		},
	}

	return svc
}

func newDeploymentForSonar(cr *tmaxv1alpha1.Sonarqube) *appsv1.Deployment {
	ls := labelsForSonar(cr.Name)

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
					Containers: []corev1.Container{{
						Image: "azssi/working:0.0.1",
						Name:  "sonarqube",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 9000,
							Name:          "sonarqube",
						}},
					}},
				},
			},
		},
	}

	return dep
}

func labelsForSonar(name string) map[string]string {
	return map[string]string{"app": "l2c", "sonarqube": name}
}
