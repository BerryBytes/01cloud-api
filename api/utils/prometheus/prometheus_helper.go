package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"

	"github.com/jinzhu/gorm"

	"strconv"
	"time"

	"01cloud-api/api/models"
	"01cloud-api/api/utils/helper"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	log "github.com/sirupsen/logrus"
)

func CretePrometheusClient(url string) (v1.API, context.Context, func()) {
	client, err := api.NewClient(api.Config{
		Address: url,
		RoundTripper: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 10 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	})
	if err != nil {
		log.Errorf("Error creating client: %v\n", err)
		return nil, nil, func() {}
	}
	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	return v1api, ctx, cancel
}

func GetAvailableSizeInPVC(cluster *models.Cluster, namespace string) map[string]float64 {
	v1api, ctx, cancel := CretePrometheusClient(cluster.PrometheusServerUrl)
	defer cancel()
	availableFs := fmt.Sprintf(`sum without(instance, node) (kubelet_volume_stats_capacity_bytes{namespace="%s"}) - sum without(instance, node) (kubelet_volume_stats_available_bytes{namespace="%s"})`, namespace, namespace)
	availableSpace, _, err := v1api.Query(ctx, availableFs, time.Now())
	if err != nil {
		log.Error(err)
	}
	response := map[string]float64{}
	js, err := json.Marshal(availableSpace)
	if err != nil {
		log.Error(err)
		return response
	}
	var mp []map[string]interface{}
	err = json.Unmarshal(js, &mp)
	if err != nil || len(mp) == 0 {
		return response
	}
	for _, m := range mp {
		value := m["value"].([]interface{})[1]
		vl, err := strconv.ParseFloat(value.(string), 32)
		if err == nil {
			response[m["metric"].(map[string]interface{})["persistentvolumeclaim"].(string)] = vl / (1024 * 1024)
		}
	}
	return response
}

func GetAvailablePVInCluster(cluster *models.Cluster) float64 {
	v1api, ctx, cancel := CretePrometheusClient(cluster.PrometheusServerUrl)
	defer cancel()
	availableFs := `sum (node_filesystem_size_bytes) by (capacity)-  sum ( sum (node_filesystem_size_bytes) by (capacity) - sum (node_filesystem_avail_bytes) by (capacity)  ) by (capacity)`
	availableSpace, _, err := v1api.Query(ctx, availableFs, time.Now())
	if err != nil {
		log.Error(err)
	}
	js, err := json.Marshal(availableSpace)
	if err != nil {
		log.Error(err)
		return 0
	}
	var mp []map[string]interface{}
	err = json.Unmarshal(js, &mp)
	if err != nil || len(mp) == 0 {
		return 0
	}
	value := mp[0]["value"].([]interface{})[1]
	vl, err := strconv.ParseFloat(value.(string), 32)
	if err != nil {
		log.Error(err)
		return 0
	}
	return vl / (1024 * 1024)
}

func GetAvailableRamInCluster(cluster *models.Cluster) float64 {
	v1api, ctx, cancel := CretePrometheusClient(cluster.PrometheusServerUrl)
	defer cancel()
	availableFs := `sum(kube_node_status_capacity_memory_bytes) by (cluster) - sum(container_memory_usage_bytes{id="/"})  by (cluster)`
	availableSpace, _, err := v1api.Query(ctx, availableFs, time.Now())
	if err != nil {
		log.Error(err)
		return 0
	}
	js, err := json.Marshal(availableSpace)
	if err != nil {
		log.Error(err)
		return 0
	}
	var mp []map[string]interface{}
	err = json.Unmarshal(js, &mp)
	if err != nil || len(mp) == 0 {
		return 0
	}
	value := mp[0]["value"].([]interface{})[1]
	vl, err := strconv.ParseFloat(value.(string), 32)
	if err != nil {
		log.Error(err)
		return 0
	}
	if vl < 0 {
		return 0
	}
	return vl / (1024 * 1024)
}

