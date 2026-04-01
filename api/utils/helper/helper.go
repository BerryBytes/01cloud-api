package helper

import (
	"01cloud-api/api/models"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	store "github.com/berrybytes/01cloud-store/model"
	log "github.com/sirupsen/logrus"
)

func GetNamespace(environment *models.Environment) string {
	return fmt.Sprintf("%s-%d-%d", os.Getenv("GCLOUD_NAMESPACE"), environment.ApplicationID, environment.ID)
}
func GetHelmNamespace(environment *models.HelmEnvironment) string {
	return fmt.Sprintf("%s-%d-%d", os.Getenv("GCLOUD_NAMESPACE"), environment.ApplicationID, environment.ID)
}
func GetCiNamespace(environment *models.Environment) string {
	return fmt.Sprintf("%s-%d-%d-ci", os.Getenv("GCLOUD_NAMESPACE"), environment.ApplicationID, environment.ID)
}

func GetOperatorNamespace(operator *models.OperatorRequest) string {
	namespace := "operators"
	if !operator.GlobalOperator {
		namespace = fmt.Sprintf("%s-%d", operator.PackageName, operator.ID)
	}
	return namespace
}

func GetCreateClusterNamespace(request *models.ClusterRequest) string {
	return fmt.Sprintf("%s-cluster-%d-%d", os.Getenv("GCLOUD_NAMESPACE"), request.OrganizationID, request.ID)
}

func GetClusterNamespace(cluster *models.Cluster) string {
	return fmt.Sprintf("%s-%d", os.Getenv("GCLOUD_NAMESPACE"), cluster.ID)
}

func GetContourNamespace(input *models.LoadBalancer) string {
	return fmt.Sprintf("%s-lb-%d", os.Getenv("GCLOUD_NAMESPACE"), input.ID)
}

func GetReleaseName(input *models.Environment) string {
	pName := "docker"
	if input.ServiceType < 2 || input.ServiceType == 5 {
		pName = input.Application.Plugin.Name
	}
	return fmt.Sprintf("%s-metadata-%d-%s", os.Getenv("GCLOUD_NAMESPACE"), input.ID, pName)
}
func GetAddOnRelease(input *models.Environment, addon *models.Environment) string {
	return fmt.Sprintf("%s-addon-%d-%d", os.Getenv("GCLOUD_NAMESPACE"), input.ID, addon.PluginVersion.Plugin.ID)
}

func ReplaceCustomVariables(inputMap map[string]interface{}, outputMap map[string]interface{}) {
	for k, v := range inputMap {
		if reflect.ValueOf(v).String() == "" {
			v = outputMap[k]
		}
		if reflect.TypeOf(v).Kind() == reflect.Map {
			ReplaceCustomVariables(inputMap[k].(map[string]interface{}), outputMap[k].(map[string]interface{}))
		} else {
			outputMap[k] = v
		}
	}
}

func GetDefaultVariable(version *models.PluginVersion) map[string]interface{} {
	schema := fmt.Sprintf("%s/values.schema.json", version.Url)
	if _, err := os.Stat(schema); err != nil {
		log.Error(schema, err)
		return nil
	}
	data, err := os.ReadFile(schema)
	if err != nil {
		log.Error(err)
	}
	inputmap := models.FromJson(data)
	outPutMap := map[string]interface{}{}
	if properties, ok := inputmap["properties"].(map[string]interface{}); ok {
		for s, i := range properties {
			models.GetValuesJson(i.(map[string]interface{}), outPutMap, s)
		}
	}
	return outPutMap
}

func DeleteKey(myMap interface{}, keys ...string) {
	if myMap == nil {
		return
	}
	if mp, ok := myMap.(map[string]interface{}); ok {
		for _, key := range keys {
			if k, ok := mp[key].(string); ok {
				delete(mp, k)
			}
		}
	}
}

func GenerateToken() string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func ConvertWorkflowType(workflow *store.WorkflowMetadata) *models.WorkflowMetadata {
	temp, _ := json.Marshal(workflow)
	newWorkflow := models.WorkflowMetadata{}
	err := json.Unmarshal(temp, &newWorkflow)
	if err != nil {
		log.Error("error from convert workflow type :: ", err)
		return nil
	}
	return &newWorkflow
}

func IsV2(r *http.Request) bool {
	apiVersion := r.Header.Get("X-API-VERSION")
	return apiVersion == "v2" || apiVersion == ""
}

