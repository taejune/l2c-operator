package main

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	errors "k8s.io/apimachinery/pkg/api/errors"

	tmaxv1alpha1 "tmax.io/l2c-operator/pkg/apis/tmax/v1alpha1"
)

const (
	sonarqubeCrName      = "system-sonarqube"
	sonarqubeCrNamespace = "l2c-system"
)

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

func addSonar() {

	cr := &tmaxv1alpha1.Sonarqube{}
	err := cli.Get(context.TODO(), types.NamespacedName{Name: sonarqubeCrName, Namespace: sonarqubeCrNamespace}, cr)
	if err != nil {
		if errors.IsNotFound(err) {

			log.Info("Create new system sonarqube CR")

			cr := &tmaxv1alpha1.Sonarqube{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sonarqubeCrName,
					Namespace: sonarqubeCrNamespace,
				},
			}

			err = cli.Create(context.TODO(), cr)
			if err != nil {
				log.Error(err, "Failed to create system sonarqube CR")
			}

		} else {
			log.Error(err, "Failed to get system sonarqube CR")
			panic(err)
		}
	}
}
