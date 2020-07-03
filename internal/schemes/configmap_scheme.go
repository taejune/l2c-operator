package schemes

import (
	"bytes"
	"errors"

	"github.com/go-yaml/yaml"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	"tmax.io/l2c-operator/internal/apis/apps"
	"tmax.io/l2c-operator/internal/utils"
	l2cv1 "tmax.io/l2c-operator/pkg/apis/tmax/v1"
)

func ConfigMap(cr *l2cv1.L2C) (*corev1.ConfigMap, error) {
	serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil, json.SerializerOptions{
		Yaml:   true,
		Pretty: true,
		Strict: false,
	})

	// Make DB PVC as YAML
	dbPvc, err := dbPvc(cr)
	if err != nil {
		return nil, err
	}
	dbPvcBuf := new(bytes.Buffer)
	if err := serializer.Encode(dbPvc, dbPvcBuf); err != nil {
		return nil, err
	}
	dbPvcYaml := dbPvcBuf.String()

	// Make DB Svc as YAML
	dbSvc := dbSvc(cr)
	dbSvcBuf := new(bytes.Buffer)
	if err := serializer.Encode(dbSvc, dbSvcBuf); err != nil {
		return nil, err
	}
	dbSvcYaml := dbSvcBuf.String()

	// Make DB Secret as YAML
	dbSecret, err := dbSecret(cr)
	if err != nil {
		return nil, err
	}
	dbSecretBuf := new(bytes.Buffer)
	if err := serializer.Encode(dbSecret, dbSecretBuf); err != nil {
		return nil, err
	}
	dbSecretYaml := dbSecretBuf.String()

	// Make DB Deploy as YAML
	dbDeploy, err := dbDeploy(cr)
	if err != nil {
		return nil, err
	}
	dbDeployBuf := new(bytes.Buffer)
	if err := serializer.Encode(dbDeploy, dbDeployBuf); err != nil {
		return nil, err
	}
	dbDeployYaml := dbDeployBuf.String()

	// Make Deployment Spec as YAML
	depSpec := deploySpec(cr)
	depSpecYaml, err := yaml.Marshal(&depSpec)
	if err != nil {
		return nil, err
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.GetConfigMapName(cr),
			Namespace: cr.Namespace,
			Labels:    utils.GetL2cLabel(cr),
		},
		Data: map[string]string{
			"deploy-spec.yaml": string(depSpecYaml),
			l2cv1.KeyDbPvc:     dbPvcYaml,
			l2cv1.KeyDbSvc:     dbSvcYaml,
			l2cv1.KeyDbSecret:  dbSecretYaml,
			l2cv1.KeyDbDeploy:  dbDeployYaml,
		},
	}, nil
}

func dbPvc(cr *l2cv1.L2C) (*corev1.PersistentVolumeClaim, error) {
	className := "csi-cephfs-sc"
	dbQuant, err := resource.ParseQuantity(cr.Spec.DbTargetStorageSize)
	if err != nil {
		return nil, err
	}
	return &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "PersistentVolumeClaim"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.GetDbPvcName(cr),
			Namespace: cr.Namespace,
			Labels:    utils.GetL2cLabel(cr),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &className,
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: dbQuant},
			},
		},
	}, nil
}

func dbSvc(cr *l2cv1.L2C) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.GetDbSvcName(cr),
			Namespace: cr.Namespace,
			Labels:    utils.GetL2cLabel(cr),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceType(cr.Spec.DbTargetServieceType),
			Ports: []corev1.ServicePort{
				{Port: 8629},
			},
			Selector: map[string]string{
				"app":  cr.Name,
				"tier": cr.Spec.DbTargetType,
			},
		},
	}
}

func dbSecret(cr *l2cv1.L2C) (*corev1.Secret, error) {
	switch cr.Spec.DbTargetType {
	case "TIBERO":
		return dbSecretTibero(cr), nil
	}
	return nil, errors.New("cannot deploy DB type " + cr.Spec.DbTargetType)
}

