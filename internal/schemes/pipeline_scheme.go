package schemes

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"tmax.io/l2c-operator/internal/utils"
	l2cv1 "tmax.io/l2c-operator/pkg/apis/tmax/v1"
)

func Pipeline(cr *l2cv1.L2C) (analyze *tektonv1.Pipeline, cicd *tektonv1.Pipeline) {
	analyzeName, cicdName := utils.GetPipelineName(cr)
	return &tektonv1.Pipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name:      analyzeName,
				Namespace: cr.Namespace,
				Labels:    utils.GetL2cLabel(cr),
			},
			Spec: tektonv1.PipelineSpec{
				Resources: []tektonv1.PipelineDeclaredResource{
					{Name: "source", Type: tektonv1.PipelineResourceTypeGit},
				},
				Params: []tektonv1.ParamSpec{
					{Name: "L2C_NAME", Description: "L2c Name", Type: tektonv1.ParamTypeString},
					{Name: "SONAR_URL", Description: "Sonar Qube server URL", Type: tektonv1.ParamTypeString},
					{Name: "SONAR_TOKEN", Description: "Token for sonar qube", Type: tektonv1.ParamTypeString},
					{Name: "SONAR_PROJECT_ID", Description: "Project ID in sonar qube", Type: tektonv1.ParamTypeString},
					{Name: "DB_MIGRATE", Description: "Whether or not to migrate database (TRUE/FALSE)", Type: tektonv1.ParamTypeString},
					{Name: "DB_FROM", Description: "Source DBMS type", Type: tektonv1.ParamTypeString},
					{Name: "DB_FROM_IP", Description: "Source DBMS IP", Type: tektonv1.ParamTypeString},
					{Name: "DB_FROM_PORT", Description: "Source DBMS port", Type: tektonv1.ParamTypeString},
					{Name: "DB_FROM_USER", Description: "Source DBMS user", Type: tektonv1.ParamTypeString},
					{Name: "DB_FROM_PASSWORD", Description: "Source DBMS password", Type: tektonv1.ParamTypeString},
					{Name: "DB_FROM_SID", Description: "Source DBMS sid", Type: tektonv1.ParamTypeString},
					{Name: "DB_TO", Description: "Target DBMS type", Type: tektonv1.ParamTypeString},
					{Name: "DB_TO_USER", Description: "Target DBMS user", Type: tektonv1.ParamTypeString},
					{Name: "DB_TO_PASSWORD", Description: "Target DBMS password", Type: tektonv1.ParamTypeString},
				},
				Tasks: []tektonv1.PipelineTask{
					{
						Name: string(l2cv1.PhaseAnalyze),
						TaskRef: &tektonv1.TaskRef{
							Name: l2cv1.TaskAnalyzeJavaMaven, // TODO: Change according to java/maven
							Kind: tektonv1.ClusterTaskKind,
						},
						Resources: &tektonv1.PipelineTaskResources{
							Inputs: []tektonv1.PipelineTaskInputResource{
								{Name: "source", Resource: "source"},
							},
						},
						Params: []tektonv1.Param{
							{Name: "SONAR_URL", Value: tektonv1.ArrayOrString{StringVal: "$(params.SONAR_URL)", Type: tektonv1.ParamTypeString}},
							{Name: "SONAR_TOKEN", Value: tektonv1.ArrayOrString{StringVal: "$(params.SONAR_TOKEN)", Type: tektonv1.ParamTypeString}},
							{Name: "SONAR_PROJECT_ID", Value: tektonv1.ArrayOrString{StringVal: "$(params.SONAR_PROJECT_ID)", Type: tektonv1.ParamTypeString}},
						},
					},
					{
						Name: string(l2cv1.PhaseDbDeploy),
						TaskRef: &tektonv1.TaskRef{
							Name: l2cv1.TaskDbDeploy, //TODO: Change according to target type
							Kind: tektonv1.ClusterTaskKind,
						},
						Params: []tektonv1.Param{
							{Name: "CM_NAME", Value: tektonv1.ArrayOrString{StringVal: utils.GetConfigMapName(cr), Type: tektonv1.ParamTypeString}},
							{Name: "DB_APP_NAME", Value: tektonv1.ArrayOrString{StringVal: utils.GetDbAppName(cr), Type: tektonv1.ParamTypeString}},
							{Name: "DB_TYPE", Value: tektonv1.ArrayOrString{StringVal: strings.ToUpper(cr.Spec.DbTargetType), Type: tektonv1.ParamTypeString}},
							{Name: "DO_MIGRATE_DB", Value: tektonv1.ArrayOrString{StringVal: "$(params.DB_MIGRATE)", Type: tektonv1.ParamTypeString}},
						},
						RunAfter: []string{string(l2cv1.PhaseAnalyze)},
					},
					{
						Name: string(l2cv1.PhaseDbMigrate),
						TaskRef: &tektonv1.TaskRef{
							Name: l2cv1.TaskDbMigrateTibero, //TODO: Change according to target type
							Kind: tektonv1.ClusterTaskKind,
						},
						Params: []tektonv1.Param{
							{Name: "DO_MIGRATE_DB", Value: tektonv1.ArrayOrString{StringVal: "$(params.DB_MIGRATE)", Type: tektonv1.ParamTypeString}},
							{Name: "s-username", Value: tektonv1.ArrayOrString{StringVal: "$(params.DB_FROM_USER)", Type: tektonv1.ParamTypeString}},
							{Name: "s-password", Value: tektonv1.ArrayOrString{StringVal: "$(params.DB_FROM_PASSWORD)", Type: tektonv1.ParamTypeString}},
							{Name: "s-type", Value: tektonv1.ArrayOrString{StringVal: "$(params.DB_FROM)", Type: tektonv1.ParamTypeString}},
							{Name: "s-sid", Value: tektonv1.ArrayOrString{StringVal: "$(params.DB_FROM_SID)", Type: tektonv1.ParamTypeString}},
							{Name: "s-port", Value: tektonv1.ArrayOrString{StringVal: "$(params.DB_FROM_PORT)", Type: tektonv1.ParamTypeString}},
							{Name: "s-ip", Value: tektonv1.ArrayOrString{StringVal: "$(params.DB_FROM_IP)", Type: tektonv1.ParamTypeString}},
							{Name: "d-username", Value: tektonv1.ArrayOrString{StringVal: "$(params.DB_TO_USER)", Type: tektonv1.ParamTypeString}},
							{Name: "d-password", Value: tektonv1.ArrayOrString{StringVal: "$(params.DB_TO_PASSWORD)", Type: tektonv1.ParamTypeString}},
							{Name: "d-type", Value: tektonv1.ArrayOrString{StringVal: "$(params.DB_TO)", Type: tektonv1.ParamTypeString}},
							{Name: "d-sid", Value: tektonv1.ArrayOrString{StringVal: "$(params.DB_TO_USER)", Type: tektonv1.ParamTypeString}},
							{Name: "d-port", Value: tektonv1.ArrayOrString{StringVal: "8629", Type: tektonv1.ParamTypeString}},
							{Name: "d-ip", Value: tektonv1.ArrayOrString{StringVal: "$(params.L2C_NAME)-db-service", Type: tektonv1.ParamTypeString}},
						},
						RunAfter: []string{string(l2cv1.PhaseDbDeploy)},
					},
				},
			},
		}, &tektonv1.Pipeline{ // Same as Template version 1.0.0
			ObjectMeta: metav1.ObjectMeta{
				Name:      cicdName,
				Namespace: cr.Namespace,
				Labels:    utils.GetL2cLabel(cr),
			},
			Spec: tektonv1.PipelineSpec{
				Resources: []tektonv1.PipelineDeclaredResource{
					{Name: "source-repo", Type: tektonv1.PipelineResourceTypeGit},
					{Name: "image", Type: tektonv1.PipelineResourceTypeImage},
				},
				Params: []tektonv1.ParamSpec{
					{Name: "app-name", Description: "Application name", Type: tektonv1.ParamTypeString},
					{Name: "deploy-cfg-name", Description: "ConfigMap name for deployment", Type: tektonv1.ParamTypeString},
					{Name: "deploy-env-json", Description: "Deployment environment variable in JSON object form", Type: tektonv1.ParamTypeString},
				},
				Tasks: []tektonv1.PipelineTask{
					{
						Name: string(l2cv1.PhaseBuild),
						TaskRef: &tektonv1.TaskRef{
							Name: l2cv1.TaskBuild,
							Kind: tektonv1.ClusterTaskKind,
						},
						Params: []tektonv1.Param{
							{Name: "BUILDER_IMAGE", Value: tektonv1.ArrayOrString{StringVal: utils.GetBuilderImageUrl(cr), Type: tektonv1.ParamTypeString}},
							{Name: "PACKAGE_SERVER_URL", Value: tektonv1.ArrayOrString{StringVal: cr.Spec.WasPackageServer, Type: tektonv1.ParamTypeString}},
							{Name: "REGISTRY_SECRET_NAME", Value: tektonv1.ArrayOrString{StringVal: cr.Spec.ImageRegSecret, Type: tektonv1.ParamTypeString}},
						},
						Resources: &tektonv1.PipelineTaskResources{
							Inputs: []tektonv1.PipelineTaskInputResource{
								{Name: "source", Resource: "source-repo"},
							},
							Outputs: []tektonv1.PipelineTaskOutputResource{
								{Name: "image", Resource: "image"},
							},
						},
					},
					{
						Name: string(l2cv1.PhaseTest),
						TaskRef: &tektonv1.TaskRef{
							Name: l2cv1.TaskTest,
							Kind: tektonv1.ClusterTaskKind,
						},
						Params: []tektonv1.Param{
							{Name: "image-url", Value: tektonv1.ArrayOrString{StringVal: "$(tasks." + string(l2cv1.PhaseBuild) + ".results.image-url)", Type: tektonv1.ParamTypeString}},
						},
						Resources: &tektonv1.PipelineTaskResources{
							Inputs: []tektonv1.PipelineTaskInputResource{
								{Name: "scanned-image", Resource: "image", From: []string{string(l2cv1.PhaseBuild)}},
							},
						},
					},
					{
						Name: string(l2cv1.PhaseDeploy),
						TaskRef: &tektonv1.TaskRef{
							Name: l2cv1.TaskDeploy,
							Kind: tektonv1.ClusterTaskKind,
						},
						Params: []tektonv1.Param{
							{Name: "app-name", Value: tektonv1.ArrayOrString{StringVal: "$(params.app-name)", Type: tektonv1.ParamTypeString}},
							{Name: "image-url", Value: tektonv1.ArrayOrString{StringVal: "$(tasks." + string(l2cv1.PhaseBuild) + ".results.image-url)", Type: tektonv1.ParamTypeString}},
							{Name: "deploy-cfg-name", Value: tektonv1.ArrayOrString{StringVal: "$(params.deploy-cfg-name)", Type: tektonv1.ParamTypeString}},
							{Name: "deploy-env-json", Value: tektonv1.ArrayOrString{StringVal: "$(params.deploy-env-json)", Type: tektonv1.ParamTypeString}},
						},
						Resources: &tektonv1.PipelineTaskResources{
							Inputs: []tektonv1.PipelineTaskInputResource{
								{Name: "image", Resource: "image"},
							},
						},
						RunAfter: []string{string(l2cv1.PhaseTest)},
					},
				},
			},
		}
}
