package tfgenerator

import (
	"encoding/json"
	"fmt"
	"strings"

	"01cloud-api/api/models"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/zclconf/go-cty/cty"
)

func setGcpNodePools(body *hclwrite.Body, req *models.ClusterRequest) {
	var bodies []string
	for i := 1; i <= int(req.NodeGroupCount); i++ {
		tempFile := hclwrite.NewEmptyFile()
		myBody := tempFile.Body()
		myBody.SetAttributeTraversal("name", getTraversalRoute(fmt.Sprintf("%s_%d", "node_group_name", i)))
		myBody.SetAttributeTraversal("machine_type", getTraversalRoute(fmt.Sprintf("%s_%d", "instance_type", i)))
		myBody.SetAttributeTraversal("min_count", getTraversalRoute(fmt.Sprintf("%s_%d", "min_node_count", i)))
		myBody.SetAttributeTraversal("max_count", getTraversalRoute(fmt.Sprintf("%s_%d", "max_node_count", i)))
		myBody.SetAttributeTraversal("disk_size_gb", getTraversalRoute(fmt.Sprintf("%s_%d", "disk_size", i)))
		myBody.SetAttributeTraversal("disk_type", getTraversalRoute(fmt.Sprintf("%s_%d", "disk_type", i)))
		myBody.SetAttributeTraversal("initial_node_count", getTraversalRoute(fmt.Sprintf("%s_%d", "initial_node_count", i)))
		myBody.SetAttributeTraversal("image_type", getTraversalRoute("image_type"))
		myBody.SetAttributeTraversal("node_autoscaling", getTraversalRoute(fmt.Sprintf("%s_%d", "node_autoscaling", i)))
		myBody.SetAttributeTraversal("preemptible", getTraversalRoute(fmt.Sprintf("%s_%d", "preemptible", i)))
		//fmt.Println(string(tempFile.Bytes()))
		tokens := hclwrite.Tokens{
			{
				Type:         hclsyntax.TokenOBrace,
				Bytes:        []byte("{"),
				SpacesBefore: 0,
			},
			{
				Type:         hclsyntax.TokenNewline,
				Bytes:        []byte("\n"),
				SpacesBefore: 0,
			},
			{
				Type:         hclsyntax.TokenIdent,
				Bytes:        tempFile.Bytes(),
				SpacesBefore: 0,
			},
			{
				Type:         hclsyntax.TokenCBrace,
				Bytes:        []byte("}"),
				SpacesBefore: 0,
			},
		}
		bodies = append(bodies, string(tokens.Bytes()))
	}
	bodyString := strings.Join(bodies, ",\n")
	tokens := hclwrite.Tokens{
		{
			Type:         hclsyntax.TokenOBrack,
			Bytes:        []byte("["),
			SpacesBefore: 0,
		},
		{
			Type:         hclsyntax.TokenNewline,
			Bytes:        []byte("\n"),
			SpacesBefore: 0,
		},
		{
			Type:         hclsyntax.TokenIdent,
			Bytes:        []byte(bodyString),
			SpacesBefore: 1,
		},
		{
			Type:         hclsyntax.TokenIdent,
			Bytes:        []byte("\n"),
			SpacesBefore: 0,
		},
		{
			Type:         hclsyntax.TokenCBrack,
			Bytes:        []byte("]"),
			SpacesBefore: 0,
		},
	}
	body.SetAttributeRaw("node_pools", tokens)
}