func dbSecretTibero(cr *l2cv1.L2C) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.GetDbSecretName(cr),
			Namespace: cr.Namespace,
			Labels:    utils.GetL2cLabel(cr),
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"MASTER_USER":     cr.Spec.DbTargetUser,
			"MASTER_PASSWORD": cr.Spec.DbTargetPassword,
			"TCS_INSTALL":     "1",
			"TCS_SID":         cr.Spec.DbTargetUser,
			"TB_SID":          cr.Spec.DbTargetUser,
			"TCS_PORT":        "8629",
		},
	}
}

func dbDeploy(cr *l2cv1.L2C) (*appsv1.Deployment, error) {
	switch cr.Spec.DbTargetType {
	case "TIBERO":
		return dbDeployTibero(cr), nil
	}
	return nil, errors.New("cannot deploy DB type " + cr.Spec.DbTargetType)
}

func dbDeployTibero(cr *l2cv1.L2C) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.GetDbDeployName(cr),
			Namespace: cr.Namespace,
			Labels:    utils.GetL2cLabel(cr),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  cr.Name,
					"tier": cr.Spec.DbTargetType,
				},
			},
			Strategy: appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  cr.Name,
						"tier": cr.Spec.DbTargetType,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "tibero",
							Image: "192.168.6.110:5000/cloud_tcs_tibero_standalone:200309",
							Env: []corev1.EnvVar{
								{
									Name: "MASTER_USER",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: utils.GetDbSecretName(cr)},
											Key:                  "MASTER_USER",
										},
									},
								},
								{
									Name: "MASTER_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: utils.GetDbSecretName(cr)},
											Key:                  "MASTER_PASSWORD",
										},
									},
								},
								{
									Name: "TCS_INSTALL",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: utils.GetDbSecretName(cr)},
											Key:                  "TCS_INSTALL",
										},
									},
								},
								{
									Name: "TCS_SID",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: utils.GetDbSecretName(cr)},
											Key:                  "TCS_SID",
										},
									},
								},
								{
									Name: "TB_SID",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: utils.GetDbSecretName(cr)},
											Key:                  "TB_SID",
										},
									},
								},
								{
									Name: "TCS_PORT",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: utils.GetDbSecretName(cr)},
											Key:                  "TCS_PORT",
										},
									},
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "tibero",
									ContainerPort: 8629,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "tibero-pvc",
									MountPath: "/tibero/mnt/tibero",
								},
							},
							ReadinessProbe: &corev1.Probe{
								InitialDelaySeconds: 5,
								PeriodSeconds:       10,
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"/bin/bash",
											"-c",
											`[ "$(echo 'SELECT COUNT(*) FROM all_tables;' > /tmp/test.sql && echo 'EXIT;' >> /tmp/test.sql && tbsql $MASTER_USER/$MASTER_PASSWORD @/tmp/test.sql | grep -E '[0-9]* row[s]? selected')" == "" ] && exit 1 || exit 0`,
										},
									},
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "tibero-pvc",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: utils.GetDbPvcName(cr),
								},
							},
						},
					},
				},
			},
		},
	}
}

func deploySpec(cr *l2cv1.L2C) *apps.Deployment {
	return &apps.Deployment{
		Spec: apps.DeploymentSpec{
			Selector: apps.DeploymentSelector{
				MatchLabels: map[string]string{
					"app":  cr.Name,
					"tier": "was",
				},
			},
			Template: apps.PodTemplate{
				Metadata: apps.Metadata{
					Labels: map[string]string{
						"app":  cr.Name,
						"tier": "was",
					},
				},
				Spec: apps.PodSpec{
					ImagePullSecrets: []apps.NameObject{
						{Name: cr.Spec.ImageRegSecret},
					},
					Containers: []apps.ContainerSpec{
						{
							Name: "app",
							Ports: []apps.PortSpec{
								{ContainerPort: cr.Spec.WasPort},
							},
						},
					},
				},
			},
		},
	}
}
