package tfgenerator

import (
	"01cloud-api/api/models"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func GenerateMainTfGcp(request *models.ClusterRequest) []byte {
	f := hclwrite.NewEmptyFile()
	body := f.Body()
	body.AppendNewBlock("data", []string{"google_client_config", "dx-gke-vpc"})
	moduleBody := body.AppendNewBlock("module", []string{"gke"}).Body()
	moduleBody.SetAttributeValue("source", cty.StringVal("terraform-google-modules/kubernetes-engine/google"))
	moduleBody.SetAttributeValue("version", cty.StringVal("11.0.0"))
	moduleBody.SetAttributeTraversal("project_id", getTraversalRoute("project_id"))
	moduleBody.SetAttributeTraversal("zones", getTraversalRoute("zone"))
	moduleBody.SetAttributeTraversal("name", getTraversalRoute("cluster_name"))
	moduleBody.SetAttributeTraversal("regional", getTraversalRoute("regional_cluster"))
	moduleBody.SetAttributeTraversal("region", getTraversalRoute("region"))
	moduleBody.SetAttributeTraversal("network", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "google_compute_network",
		},
		hcl.TraverseAttr{
			Name: "dx-gke-vpc",
		},
		hcl.TraverseAttr{
			Name: "name",
		},
	})
	moduleBody.SetAttributeTraversal("subnetwork", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "google_compute_subnetwork",
		},
		hcl.TraverseAttr{
			Name: "public-subnet",
		}, hcl.TraverseAttr{
			Name: "name",
		},
	})
	moduleBody.SetAttributeTraversal("ip_range_pods", getTraversalRoute("ip_range_pods"))
	moduleBody.SetAttributeTraversal("ip_range_services", getTraversalRoute("ip_range_services"))
	moduleBody.SetAttributeValue("create_service_account", cty.BoolVal(false))
	moduleBody.SetAttributeTraversal("network_policy", getTraversalRoute("network_policy"))
	moduleBody.SetAttributeTraversal("kubernetes_version", getTraversalRoute("cluster_version"))
	moduleBody.SetAttributeTraversal("remove_default_node_pool", getTraversalRoute("remove_default_node_pool"))
	setGcpNodePools(moduleBody, request)
	setNodePoolLabels(moduleBody, request)
	return f.Bytes()
}