func GetNumberOfePvcInCluster(cluster *models.Cluster) int64 {
	v1api, ctx, cancel := CretePrometheusClient(cluster.PrometheusServerUrl)
	defer cancel()
	availablePv := `sum (kube_persistentvolume_info) by (cluster)`
	avalableNoOfPv, _, err := v1api.Query(ctx, availablePv, time.Now())
	if err != nil {
		log.Error(err)
		return 0
	}
	js, err := json.Marshal(avalableNoOfPv)
	if err != nil {
		return 0
	}
	var mp []map[string]interface{}
	err = json.Unmarshal(js, &mp)
	if err != nil || len(mp) == 0 {
		return 0
	}
	value := mp[0]["value"].([]interface{})[1]
	vl, err := strconv.ParseInt(value.(string), 10, 64)
	if err != nil {
		log.Error(err)
		return 0
	}
	return vl
}

func ValidateClusterResource(db *gorm.DB, environment *models.Environment) bool {
	cluster := environment.Application.Cluster
	availableRam := GetAvailableRamInCluster(cluster)
	availableCpu := GetAvailableCoreInCluster(cluster)
	//availablePvc := GetAvailablePVInCluster(cluster)

	usage, _ := environment.GetUsedResource(db, environment.Application.ClusterID, true)
	log.Debugf("Memory used in cluster %d is %d", cluster.ID, usage.Memory)
	remaining := float64(cluster.TotalMemory)*cluster.ProvisionPercentage - float64(usage.Memory)
	log.Debugf("Remaining Memory in cluster %d is %f", cluster.ID, remaining)
	replicas := environment.Replicas
	if environment.Resource != nil {
		if availableCpu > 0 && availableRam > 0 {
			if (uint64(replicas)*2*environment.Resource.Memory) > uint64(availableRam) ||
				(uint64(replicas)*2*environment.Resource.Cores) > uint64(availableCpu) {
				return false
			}
		}
		if float64(uint64(replicas)*environment.Resource.Memory) > remaining {
			return false
		}
	}
	pvcs := GetNumberOfePvcInCluster(cluster)
	if pvcs > 0 && uint64(pvcs) > cluster.PvCapacity {
		return false
	}
	return true
}

func GetAvailableCoreInCluster(cluster *models.Cluster) float64 {
	v1api, ctx, cancel := CretePrometheusClient(cluster.PrometheusServerUrl)
	defer cancel()
	availableCore := `sum (machine_cpu_cores) by (cluster) - sum (rate (container_cpu_usage_seconds_total{id="/"}[1m])) by (cluster)`
	avCore, _, err := v1api.Query(ctx, availableCore, time.Now())
	if err != nil {
		log.Error(err)
	}
	js, err := json.Marshal(avCore)
	if err != nil {
		return 0
	}
	var mp []map[string]interface{}
	err = json.Unmarshal(js, &mp)
	if err != nil || len(mp) == 0 {
		return 0
	}
	value := mp[0]["value"].([]interface{})[1]
	vl, err := strconv.ParseFloat(value.(string), 32)
	if err != nil {
		log.Error(err)
		return 0
	}
	if vl < 0 {
		return 0
	}
	return vl * 1000
}