func setAwsNodePools(body *hclwrite.Body, request *models.ClusterRequest) {
	var bodies []string
	ngDetail, err := getNodeGroupDetail(request.NodeGroupDetail.RawMessage)
	if err != nil {
		return
	}

	for i, v := range ngDetail {
		mainBody := hclwrite.NewEmptyFile()
		tempFile := hclwrite.NewEmptyFile()
		myBody := tempFile.Body()
		myBody.SetAttributeTraversal("name", getTraversalRoute(fmt.Sprintf("%s_%d", "node_group_name", i+1)))
		myBody.SetAttributeTraversal("min_capacity", getTraversalRoute(fmt.Sprintf("%s_%d", "min_node_count", i+1)))
		myBody.SetAttributeTraversal("max_capacity", getTraversalRoute(fmt.Sprintf("%s_%d", "max_node_count", i+1)))
		myBody.SetAttributeTraversal("disk_size", getTraversalRoute(fmt.Sprintf("%s_%d", "disk_size", i+1)))
		myBody.SetAttributeTraversal("desired_capacity", getTraversalRoute(fmt.Sprintf("%s_%d", "initial_node_count", i+1)))
		myBody.SetAttributeTraversal("node_autoscaling", getTraversalRoute(fmt.Sprintf("%s_%d", "node_autoscaling", i+1)))
		myBody.SetAttributeTraversal("preemptible", getTraversalRoute(fmt.Sprintf("%s_%d", "preemptible", i+1)))
		//fmt.Println(string(tempFile.Bytes()))
		labels := v.NodeGroupLabels

		values := map[string]cty.Value{}
		for _, v := range labels {
			values[v.Key] = cty.StringVal(v.Value)
		}
		myBody.SetAttributeValue("k8s_labels", cty.ObjectVal(values))
		myBody.SetAttributeRaw("instance_types", hclwrite.Tokens{
			{
				Type:         hclsyntax.TokenOBrack,
				Bytes:        []byte("["),
				SpacesBefore: 0,
			},
			{
				Type:         hclsyntax.TokenIdent,
				Bytes:        []byte(fmt.Sprintf("%s_%d", "var.instance_type", i+1)),
				SpacesBefore: 1,
			},
			{
				Type:         hclsyntax.TokenCBrack,
				Bytes:        []byte("]"),
				SpacesBefore: 0,
			},
		})
		tokens := hclwrite.Tokens{
			{
				Type:         hclsyntax.TokenOBrace,
				Bytes:        []byte("{"),
				SpacesBefore: 0,
			},
			{
				Type:         hclsyntax.TokenNewline,
				Bytes:        []byte("\n"),
				SpacesBefore: 0,
			},
			{
				Type:         hclsyntax.TokenIdent,
				Bytes:        tempFile.Bytes(),
				SpacesBefore: 0,
			},
			{
				Type:         hclsyntax.TokenCBrace,
				Bytes:        []byte("}"),
				SpacesBefore: 0,
			},
		}
		mainBody.Body().SetAttributeRaw(v.NodeGroupName, tokens)
		bodies = append(bodies, string(mainBody.Bytes()))
	}
	//fmt.Println(string(bodies))
	bodyString := strings.Join(bodies, "\n")
	tokens := hclwrite.Tokens{
		{
			Type:         hclsyntax.TokenOBrace,
			Bytes:        []byte("{"),
			SpacesBefore: 0,
		},
		{
			Type:         hclsyntax.TokenNewline,
			Bytes:        []byte("\n"),
			SpacesBefore: 0,
		},
		{
			Type:         hclsyntax.TokenIdent,
			Bytes:        []byte(bodyString),
			SpacesBefore: 1,
		},
		{
			Type:         hclsyntax.TokenIdent,
			Bytes:        []byte("\n"),
			SpacesBefore: 0,
		},
		{
			Type:         hclsyntax.TokenCBrace,
			Bytes:        []byte("}"),
			SpacesBefore: 0,
		},
	}
	body.SetAttributeRaw("node_groups", tokens)
}

func setNodePoolLabels(body *hclwrite.Body, request *models.ClusterRequest) {
	labelsMap := map[string]cty.Value{}
	nodeGroups, err := getNodeGroupDetail(request.NodeGroupDetail.RawMessage)
	if err == nil {
		for _, v := range nodeGroups {
			values := map[string]cty.Value{}
			labels := v.NodeGroupLabels
			for _, v := range labels {
				values[v.Key] = cty.StringVal(v.Value)
			}
			labelsMap[v.NodeGroupName] = cty.ObjectVal(values)
		}
	}
	body.SetAttributeValue("node_pools_labels", cty.ObjectVal(labelsMap))
}

func getTraversalRoute(varName string) hcl.Traversal {
	return hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: varName,
		},
	}
}

func getNodeGroupDetail(message json.RawMessage) ([]NodeGroupDetail, error) {
	var js []NodeGroupDetail
	err := json.Unmarshal(message, &js)
	if err != nil {
		return js, err
	}
	return js, err
}

// func getLabels(message []map[string]interface{}) ([]Label, error) {
// 	var js []Label
// 	jsn, _ := json.Marshal(message)
// 	err := json.Unmarshal(jsn, &js)
// 	if err != nil {
// 		return js, err
// 	}
// 	return js, err
// }

func getNfs(message json.RawMessage) (NfsDetail, error) {
	var js NfsDetail
	err := json.Unmarshal(message, &js)
	if err != nil {
		return js, err
	}
	return js, err
}

func getZone(zone postgres.Jsonb) ([]cty.Value, error) {
	var zoneArray []string
	err := json.Unmarshal(zone.RawMessage, &zoneArray)
	if err != nil {
		return nil, err
	}
	var arr []cty.Value
	for _, v := range zoneArray {
		arr = append(arr, cty.StringVal(v))
	}
	return arr, nil
}
