package tfgenerator

import (
	"errors"
	"fmt"

	"01cloud-api/api/models"
	"01cloud-api/api/utils/constants"

	"github.com/gosimple/slug"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type NodeGroupDetail struct {
	NodeGroupName    string  `json:"node_group_name"`
	InstanceType     string  `json:"instance_type"`
	DiskType         string  `json:"disk_type"`
	DiskSize         string  `json:"disk_size"`
	MinNodeCount     int64   `json:"min_node_count"`
	MaxNodeCount     int64   `json:"max_node_count"`
	InitialNodeCount int64   `json:"initial_node_count"`
	NodeAutoScaling  bool    `json:"node_autoscaling"`
	NodeGroupLabels  []Label `json:"node_group_labels"`
	Preemptible      bool    `json:"preemptible"`
}
type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type NfsDetail struct {
	FirestoreInstanceName string `json:"filestore_instance_name"`
	FirestoreTier         string `json:"filestore_tier"`
	FirestoreZone         string `json:"filestore_zone"`
	FirestoreCapacity     string `json:"filestore_capacity"`
}

func GenerateTFvars(request *models.ClusterRequest) ([]byte, []byte) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	val, err := getZone(request.Zone)
	if err == nil {
		rootBody.SetAttributeValue("zone", cty.ListVal(val))
	}
	rootBody.SetAttributeValue("region", cty.StringVal(request.Region))
	rootBody.SetAttributeValue("project_id", cty.StringVal(request.ProjectId))
	rootBody.SetAttributeValue("cluster_name", cty.StringVal(request.Name))
	rootBody.SetAttributeValue("cluster_version", cty.StringVal(request.Version))
	rootBody.SetAttributeValue("vpc_name", cty.StringVal(request.VpcName))
	rootBody.SetAttributeValue("subnet_cidr_range", cty.StringVal(request.SubnetCidrRange))
	rootBody.SetAttributeValue("network_cidr", cty.StringVal(request.NetworkCidr))
	rootBody.SetAttributeValue("network_policy", cty.BoolVal(request.NetworkPolicy))
	rootBody.SetAttributeValue("pvc-writemany", cty.BoolVal(request.PvcWriteMany))
	if request.Provider == constants.GCP {
		nfsDetail, err := getNfs(request.NfsDetail.RawMessage)
		if err == nil {
			rootBody.SetAttributeValue("filestore_instance_name", cty.StringVal(nfsDetail.FirestoreInstanceName))
			rootBody.SetAttributeValue("filestore_tier", cty.StringVal(nfsDetail.FirestoreTier))
			rootBody.SetAttributeValue("filestore_zone", cty.StringVal(nfsDetail.FirestoreZone))
			rootBody.SetAttributeValue("filestore_capacity", cty.StringVal(nfsDetail.FirestoreCapacity))
		}
	}
	rootBody.SetAttributeValue("regional_cluster", cty.BoolVal(request.RegionalCluster))
	rootBody.SetAttributeValue("remove_default_node_pool", cty.BoolVal(request.RemoveDefaultNodePool))
	rootBody.SetAttributeValue("node_group_count", cty.NumberIntVal(int64(request.NodeGroupCount)))

	nodeGroups, err := getNodeGroupDetail(request.NodeGroupDetail.RawMessage)
	if err == nil {
		for i, v := range nodeGroups {
			rootBody.SetAttributeValue(fmt.Sprintf("%s_%d", "node_group_name", i+1), cty.StringVal(v.NodeGroupName))
			rootBody.SetAttributeValue(fmt.Sprintf("%s_%d", "instance_type", i+1), cty.StringVal(v.InstanceType))
			rootBody.SetAttributeValue(fmt.Sprintf("%s_%d", "disk_type", i+1), cty.StringVal(v.DiskType))
			rootBody.SetAttributeValue(fmt.Sprintf("%s_%d", "disk_size", i+1), cty.StringVal(v.DiskSize))
			rootBody.SetAttributeValue(fmt.Sprintf("%s_%d", "min_node_count", i+1), cty.NumberIntVal(v.MinNodeCount))
			rootBody.SetAttributeValue(fmt.Sprintf("%s_%d", "max_node_count", i+1), cty.NumberIntVal(v.MaxNodeCount))
			rootBody.SetAttributeValue(fmt.Sprintf("%s_%d", "initial_node_count", i+1), cty.NumberIntVal(v.InitialNodeCount))
			rootBody.SetAttributeValue(fmt.Sprintf("%s_%d", "node_autoscaling", i+1), cty.BoolVal(v.NodeAutoScaling))
			rootBody.SetAttributeValue(fmt.Sprintf("%s_%d", "preemptible", i+1), cty.BoolVal(v.Preemptible))
		}
	} else {
		logrus.Error(err)
		return nil, nil
	}
	logrus.Debugf("%s :: ", f.Bytes())
	variables := GenerateVariables(rootBody)

	return f.Bytes(), variables
}

