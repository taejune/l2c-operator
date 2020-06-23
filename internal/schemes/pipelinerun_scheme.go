package schemes

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"tmax.io/l2c-operator/internal/utils"
	l2cv1 "tmax.io/l2c-operator/pkg/apis/tmax/v1"
)

func PipelineRun(l2cRun *l2cv1.L2CRun, l2c *l2cv1.L2C) (analyze *tektonv1.PipelineRun, cicd *tektonv1.PipelineRun) {
	analyzeName, cicdName := utils.GetPipelineRunName(l2cRun)
	analyzeP, cicdP := utils.GetPipelineName(l2c)
	gitResName, imgResName := utils.GetPipelineResourceName(l2c)
	sonarUrl, sonarToken := utils.GetSonarServerAccessInfo()
	doDbMigrate := "TRUE"
	if !l2c.Spec.DbMigrate {
		doDbMigrate = "FALSE"
	}
	return &tektonv1.PipelineRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      analyzeName,
				Namespace: l2c.Namespace,
				Labels:    utils.GetL2cRunLabel(l2cRun, l2c, "analyze-migrate"),
			},
			Spec: tektonv1.PipelineRunSpec{
				ServiceAccountName: utils.GetServiceAccountName(l2c),
				PipelineRef: &tektonv1.PipelineRef{
					Name: analyzeP,
				},
				Resources: []tektonv1.PipelineResourceBinding{
					{
						Name:        "source",
						ResourceRef: &tektonv1.PipelineResourceRef{Name: gitResName},
					},
				},
				Params: []tektonv1.Param{
					{Name: "L2C_NAME", Value: tektonv1.ArrayOrString{StringVal: l2c.Name, Type: tektonv1.ParamTypeString}},
					{Name: "SONAR_URL", Value: tektonv1.ArrayOrString{StringVal: sonarUrl, Type: tektonv1.ParamTypeString}},
					{Name: "SONAR_TOKEN", Value: tektonv1.ArrayOrString{StringVal: sonarToken, Type: tektonv1.ParamTypeString}},
					{Name: "SONAR_PROJECT_ID", Value: tektonv1.ArrayOrString{StringVal: l2c.Spec.ProjectName, Type: tektonv1.ParamTypeString}},
					{Name: "DB_MIGRATE", Value: tektonv1.ArrayOrString{StringVal: doDbMigrate, Type: tektonv1.ParamTypeString}},
					{Name: "DB_FROM", Value: tektonv1.ArrayOrString{StringVal: l2c.Spec.DbSourceType, Type: tektonv1.ParamTypeString}},
					{Name: "DB_FROM_IP", Value: tektonv1.ArrayOrString{StringVal: l2c.Spec.DbSourceHost, Type: tektonv1.ParamTypeString}},
					{Name: "DB_FROM_PORT", Value: tektonv1.ArrayOrString{StringVal: fmt.Sprint(l2c.Spec.DbSourcePort), Type: tektonv1.ParamTypeString}},
					{Name: "DB_FROM_USER", Value: tektonv1.ArrayOrString{StringVal: l2c.Spec.DbSourceUser, Type: tektonv1.ParamTypeString}},
					{Name: "DB_FROM_PASSWORD", Value: tektonv1.ArrayOrString{StringVal: l2c.Spec.DbSourcePassword, Type: tektonv1.ParamTypeString}},
					{Name: "DB_FROM_SID", Value: tektonv1.ArrayOrString{StringVal: l2c.Spec.DbSourceSid, Type: tektonv1.ParamTypeString}},
					{Name: "DB_TO", Value: tektonv1.ArrayOrString{StringVal: l2c.Spec.DbTargetType, Type: tektonv1.ParamTypeString}},
					{Name: "DB_TO_USER", Value: tektonv1.ArrayOrString{StringVal: l2c.Spec.DbTargetUser, Type: tektonv1.ParamTypeString}},
					{Name: "DB_TO_PASSWORD", Value: tektonv1.ArrayOrString{StringVal: l2c.Spec.DbTargetPassword, Type: tektonv1.ParamTypeString}},
				},
			},
		}, &tektonv1.PipelineRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cicdName,
				Namespace: l2c.Namespace,
				Labels:    utils.GetL2cRunLabel(l2cRun, l2c, "ci-cd"),
			},
			Spec: tektonv1.PipelineRunSpec{
				ServiceAccountName: utils.GetServiceAccountName(l2c),
				PipelineRef: &tektonv1.PipelineRef{
					Name: cicdP,
				},
				Resources: []tektonv1.PipelineResourceBinding{
					{Name: "source-repo", ResourceRef: &tektonv1.PipelineResourceRef{Name: gitResName}},
					{Name: "image", ResourceRef: &tektonv1.PipelineResourceRef{Name: imgResName}},
				},
				Params: []tektonv1.Param{
					{Name: "app-name", Value: tektonv1.ArrayOrString{StringVal: l2c.Name, Type: tektonv1.ParamTypeString}},
					{Name: "deploy-cfg-name", Value: tektonv1.ArrayOrString{StringVal: utils.GetConfigMapName(l2c), Type: tektonv1.ParamTypeString}},
					{Name: "deploy-env-json", Value: tektonv1.ArrayOrString{StringVal: "", Type: tektonv1.ParamTypeString}}, //TODO: DB access info
				},
			},
		}
}
