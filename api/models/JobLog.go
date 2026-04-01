package models

import "time"

type JobLog struct {
	Log       string
	Status    string
	Owner     string
	Type      string
	PodName   string
	Namespace string
	DateTime  time.Time
}
