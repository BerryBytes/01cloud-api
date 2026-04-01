package models

type AutoScaler struct {
	Enabled                 bool                    `json:"enabled"`
	HorizontalPodAutoScaler HorizontalPodAutoScaler `json:"horizontal_pod_autoscaler"`
	VerticalPodAutoScaler   *VerticalPodAutoScaler  `json:"vertical_pod_autoscaler"`
	AdvancedScheduling      AdvancedScheduling      `json:"advanced_scheduling"`
}
type HorizontalPodAutoScaler struct {
	MinReplicas   int32         `json:"min_replicas"`
	MaxReplicas   int32         `json:"max_replicas"`
	Metrics       []ScaleMetric `json:"metrics"`
	ScaleUpRule   *ScalingRules `json:"scale_up_rule"`
	ScaleDownRule *ScalingRules `json:"scale_down_rule"`
}
type VerticalPodAutoScaler struct {
	UpdateMode          string   `json:"update_mode"`
	Mode                string   `json:"mode"`
	MinAllowedCpu       string   `json:"min_allowed_cpu"`
	MinAllowedMemory    string   `json:"min_allowed_memory"`
	MaxAllowedCpu       string   `json:"max_allowed_cpu"`
	MaxAllowedMemory    string   `json:"max_allowed_memory"`
	ControlledResources []string `json:"controlled_resources"`
	ControlledValues    string   `json:"controlled_values"`
}

type ScalingRules struct {
	AverageTraffic int32  `json:"average_traffic"`
	Type           string `json:"type"`
	Value          int32  `json:"value"`
	TimeInterval   int32  `json:"time_interval"`
}
type AdvancedScheduling struct {
	MaxSkew            int32    `json:"max_skew"`
	MinDomains         int32    `json:"min_domains"`
	TopologyKey        string   `json:"topology_key"`
	WhenUnsatisfiable  string   `json:"when_unsatisfiable"`
	LabelSelector      Label    `json:"label_selector"`
	MatchLabelKeys     []string `json:"match_label_keys"`
	NodeAffinityPolicy string   `json:"node_affinity_policy"`
	NodeTaintPolicy    string   `json:"node_taint_policy"`
}

type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ScaleMetric struct {
	Name               string `json:"name"`
	AverageUtilization int32  `json:"average_utilization"`
}

func (data *AdvancedScheduling) ValidateAdvancedSheduling() {
	if data.MaxSkew == 0 {
		data.MaxSkew = 1
	}
	if data.TopologyKey == "" {
		data.TopologyKey = "kubernetes.io/hostname"
	}
	if data.LabelSelector == (Label{}) {
		data.LabelSelector.Key = "app.kubernetes.io/provider"
		data.LabelSelector.Value = "zerone"
	}
	if data.WhenUnsatisfiable == "" {
		data.WhenUnsatisfiable = "DoNotSchedule"
	}
}