func GetHPAGraph(insight models.Insight, environment *models.Environment) (*map[string]interface{}, error) {
	var hpaValue models.AutoScaler
	err := json.Unmarshal(environment.AutoScaler.RawMessage, &hpaValue)
	if err != nil || !hpaValue.Enabled {
		return nil, nil
	}
	v1api, ctx, cancel := CretePrometheusClient(environment.Application.Cluster.PrometheusServerUrl)
	defer cancel()
	namespace := insight.Namespace
	difference := insight.End - insight.Start
	var step = time.Minute * 60

	if difference <= (60 * 60) {
		step = time.Minute
	}

	r := v1.Range{
		Start: insight.StartTime(),
		End:   insight.EndTime(),
		Step:  step,
	}
	currentReplicas := fmt.Sprintf(`kube_horizontalpodautoscaler_status_current_replicas{horizontalpodautoscaler="%s",namespace="%s"}`, helper.GetReleaseName(environment), namespace)
	crResult, _, err := v1api.QueryRange(ctx, currentReplicas, r)
	if err != nil {
		log.Error(err)
	}
	desiredReplicas := fmt.Sprintf(`kube_horizontalpodautoscaler_status_desired_replicas{horizontalpodautoscaler="%s",namespace="%s"}`, helper.GetReleaseName(environment), namespace)
	drResult, _, err := v1api.QueryRange(ctx, desiredReplicas, r)
	if err != nil {
		log.Error(err)
	}
	minReplicas := fmt.Sprintf(`kube_horizontalpodautoscaler_spec_min_replicas{horizontalpodautoscaler="%s",namespace="%s"}`, helper.GetReleaseName(environment), namespace)
	mnResult, _, err := v1api.QueryRange(ctx, minReplicas, r)
	if err != nil {
		log.Error(err)
	}
	maxReplicas := fmt.Sprintf(`kube_horizontalpodautoscaler_spec_max_replicas{horizontalpodautoscaler="%s",namespace="%s"}`, helper.GetReleaseName(environment), namespace)
	mxResult, _, err := v1api.QueryRange(ctx, maxReplicas, r)
	if err != nil {
		log.Error(err)
	}
	return &map[string]interface{}{
		"current_min_replica": hpaValue.HorizontalPodAutoScaler.MinReplicas,
		"current_max_replica": hpaValue.HorizontalPodAutoScaler.MaxReplicas,
		"min_replica":         mnResult,
		"max_replica":         mxResult,
		"desired":             drResult,
		"running":             crResult,
	}, nil
}

func GetInsight(insight models.Insight, environment *models.Environment) (*map[string]interface{}, error) {
	v1api, ctx, cancel := CretePrometheusClient(environment.Application.Cluster.PrometheusServerUrl)
	defer cancel()
	namespace := insight.Namespace
	difference := insight.End - insight.Start
	var step = time.Minute * 60

	if difference <= (60 * 60) {
		step = time.Minute
	}

	r := v1.Range{
		Start: insight.StartTime(),
		End:   insight.EndTime(),
		Step:  step,
	}
	cpuQuery := fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{namespace="%s"}[%vm])) by (pod)`, namespace, step.Minutes())

	cpuUsages, _, err := v1api.QueryRange(ctx, cpuQuery, r)
	if err != nil {
		log.Error(err)
	}
	memoryUsages, _, _ := v1api.QueryRange(ctx, fmt.Sprintf(`sum(container_memory_working_set_bytes{cluster="", namespace="%s", container!="", image!=""}) by (pod)`, namespace), r)
	receivedBandwidth, _, _ := v1api.QueryRange(ctx, fmt.Sprintf("sum(irate(container_network_receive_bytes_total{namespace=~\"%s\"}[%ds])) by (pod)", namespace, step.Round(time.Second)), r)
	transmitedBandwidth, _, _ := v1api.QueryRange(ctx, fmt.Sprintf("sum(irate(container_network_transmit_bytes_total{namespace=~\"%s\"}[%ds])) by (pod)", namespace, step.Round(time.Second)), r)
	return &map[string]interface{}{
		"total_cpu":     environment.Resource.Cores,
		"total_memory":  environment.Resource.Memory,
		"memory_usages": memoryUsages,
		"cpu_usages":    cpuUsages,
		"data_transfer": &map[string]interface{}{
			"transfer": transmitedBandwidth,
			"receive":  receivedBandwidth,
		},
	}, nil
}

func GetHelmInsight(insight models.Insight, environment *models.HelmEnvironment) (*map[string]interface{}, error) {
	v1api, ctx, cancel := CretePrometheusClient(environment.Application.Cluster.PrometheusServerUrl)
	defer cancel()
	namespace := insight.Namespace
	difference := insight.End - insight.Start
	var step = time.Minute * 60

	if difference <= (60 * 60) {
		step = time.Minute
	}

	r := v1.Range{
		Start: insight.StartTime(),
		End:   insight.EndTime(),
		Step:  step,
	}
	cpuQuery := fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{namespace="%s"}[%vm])) by (pod)`, namespace, step.Minutes())

	cpuUsages, _, err := v1api.QueryRange(ctx, cpuQuery, r)
	if err != nil {
		log.Error(err)
	}
	memoryUsages, _, _ := v1api.QueryRange(ctx, fmt.Sprintf(`sum(container_memory_working_set_bytes{cluster="", namespace="%s", container!="", image!=""}) by (pod)`, namespace), r)
	receivedBandwidth, _, _ := v1api.QueryRange(ctx, fmt.Sprintf("sum(irate(container_network_receive_bytes_total{namespace=~\"%s\"}[%ds])) by (pod)", namespace, step.Round(time.Second)), r)
	transmitedBandwidth, _, _ := v1api.QueryRange(ctx, fmt.Sprintf("sum(irate(container_network_transmit_bytes_total{namespace=~\"%s\"}[%ds])) by (pod)", namespace, step.Round(time.Second)), r)
	return &map[string]interface{}{
		"memory_usages": memoryUsages,
		"cpu_usages":    cpuUsages,
		"data_transfer": &map[string]interface{}{
			"transfer": transmitedBandwidth,
			"receive":  receivedBandwidth,
		},
	}, nil
}

