package template

type TemplateInstance struct {
	APIVersion string               `json:"apiVersion" yaml:"apiVersion"`
	Kind       string               `json:"kind" yaml:"kind"`
	Metadata   Metadata             `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Spec       TemplateInstanceSpec `json:"spec" yaml:"spec"`
}

type Metadata struct {
	Name      string            `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Labels    map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

type TemplateInstanceSpec struct {
	Template TemplateInstanceSpecTemplate `json:"template" yaml:"template"`
}

type TemplateInstanceSpecTemplate struct {
	Metadata   TemplateInstanceSpecParamMetadata `json:"metadata" yaml:"metadata"`
	Parameters []TemplateInstanceParam           `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

type TemplateInstanceSpecParamMetadata struct {
	Name string `json:"name" yaml:"name"`
}

type TemplateInstanceParam struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}