func ToJson(data interface{}) (map[string]interface{}, error) {
	if data == nil {
		return nil, fmt.Errorf("interface value nil")
	}
	msg, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	res := map[string]interface{}{}
	err = json.Unmarshal(msg, &res)
	return res, err
}
func ReadConfigFile(path string) ([]byte, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func ReadAdminConfigFile() (*models.Config, error) {
	file, err := os.ReadFile("/data/admin/setting.json")
	if err != nil {
		return nil, err
	}
	config := models.Config{}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func CheckUserQuota(user *models.User, cat string, count int) error {
	if user.Quotas.RawMessage != nil {
		quotas := &models.Quotas{}
		quotasBytes, _ := json.Marshal(user.Quotas)
		err := json.Unmarshal(quotasBytes, &quotas)
		if err != nil {
			return err
		}
		if cat == "p" && quotas.UserProject <= count {
			return fmt.Errorf("project quota exceeded")
		}
		if cat == "o" && quotas.UserOrganization <= count {
			return fmt.Errorf("organization quota exceeded")
		}
	} else {
		config, err := ReadAdminConfigFile()
		if err != nil {
			return err
		}
		if cat == "p" && config.Quotas.UserProject <= count {
			return fmt.Errorf("project quota exceeded")
		}
		if cat == "o" && config.Quotas.UserOrganization <= count {
			return fmt.Errorf("organization quota exceeded")
		}
	}
	return nil
}

func IsEmptyStruct(s interface{}) bool {
	return reflect.DeepEqual(s, reflect.Zero(reflect.TypeOf(s)).Interface())
}

func GetValues(obj map[string]interface{}, fields ...string) (interface{}, error) {
	var val interface{} = obj
	for i, field := range fields {
		if val == nil {
			return nil, fmt.Errorf("obj is nil")
		}
		if m, ok := val.(map[string]interface{}); ok {
			val, ok = m[field]
			if !ok {
				return nil, fmt.Errorf("%s field not found", field)
			}
		} else {
			return nil, fmt.Errorf("%v error :: %v is of the type %T", "."+strings.Join(fields[:i+1], "."), val, val)
		}
	}
	return val, nil
}

func SetValues(data map[string]interface{}, value interface{}, keys ...string) {
	mp := setNestedValues(value, keys...)
	replaceMap(data, mp)
}

func setNestedValues(value interface{}, fields ...string) map[string]interface{} {
	data := map[string]interface{}{
		fields[len(fields)-1]: value,
	}
	for i := len(fields) - 2; i >= 0; i-- {
		data = map[string]interface{}{
			fields[i]: data,
		}
	}
	return data
}
func replaceMap(original, replacement map[string]interface{}) {
	for key, value := range replacement {
		if originalVal, ok := original[key]; ok {
			if childMap, ok := value.(map[string]interface{}); ok {
				replaceMap(originalVal.(map[string]interface{}), childMap)
			} else {
				original[key] = value
			}
		} else {
			original[key] = value
		}
	}
}

func GetBytes(mp interface{}) []byte {
	bytes, _ := json.Marshal(mp)
	return bytes
}

func ConvertInterfaceToMapArray(data interface{}) []map[string]interface{} {
	arrMap := []map[string]interface{}{}
	switch v := data.(type) {
	case map[string]interface{}:
		arrMap = append(arrMap, v)
	case []interface{}:
		for _, j := range v {
			if mp, ok := j.(map[string]interface{}); ok {
				arrMap = append(arrMap, mp)
			}
		}
	default:
		return arrMap
	}
	return arrMap
}

func SendAlertDeleteRequest(eid uint) {
	requestURL := fmt.Sprintf("%s/monitoring/uninstall-alerts/%d?token=%s", os.Getenv("BASE_URL"), eid, os.Getenv("API_SECRET"))
	request, err := http.NewRequest("DELETE", requestURL, nil)
	if err != nil {
		log.Error("error creating request:", err)
		return
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Error("error sending request ::", err)
		return
	}
	defer resp.Body.Close()
}

func ContainsString(slice []string, str string) bool {
	for _, s := range slice {
		if strings.Contains(s, str) {
			return true
		}
	}
	return false
}

func GetAddonNamespace(input *models.Environment) string {
	namespace := fmt.Sprintf("%s-%d-%d-%s", os.Getenv("GCLOUD_NAMESPACE"), input.Parent.ApplicationID, input.ParentID, input.PluginVersion.Plugin.Name)
	return namespace
}
