package main

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	errors "k8s.io/apimachinery/pkg/api/errors"
	// sonar "k8s.tmax.io/sonarqube/internal/pkg/sonarcalls"
	sonar "github.com/taejune/sonar-client-go/client"
)

const (
	SONAR_SVC_NAME      = "l2c-sonar"
	SONAR_SVC_NAMESPACE = "default"
	SONAR_SVC_PORT      = 9000
)

type SonarqubeCustomResource struct {
	name      string
	namespace string
	ip        string
	hostname  string
	port      int
	token     string
}

var cr = SonarqubeCustomResource{
	name:      SONAR_SVC_NAME,
	namespace: SONAR_SVC_NAMESPACE,
	port:      SONAR_SVC_PORT,
}

var cli, _ = getClient(client.Options{})

func getClient(options client.Options) (client.Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	c, err := client.New(cfg, options)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func CreateSonarqube(ctx context.Context) {

	svc := &corev1.Service{}
	err := cli.Get(ctx, types.NamespacedName{Name: SONAR_SVC_NAME, Namespace: SONAR_SVC_NAMESPACE}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Create sonarqube Service for L2C")
			err := cli.Create(ctx, getSonarService())
			if err != nil {
				panic(err.Error())
			}
		}
		log.Error(err, "Cannot get l2c-sonar Service")
	}

	err = cli.Get(ctx, types.NamespacedName{Name: SONAR_SVC_NAME, Namespace: SONAR_SVC_NAMESPACE}, svc)
	if err != nil {
		log.Error(err, "Cannot get l2c-sonar Service")
		panic(err.Error())
	}

	if len(svc.Status.LoadBalancer.Ingress) < 1 {
		panic("Cannot get service IP")
	}

	addr := fmt.Sprintf("http://%s:%d", svc.Status.LoadBalancer.Ingress[0].IP, SONAR_SVC_PORT)
	sonar.SetConfig(addr, "admin", "admin")

	dep := &appsv1.Deployment{}
	err = cli.Get(context.TODO(), types.NamespacedName{Name: SONAR_SVC_NAME, Namespace: SONAR_SVC_NAMESPACE}, dep)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		sonar := deploymentForSonarqube()
		log.Info("Creating a new Deployment", "Deployment.Namespace", sonar.Namespace, "Deployment.Name", sonar.Name)
		err = cli.Create(context.TODO(), sonar)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", sonar.Namespace, "Deployment.Name", sonar.Name)
			panic(err.Error())
		}
		log.Info("Sonarqube for L2C Deployment created...")
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		panic(err.Error())
	}

	log.Info("Sonarqube is ready")
}

func getSonarService() *corev1.Service {
	labels := labelsForSonar(SONAR_SVC_NAME)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SONAR_SVC_NAME,
			Namespace: SONAR_SVC_NAMESPACE,
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

func deploymentForSonarqube() *appsv1.Deployment {
	ls := labelsForSonar(SONAR_SVC_NAME)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SONAR_SVC_NAME,
			Namespace: SONAR_SVC_NAMESPACE,
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