func GenerateVariables(request *hclwrite.Body) []byte {
	f := hclwrite.NewEmptyFile()
	body := f.Body()

	for k := range request.Attributes() {
		body.AppendNewBlock("variable", []string{k})
	}
	ipRange := body.AppendNewBlock("variable", []string{"ip_range_pods"})
	ipRange.Body().SetAttributeValue("description", cty.StringVal("The secondary ip range to use for pods"))
	ipRange.Body().SetAttributeValue("default", cty.StringVal(""))

	ipRangeServices := body.AppendNewBlock("variable", []string{"ip_range_services"})
	ipRangeServices.Body().SetAttributeValue("description", cty.StringVal("The secondary ip range to use for services"))
	ipRangeServices.Body().SetAttributeValue("default", cty.StringVal(""))

	imageType := body.AppendNewBlock("variable", []string{"image_type"})
	imageType.Body().SetAttributeValue("default", cty.StringVal("COS"))
	return f.Bytes()
}

func GenerateBackendTF(request *models.ClusterRequest) []byte {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	block := rootBody.AppendNewBlock("terraform", nil).Body()
	gcsBlock := block.AppendNewBlock("backend", []string{"gcs"}).Body()
	gcsBlock.SetAttributeValue("bucket", cty.StringVal(GetBucketName(request.Organization)))
	gcsBlock.SetAttributeValue("prefix", cty.StringVal(fmt.Sprintf("%s/%s", slug.Make(fmt.Sprintf("%s-%d", request.Name, request.ID)), "terraform/terraform.tfstate")))
	gcsBlock.SetAttributeValue("credentials", cty.StringVal("zerone-devops-lab.json"))
	return f.Bytes()
}

func GetTrafformFolder(request *models.ClusterRequest) string {
	organizationName := "zerone"
	if request.Organization != nil {
		organizationName = slug.Make(request.Organization.Name)
	}
	return fmt.Sprintf("/data/terraform/%s/%s", organizationName, slug.Make(fmt.Sprintf("%s-%d", request.Name, request.ID)))
}
func GetBucketName(org *models.Organization) string {
	organizationName := "zerone"
	orgId := 0
	if org != nil {
		organizationName = org.Name
		orgId = int(org.ID)
	}
	return slug.Make(fmt.Sprintf("%s-%d", organizationName, orgId))
}
func GetObjectName(request *models.ClusterRequest) string {
	return slug.Make(fmt.Sprintf("%s-%d", request.Name, request.ID))
}
func containsNodegroupname(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
func DuplicateNodegroupname(request *models.ClusterRequest) error {
	var nodegroupName []string
	nodeGroups, err := getNodeGroupDetail(request.NodeGroupDetail.RawMessage)
	if err != nil {
		return err
	}
	for _, v := range nodeGroups {
		if containsNodegroupname(nodegroupName, v.NodeGroupName) {
			return errors.New("duplicate node groupname")
		}
		nodegroupName = append(nodegroupName, v.NodeGroupName)
	}
	return nil
}
