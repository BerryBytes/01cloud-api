package models

type EnvironmentState struct {
	Namespace string
	PodName   string
	Status    string
	ReleaseId string
	Label     string
	LabelKey  string
}
