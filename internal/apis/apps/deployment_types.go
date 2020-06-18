package apps

// Only for generating YAML which is to be stored in ConfigMap

type Deployment struct {
	Spec DeploymentSpec `json:"spec" yaml:"spec"`
}

type DeploymentSpec struct {
	Selector DeploymentSelector `json:"selector,omitempty" yaml:"selector,omitempty"`
	Template PodTemplate        `json:"template" yaml:"template"`
}

type DeploymentSelector struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty" yaml:"matchLabels,omitempty"`
}

type PodTemplate struct {
	Metadata Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Spec     PodSpec  `json:"spec" yaml:"spec"`
}

type Metadata struct {
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

type PodSpec struct {
	Containers       []ContainerSpec `json:"containers" yaml:"containers"`
	ImagePullSecrets []NameObject    `json:"imagePullSecrets,omitempty" yaml:"imagePullSecrets,omitempty"`
}

type ContainerSpec struct {
	Name  string     `json:"name" yaml:"name"`
	Ports []PortSpec `json:"ports,omitempty" yaml:"ports,omitempty"`
}

type PortSpec struct {
	ContainerPort int32 `json:"containerPort" yaml:"containerPort"`
}

type NameObject struct {
	Name string `json:"name" yaml:"name"`
}
