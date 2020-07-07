package vscode

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	tmaxv1alpha1 "tmax.io/l2c-operator/pkg/apis/tmax/v1alpha1"
)

// labelsForMemcached returns the labels for selecting the resources
// belonging to the given memcached CR name.
func labels(name string) map[string]string {
	return map[string]string{"app": "l2c", "VSCode_cr": name}
}

// deployment returns a memcached Deployment object
func (r *ReconcileVSCode) NewDepMeta(cr *tmaxv1alpha1.VSCode) *appsv1.Deployment {
	ls := labels(cr.Name)

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
									Value: cr.Spec.ProjectName,
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
	// Set VSCode instance as the owner and controller
	controllerutil.SetControllerReference(cr, dep, r.scheme)
	return dep
}

func (r *ReconcileVSCode) NewServiceMeta(cr *tmaxv1alpha1.VSCode) *corev1.Service {
	labels := labels(cr.Name)

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
	// Set VSCode instance as the owner and controller
	controllerutil.SetControllerReference(cr, svc, r.scheme)
	return svc
}

func (r *ReconcileVSCode) NewConfigmapMeta(cr *tmaxv1alpha1.VSCode, token string) *corev1.ConfigMap {
	labels := labels(cr.Name)

	svc := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		// XXX: sonarqube's endpoint address is set manually. (Fix it to be injected)
		Data: map[string]string{
			"settings.json": fmt.Sprintf("{\n    \"sonarlint.connectedMode.connections.sonarqube\": [\n        {\n            \"serverUrl\": \"http://%s\",\n            \"token\": \"%s\"\n         }\n    ],\n    \"sonarlint.connectedMode.project\": {\n        \"projectKey\": \"%s\"\n    },\n    \"java.semanticHighlighting.enabled\": true,\n    \"sonarlint.ls.javaHome\": \"/usr/lib/jvm/java-11-openjdk-amd64\",\n    \"java.home\": \"/usr/lib/jvm/java-11-openjdk-amd64\"\n}\n", cr.Name, token, cr.Spec.ProjectName),
		},
	}
	// Set VSCode instance as the owner and controller
	controllerutil.SetControllerReference(cr, svc, r.scheme)
	return svc
}

func (r *ReconcileVSCode) NewSecretMeta(cr *tmaxv1alpha1.VSCode) *corev1.Secret {
	labels := labels(cr.Name)

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
	// Set VSCode instance as the owner and controller
	controllerutil.SetControllerReference(cr, secret, r.scheme)
	return secret
}
