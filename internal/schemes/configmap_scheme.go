package schemes

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-yaml/yaml"

	"tmax.io/l2c-operator/internal/apis/apps"
	"tmax.io/l2c-operator/internal/apis/tmax/template"
	"tmax.io/l2c-operator/internal/utils"
	l2cv1 "tmax.io/l2c-operator/pkg/apis/tmax/v1"
)

func ConfigMap(cr *l2cv1.L2C) (*corev1.ConfigMap, error) {
	// Make DB TemplateInstance as YAML
	dbTi := dbTemplateInstance(cr)
	dbTiYaml, err := yaml.Marshal(&dbTi)
	if err != nil {
		return nil, err
	}

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
			"db-ti.yaml":       string(dbTiYaml),
			"deploy-spec.yaml": string(depSpecYaml),
		},
	}, nil
}

func dbTemplateInstance(cr *l2cv1.L2C) *template.TemplateInstance {
	return &template.TemplateInstance{
		APIVersion: "tmax.io/v1",
		Kind:       "TemplateInstance",
		Metadata: template.Metadata{
			Name:      utils.GetDbTemplateInstanceName(cr),
			Namespace: cr.Namespace,
			Labels: map[string]string{
				"sonar-project-id": cr.Spec.ProjectName,
			},
		},
		Spec: template.TemplateInstanceSpec{
			Template: template.TemplateInstanceSpecTemplate{
				Metadata: template.TemplateInstanceSpecParamMetadata{
					Name: utils.GetDbTemplateName(cr),
				},
				Parameters: []template.TemplateInstanceParam{
					{Name: "APP_NAME", Value: utils.GetDbAppName(cr)},
					{Name: "NAMESPACE", Value: cr.Namespace},
					{Name: "DB_STORAGE", Value: cr.Spec.DbTargetStorageSize},
					{Name: "SERVICE_TYPE", Value: cr.Spec.DbTargetServieceType},
					{Name: "MASTER_USER", Value: cr.Spec.DbTargetUser},
					{Name: "MASTER_PASSWORD", Value: cr.Spec.DbTargetPassword},
					{Name: "TCS_INSTALL", Value: "1"},
					{Name: "TCS_SID", Value: cr.Spec.DbTargetUser},
					{Name: "TB_SID", Value: cr.Spec.DbTargetUser},
					{Name: "TCS_PORT", Value: "8629"},
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
