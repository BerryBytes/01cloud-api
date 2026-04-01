package models

import (
	"encoding/json"
	"errors"
	"html"
	"strings"
	"time"

	"01cloud-api/api/utils/constants"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type ClusterRequest struct {
	gorm.Model
	Name                  string         `gorm:"size:255;not null;" json:"cluster_name"`
	Version               string         `gorm:"size:255;null;" json:"cluster_version"`
	Region                string         `gorm:"size:255;not null;" json:"region"`
	Zone                  postgres.Jsonb `json:"zone"`
	Provider              string         `gorm:"size:255;not null;" json:"provider"`
	ProviderName          string         `gorm:"size:255;null;" json:"provider_name"`
	ProjectId             string         `gorm:"size:255;null;" json:"project_id"`
	Credential            string         `json:"credentials"`
	AccessKey             string         `json:"access_key"`
	SecretKey             string         `json:"secret_key"`
	VpcName               string         `gorm:"size:255;null;" json:"vpc_name"`
	SubnetCidrRange       string         `gorm:"size:255;null;" json:"subnet_cidr_range"`
	NetworkCidr           string         `gorm:"size:255;null;" json:"network_cidr"`
	NetworkPolicy         bool           `json:"network_policy"`
	PvcWriteMany          bool           `json:"pvc_write_many"`
	Active                bool           `json:"active"`
	Status                string         `json:"status"`
	Type                  string         `json:"type"`
	TLS                   string         `gorm:"size:255;null;" json:"tls"`
	NfsDetail             postgres.Jsonb `json:"nfs_detail"`
	RegionalCluster       bool           `json:"regional_cluster"`
	RemoveDefaultNodePool bool           `json:"remove_default_node_pool"`
	NodeGroupCount        uint64         `json:"node_group_count"`
	NodeGroupDetail       postgres.Jsonb `json:"node_group_detail"`
	Organization          *Organization  `gorm:"foreignkey:OrganizationID" json:"organization,omitempty"`
	OrganizationID        uint64         `gorm:"default:0" json:"organization_id"`
	SubscriptionID        string         `gorm:"default:0" json:"subscription_id"`
	Cluster               *Cluster       `gorm:"foreignkey:ClusterID" json:"cluster,omitempty"`
	ClusterID             uint64         `gorm:"default:0" json:"cluster_id"`
	ErrorMessage          postgres.Jsonb `json:"error_message"`
}

type VClusterRequest struct {
	Name           string `json:"cluster_name"`
	Region         string `json:"region"`
	Provider       string `json:"provider"`
	SubscriptionID string `gorm:"default:0" json:"subscription_id"`
	VClusterAPIKEY string `json:"vcluster_api_key"`
	Kube_version   string `json:"kube_version"`
}

type ClusterRequestInterface interface {
	Save(db *gorm.DB, data *ClusterRequest) (*ClusterRequest, error)
	FindAllByOrganization(db *gorm.DB, orgId int64) (*[]ClusterRequest, error)
	FindAllWithInactive(db *gorm.DB) (*[]ClusterRequest, error)
	Find(db *gorm.DB, pid uint64) (*ClusterRequest, error)
	FindWithDns(db *gorm.DB, pid uint64) (*ClusterRequest, error)
	Delete(db *gorm.DB, id uint64) (int64, error)
	ChangeStatus(db *gorm.DB, data *ClusterRequest) (*ClusterRequest, error)
	EnableDisableCluster(db *gorm.DB, data *ClusterRequest) (*ClusterRequest, error)
	Update(db *gorm.DB, data *ClusterRequest) (*ClusterRequest, error)
	UpdateErrorMessage(db *gorm.DB, data *ClusterRequest) error
}
type ClusterRequestType struct {
}

func NewClusterRequest() ClusterRequestInterface {
	return &ClusterRequestType{}
}

func (data *ClusterRequest) Validate() error {
	if data.Name == "" {
		return errors.New("required name")
	}
	if data.Provider == "" {
		return errors.New("required provider name")
	}
	if data.Region == "" {
		return errors.New("required region")
	}
	if data.NodeGroupDetail.RawMessage == nil {
		return errors.New("required node group details")
	}
	//if data.Zone.RawMessage == nil {
	//	return errors.New("Required zone")
	//}

	if data.VpcName == "" {
		return errors.New("required vpc name")
	}
	//if data.ProjectId == "" {
	//	return errors.New("Required project id")
	//}
	return nil
}

func (data *VClusterRequest) VClusterValidate() error {
	if data.Name == "" {
		return errors.New("required cluster name")
	}
	if data.Region == "" {
		return errors.New("required region")
	}
	if data.SubscriptionID == "" {
		return errors.New("required subscription plan")
	}
	return nil
}

func (r *VClusterRequest) NewClusterRequest() *ClusterRequest {
	if r.Provider == "" {
		r.Provider = "zerone"
	}
	data := &ClusterRequest{
		Name:           r.Name,
		Region:         r.Region,
		Provider:       r.Provider,
		ProviderName:   r.Provider,
		SubscriptionID: r.SubscriptionID,
	}
	return data
}

func (data *ClusterRequest) ValidatePermission() error {
	if data.Provider == "" {
		return errors.New("required provider Name")
	}
	if data.Provider == constants.GCP && data.Credential == "" {
		return errors.New("required credentials file")
	}
	if data.Provider == constants.EKS && data.AccessKey == "" && data.SecretKey == "" && data.Region == "" {
		return errors.New("required access key and secret key and region")
	}

	return nil
}

func (data *ClusterRequest) Prepare() {
	data.ID = 0
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
	data.Region = html.EscapeString(strings.TrimSpace(data.Region))
	data.Active = true
	data.Status = constants.ClusterStatusDrafted
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
}

func (d *ClusterRequestType) Save(db *gorm.DB, data *ClusterRequest) (*ClusterRequest, error) {
	err = db.Model(&ClusterRequest{}).Create(data).Error
	if err != nil {
		return &ClusterRequest{}, err
	}
	return data, nil
}

func (d *ClusterRequestType) FindAllByOrganization(db *gorm.DB, orgId int64) (*[]ClusterRequest, error) {
	datas := []ClusterRequest{}
	err = db.Model(&ClusterRequest{}).
		Where("organization_id = ?", orgId).
		Order("id desc").
		Preload("Cluster").
		Find(&datas).Error
	if err != nil {
		return &[]ClusterRequest{}, err
	}
	return &datas, nil
}

func (d *ClusterRequestType) FindAllWithInactive(db *gorm.DB) (*[]ClusterRequest, error) {
	datas := []ClusterRequest{}
	err = db.Model(&ClusterRequest{}).Order("id desc").Find(&datas).Error
	if err != nil {
		return &[]ClusterRequest{}, err
	}
	return &datas, nil
}

func (d *ClusterRequestType) Find(db *gorm.DB, pid uint64) (*ClusterRequest, error) {
	data := ClusterRequest{}
	err = db.Model(&ClusterRequest{}).Where("id = ?", pid).Preload("Organization").Preload("Cluster").Take(&data).Error
	if err != nil {
		return &ClusterRequest{}, err
	}
	return &data, nil
}

func (d *ClusterRequestType) FindWithDns(db *gorm.DB, pid uint64) (*ClusterRequest, error) {
	var err error
	data := ClusterRequest{}
	err = db.Model(&ClusterRequest{}).Where("id = ?", pid).Preload("Organization").Preload("Cluster").Preload("Cluster.DNS").Preload("Cluster.Organization").Preload("Cluster.ImageRegistry").Take(&data).Error
	if err != nil {
		return &ClusterRequest{}, err
	}
	return &data, nil
}

func (data *ClusterRequest) ValidateOrganization(db *gorm.DB) (*Organization, error) {
	if data.OrganizationID != 0 {

		organizationInterface := NewOrganization()
		orgn, err := organizationInterface.Find(db, uint(data.OrganizationID))
		if err != nil {
			return nil, errors.New("invalid organization")
		}
		return orgn, nil
	}
	return nil, nil
}

func (d *ClusterRequestType) Delete(db *gorm.DB, id uint64) (int64, error) {
	db = db.Model(&ClusterRequest{}).Where("id = ?", id).Take(&ClusterRequest{}).Delete(&ClusterRequest{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("ClusterRequest not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}

func (d *ClusterRequestType) ChangeStatus(db *gorm.DB, data *ClusterRequest) (*ClusterRequest, error) {
	var err error
	mp := map[string]interface{}{
		"status": data.Status,
	}
	err = db.Model(&ClusterRequest{}).Where("id = ?", data.ID).UpdateColumns(mp).Error
	if err != nil {
		return &ClusterRequest{}, err
	}
	return data, nil
}
func (d *ClusterRequestType) EnableDisableCluster(db *gorm.DB, data *ClusterRequest) (*ClusterRequest, error) {
	var err error
	mp := map[string]interface{}{
		"active": data.Active,
	}
	err = db.Model(&ClusterRequest{}).Where("id = ?", data.ID).UpdateColumns(mp).Error
	if err != nil {
		return &ClusterRequest{}, err
	}
	return data, nil
}

func (d *ClusterRequestType) Update(db *gorm.DB, data *ClusterRequest) (*ClusterRequest, error) {
	var err error
	clusterRequest := map[string]interface{}{
		"pvc_write_many":           data.PvcWriteMany,
		"network_policy":           data.NetworkPolicy,
		"active":                   data.Active,
		"regional_cluster":         data.RegionalCluster,
		"remove_default_node_pool": data.RemoveDefaultNodePool,
	}
	if data.ClusterID != 0 {
		clusterRequest["cluster_id"] = data.ClusterID
	}
	if data.ProjectId != "" {
		clusterRequest["project_id"] = data.ProjectId
	}
	if data.NodeGroupCount != 0 {
		clusterRequest["node_group_count"] = data.NodeGroupCount
	}
	if data.OrganizationID != 0 {
		clusterRequest["organization_id"] = data.OrganizationID
	}
	if data.Name != "" {
		clusterRequest["name"] = data.Name
	}
	if data.Version != "" {
		clusterRequest["version"] = data.Version
	}
	if data.Provider != "" {
		clusterRequest["provider"] = data.Provider
	}
	if data.Region != "" {
		clusterRequest["region"] = data.Region
	}
	if data.Credential != "" {
		clusterRequest["credential"] = data.Credential
	}
	if data.AccessKey != "" {
		clusterRequest["access_key"] = data.AccessKey
	}
	if data.SecretKey != "" {
		clusterRequest["secret_key"] = data.SecretKey
	}
	if data.VpcName != "" {
		clusterRequest["vpc_name"] = data.VpcName
	}
	if data.TLS != "" {
		clusterRequest["tls"] = data.TLS
	}
	if data.Type != "" {
		clusterRequest["type"] = data.Type
	}
	if data.Status != "" {
		clusterRequest["status"] = data.Status
	}
	if data.SubnetCidrRange != "" {
		clusterRequest["subnet_cidr_range"] = data.SubnetCidrRange
	}
	if data.NetworkCidr != "" {
		clusterRequest["network_cidr"] = data.NetworkCidr
	}
	if data.NodeGroupDetail.RawMessage != nil {
		clusterRequest["node_group_detail"] = data.NodeGroupDetail
	}
	if data.Zone.RawMessage != nil {
		clusterRequest["zone"] = data.Zone
	}
	if data.NfsDetail.RawMessage != nil {
		clusterRequest["nfs_detail"] = data.NfsDetail
	}
	err = db.Model(&ClusterRequest{}).Where("id = ?", data.ID).Updates(clusterRequest).Error
	if err != nil {
		return &ClusterRequest{}, err
	}
	return data, nil
}

func (data *ClusterRequestType) UpdateErrorMessage(db *gorm.DB, clusterRequest *ClusterRequest) error {
	return db.Model(&ClusterRequest{}).Where("id = ?", clusterRequest.ID).UpdateColumns(map[string]interface{}{
		"error_message": clusterRequest.ErrorMessage,
	}).Error
}

func (data *ClusterRequest) ToJson() (map[string]interface{}, error) {
	msg, _ := json.Marshal(data)
	res := map[string]interface{}{}
	err = json.Unmarshal(msg, &res)
	return res, err
}