func GetInsightOverview(insight models.Insight, environment *models.Environment) (*map[string]interface{}, error) {
	v1api, ctx, cancel := CretePrometheusClient(environment.Application.Cluster.PrometheusServerUrl)
	defer cancel()
	namespace := insight.Namespace
	difference := insight.End - insight.Start
	step := time.Minute * 60
	//if difference <= (60 * 60 * 24) {
	//	step = time.Minute * 60
	//}
	if difference <= (60 * 60) {
		step = time.Minute * 15
	}

	r := v1.Range{
		Start: insight.StartTime(),
		End:   insight.EndTime(),
		Step:  step,
	}
	cpuQuery := fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{namespace="%s"}[%vm])) by (namespace)`, namespace, step.Minutes())

	cpuUsages, _, err := v1api.QueryRange(ctx, cpuQuery, r)
	if err != nil {
		log.Error(err)
	}
	memoryUsages, _, err := v1api.QueryRange(ctx, fmt.Sprintf(`sum(container_memory_working_set_bytes{cluster="", namespace="%s" , container!="", image!=""}) by (namespace)`, namespace), r)
	if err != nil {
		log.Error(err)
	}
	diskUsages, _, err := v1api.QueryRange(ctx, fmt.Sprintf(`(sum without(instance, node) (kubelet_volume_stats_capacity_bytes{namespace="%s"}) - sum without(instance, node) (kubelet_volume_stats_available_bytes{namespace="%s"}))`, namespace, namespace), r)
	if err != nil {
		log.Error(err)
	}

	receivedBandwidth, _, _ := v1api.QueryRange(ctx, fmt.Sprintf("sum(irate(container_network_receive_bytes_total{namespace=~\"%s\"}[%ds])) by (namespace)", namespace, step.Round(time.Second)), r)
	transmitedBandwidth, _, _ := v1api.QueryRange(ctx, fmt.Sprintf("sum(irate(container_network_transmit_bytes_total{namespace=~\"%s\"}[%ds])) by (namespace)", namespace, step.Round(time.Second)), r)
	var totalPv uint64 = 0
	for _, p := range environment.Storage {
		totalPv = totalPv + p.Capacity
	}
	return &map[string]interface{}{
		"env_name":      environment.Name,
		"total_cpu":     environment.Resource.Cores,
		"total_memory":  environment.Resource.Memory,
		"total_pv":      totalPv,
		"memory_usages": memoryUsages,
		"cpu_usages":    cpuUsages,
		"disk_usages":   diskUsages,
		"data_transfer": map[string]interface{}{
			"transfer": transmitedBandwidth,
			"receive":  receivedBandwidth,
		},
	}, nil
}
func GetHelmInsightOverview(insight models.Insight, environment *models.HelmEnvironment) (*map[string]interface{}, error) {
	v1api, ctx, cancel := CretePrometheusClient(environment.Application.Cluster.PrometheusServerUrl)
	defer cancel()
	namespace := insight.Namespace
	difference := insight.End - insight.Start
	step := time.Minute * 60
	//if difference <= (60 * 60 * 24) {
	//	step = time.Minute * 60
	//}
	if difference <= (60 * 60) {
		step = time.Minute
	}

	r := v1.Range{
		Start: insight.StartTime(),
		End:   insight.EndTime(),
		Step:  step,
	}
	cpuQuery := fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{namespace="%s"}[%vm])) by (namespace)`, namespace, step.Minutes())

	cpuUsages, _, err := v1api.QueryRange(ctx, cpuQuery, r)
	if err != nil {
		log.Error(err)
	}
	memoryUsages, _, err := v1api.QueryRange(ctx, fmt.Sprintf(`sum(container_memory_working_set_bytes{cluster="", namespace="%s" , container!="", image!=""}) by (namespace)`, namespace), r)
	if err != nil {
		log.Error(err)
	}
	diskUsages, _, err := v1api.QueryRange(ctx, fmt.Sprintf(`(sum without(instance, node) (kubelet_volume_stats_capacity_bytes{namespace="%s"}) - sum without(instance, node) (kubelet_volume_stats_available_bytes{namespace="%s"}))`, namespace, namespace), r)
	if err != nil {
		log.Error(err)
	}

	receivedBandwidth, _, _ := v1api.QueryRange(ctx, fmt.Sprintf("sum(irate(container_network_receive_bytes_total{namespace=~\"%s\"}[%ds])) by (namespace)", namespace, step.Round(time.Second)), r)
	transmitedBandwidth, _, _ := v1api.QueryRange(ctx, fmt.Sprintf("sum(irate(container_network_transmit_bytes_total{namespace=~\"%s\"}[%ds])) by (namespace)", namespace, step.Round(time.Second)), r)

	return &map[string]interface{}{
		"memory_usages": memoryUsages,
		"cpu_usages":    cpuUsages,
		"disk_usages":   diskUsages,
		"data_transfer": map[string]interface{}{
			"transfer": transmitedBandwidth,
			"receive":  receivedBandwidth,
		},
	}, nil
}

