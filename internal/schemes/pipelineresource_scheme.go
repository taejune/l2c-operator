package schemes

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"tmax.io/l2c-operator/internal/utils"
	l2cv1 "tmax.io/l2c-operator/pkg/apis/tmax/v1"
)

func PipelineResource(cr *l2cv1.L2C) (git *tektonv1.PipelineResource, img *tektonv1.PipelineResource) {
	gitResourceName, imgResourceName := utils.GetPipelineResourceName(cr)
	return &tektonv1.PipelineResource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      gitResourceName,
				Namespace: cr.Namespace,
				Labels:    utils.GetL2cLabel(cr),
			},
			Spec: tektonv1.PipelineResourceSpec{
				Type: tektonv1.PipelineResourceTypeGit,
				Params: []tektonv1.ResourceParam{
					{
						Name:  "url",
						Value: cr.Spec.GitUrl,
					},
					{
						Name:  "revision",
						Value: cr.Spec.GitRevision,
					},
				},
			},
		}, &tektonv1.PipelineResource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      imgResourceName,
				Namespace: cr.Namespace,
				Labels:    utils.GetL2cLabel(cr),
			},
			Spec: tektonv1.PipelineResourceSpec{
				Type: tektonv1.PipelineResourceTypeImage,
				Params: []tektonv1.ResourceParam{
					{
						Name:  "url",
						Value: cr.Spec.ImageUrl,
					},
				},
			},
		}
}
