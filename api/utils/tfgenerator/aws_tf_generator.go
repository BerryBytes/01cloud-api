package tfgenerator

import (
	"01cloud-api/api/models"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func GenerateMainTfAws(request *models.ClusterRequest) []byte {
	f := hclwrite.NewEmptyFile()
	body := f.Body()
	moduleBody := body.AppendNewBlock("module", []string{"eks"}).Body()
	moduleBody.SetAttributeValue("source", cty.StringVal("github.com/BerryBytes/terraform-aws-eks"))
	moduleBody.SetAttributeTraversal("cluster_name", getTraversalRoute("cluster_name"))
	moduleBody.SetAttributeTraversal("cluster_version", getTraversalRoute("cluster_version"))
	moduleBody.SetAttributeTraversal("subnets", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "aws_subnet",
		},
		hcl.TraverseAttr{
			Name: "zerone-public-subnet[*]",
		},
		hcl.TraverseAttr{
			Name: "id",
		},
	})
	moduleBody.SetAttributeTraversal("vpc_id", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "aws_vpc",
		},
		hcl.TraverseAttr{
			Name: "zerone",
		}, hcl.TraverseAttr{
			Name: "id",
		},
	})
	setAwsNodePools(moduleBody, request)
	return f.Bytes()
}