func GetInsightOverviewCluster(insight models.Insight, cluster *models.Cluster) (*map[string]interface{}, error) {
	v1api, ctx, cancel := CretePrometheusClient(cluster.PrometheusServerUrl)
	defer cancel()
	difference := insight.End - insight.Start
	step := time.Minute * 60
	if difference <= (60 * 60) {
		step = time.Minute
	}

	r := v1.Range{
		Start: insight.StartTime(),
		End:   insight.EndTime(),
		Step:  step,
	}

	cpuUsages, _, err := v1api.QueryRange(ctx, `sum (container_cpu_usage_seconds_total{id="/",kubernetes_io_hostname=~"^.*$"})`, r)
	if err != nil {
		log.Error(err)
	}
	memoryUsages, _, err := v1api.QueryRange(ctx, `sum (container_memory_working_set_bytes{id="/",kubernetes_io_hostname=~"^.*$"})`, r)
	if err != nil {
		log.Error(err)
	}
	diskUsages, _, err := v1api.QueryRange(ctx, `sum (container_fs_usage_bytes{id="/"})`, r)
	if err != nil {
		log.Error(err)
	}

	cpuTotal, _, err := v1api.QueryRange(ctx, `sum (machine_cpu_cores{kubernetes_io_hostname=~"^.*$"})`, r)
	if err != nil {
		log.Error(err)
	}
	memoryTotal, _, err := v1api.QueryRange(ctx, `sum (machine_memory_bytes{kubernetes_io_hostname=~"^.*$"})`, r)
	if err != nil {
		log.Error(err)
	}
	diskTotal, _, err := v1api.QueryRange(ctx, `sum (container_fs_limit_bytes{id="/"})`, r)
	if err != nil {
		log.Error(err)
	}

	nodes, _, err := v1api.QueryRange(ctx, `sum(kube_node_info{node=~".*"})`, r)
	if err != nil {
		log.Error(err)
	}

	pods, _, err := v1api.QueryRange(ctx, `sum(kube_pod_status_phase{namespace=~".*", phase="Running"})`, r)
	if err != nil {
		log.Error(err)
	}

	return &map[string]interface{}{
		"total_cpu":     cpuTotal,
		"total_memory":  memoryTotal,
		"total_disk":    diskTotal,
		"memory_usages": memoryUsages,
		"cpu_usages":    cpuUsages,
		"disk_usages":   diskUsages,
		"nodes_running": nodes,
		"pods_running":  pods,
	}, nil
}

