package utils

import (
	"os"
	"strings"

	l2cv1 "tmax.io/l2c-operator/pkg/apis/tmax/v1"
)

func GetConfigMapName(cr *l2cv1.L2C) string {
	return cr.Name + "-cm"
}

func GetServiceName(cr *l2cv1.L2C) string {
	return cr.Name + "-was-svc"
}

func GetServiceAccountName(cr *l2cv1.L2C) string {
	return cr.Name + "-account"
}

func GetRoleName() string {
	return "l2c-role"
}

func GetBindingName(cr *l2cv1.L2C) string {
	return cr.Name + "-binding"
}

func GetPipelineResourceName(cr *l2cv1.L2C) (git string, img string) {
	return cr.Name + "-git", cr.Name + "-img"
}

func GetPipelineName(cr *l2cv1.L2C) (analyze string, cicd string) {
	return cr.Name + "-analyze-migrate", cr.Name + "-cicd"
}

func GetPipelineRunName(cr *l2cv1.L2CRun) (analyze string, cicd string) {
	return cr.Name + "-analyze-migrate", cr.Name + "-cicd"
}

func GetDbAppName(cr *l2cv1.L2C) string {
	return cr.Name + "-db"
}

func GetDbTemplateInstanceName(cr *l2cv1.L2C) string {
	return cr.Name + "-db-instance"
}

func GetDbTemplateName(cr *l2cv1.L2C) string {
	return strings.ToLower(cr.Spec.DbTargetType) + "-template"
}

func GetRegistryUrl() string {
	registry := os.Getenv("REGISTRY_URL")
	if registry == "" {
		registry = "192.168.6.110:5000"
	}
	return registry
}

func GetBuilderImageUrl(cr *l2cv1.L2C) string {
	registry := GetRegistryUrl()
	switch cr.Spec.WasTargetType {
	case "jeus":
		return registry + "/s2i-jeus:8"
	default:
		return "ERR_NOT_SUPPORTED"
	}
}

func GetL2cLabel(l2c *l2cv1.L2C) map[string]string {
	return map[string]string{
		l2cv1.LabelL2cName: l2c.Name,
	}
}

func GetL2cRunLabel(l2cRun *l2cv1.L2CRun, l2c *l2cv1.L2C, phase string) map[string]string {
	label := GetL2cLabel(l2c)
	label[l2cv1.LabelL2cRunName] = l2cRun.Name
	label[l2cv1.LabelL2cRunPhase] = phase
	return label
}