func GetDataTransfer(input *models.Environment) (float64, float64, error) {
	transmitValue, receiveValue := 0.0, 0.0
	if input == nil || input.Application.Cluster == nil {
		return transmitValue, receiveValue, nil
	}
	log.Debug("prometheus server url :: ", input.Application.Cluster.PrometheusServerUrl)
	v1api, ctx, cancel := CretePrometheusClient(input.Application.Cluster.PrometheusServerUrl)
	defer cancel()
	t := time.Now()
	dateend := t
	datestart := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	effectiveTime := int((dateend.Sub(datestart).Hours() + 1) / 24)
	transmitQuery := fmt.Sprintf("sum by (container_label_workload) (rate(container_network_transmit_bytes_total{namespace=~\"%s\"}[%dd]))", helper.GetNamespace(input), effectiveTime)
	receiveQuery := fmt.Sprintf("sum by (container_label_workload) (rate(container_network_receive_bytes_total{namespace=~\"%s\"}[%dd]))", helper.GetNamespace(input), effectiveTime)
	trasmitrate, _, err := v1api.Query(ctx, transmitQuery, time.Now())
	if err != nil {
		return transmitValue, receiveValue, nil
	}
	receiverate, _, err := v1api.Query(ctx, receiveQuery, time.Now())
	if err != nil {
		return transmitValue, receiveValue, nil
	}
	transmitValue, err = ExtractRateValue(trasmitrate)
	if err != nil {
		return transmitValue, receiveValue, err
	}
	receiveValue, err = ExtractRateValue(receiverate)
	if err != nil {
		return transmitValue, receiveValue, err
	}
	log.Debugf("receive rate :: %f transmit rate :: %f", receiveValue, transmitValue)
	return transmitValue, receiveValue, nil
}

func ExtractRateValue(value model.Value) (float64, error) {
	rate := 0.0
	bytes, err := json.Marshal(value)
	if err != nil {
		return rate, err
	}
	result := []map[string]interface{}{}
	err = json.Unmarshal(bytes, &result)
	if err != nil || len(result) == 0 {
		return rate, err
	}
	for _, m := range result {
		value := m["value"].([]interface{})[1]
		rate, err = strconv.ParseFloat(value.(string), 32)
		if err != nil {
			return rate, err
		}
		rate = math.Round(rate*100) / 100
	}
	return rate, nil
}
