package imrole

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sort"

	log "github.com/sirupsen/logrus"

	"01cloud-api/api/models"
	"01cloud-api/api/utils/constants"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/cloudflare/cloudflare-go"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
)

func CheckAwsPermissions(request models.ClusterRequest) ([]string, []string, error) {
	requiredAction := constants.AwsRequiredPermission
	return CheckAWSCredentials(requiredAction, request.AccessKey, request.SecretKey, request.Region)
}
func CheckAwsDNSPermissions(request models.DNS) ([]string, []string, error) {
	requiredAction := []string{
		"route53:*",
		"route53domains:*",
	}
	return CheckAWSCredentials(requiredAction, request.AccessKey, request.SecretKey, request.Region)
}

func CheckAWSCredentials(requiredAction []string, accessKey string, secretKey string, region string) ([]string, []string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		return nil, nil, errors.New("invalid tokens")
	}

	svc := iam.New(sess)
	actions := []*string{}

	for _, r := range requiredAction {
		v := r
		actions = append(actions, &v)
	}
	hasPermission := []string{}
	noPermission := []string{}

	userOutput, err := svc.GetUser(&iam.GetUserInput{})
	if err != nil {
		log.Error(err)
		return nil, requiredAction, errors.New("required iam:GetUser permission")
	}
	maxItem := int64(1000)
	actionlist, err := svc.SimulatePrincipalPolicy(&iam.SimulatePrincipalPolicyInput{
		ActionNames:     actions,
		MaxItems:        &maxItem,
		PolicySourceArn: userOutput.User.Arn,
	})
	if err == nil {
		for _, v := range actionlist.EvaluationResults {
			if *v.EvalDecision == "allowed" {
				hasPermission = append(hasPermission, *v.EvalActionName)
			} else {
				noPermission = append(noPermission, *v.EvalActionName)
			}
		}
	}
	return hasPermission, noPermission, err
}

func CheckGCPDNSPermission(request models.DNS) ([]string, []string, error) {
	return CheckGCPCredentials([]string{
		"roles/dns.admin",
	}, request.Credential)
}
func CheckGcpPermissions(request models.ClusterRequest) ([]string, []string, error) {
	requiredRoles := constants.GcpRequiredPermission
	return CheckGCPCredentials(requiredRoles, request.Credential)
}

func CheckGCPCredentials(requiredRoles []string, credentialFile string) ([]string, []string, error) {
	ctx := context.Background()

	client, err := cloudresourcemanager.NewService(ctx, option.WithCredentialsFile(credentialFile))
	if err != nil {
		return nil, nil, err
	}
	content, err := os.ReadFile(credentialFile)
	if err != nil {
		return nil, nil, err
	}
	serviceAccount := struct {
		ClientEmail string `json:"client_email"`
		ProjectId   string `json:"project_id"`
	}{}
	_ = json.Unmarshal(content, &serviceAccount)

	po, err := client.Projects.GetIamPolicy(serviceAccount.ProjectId, &cloudresourcemanager.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return []string{}, requiredRoles, err
	}
	roles := []string{}
	for _, b := range po.Bindings {
		if contains(b.Members, "serviceAccount:"+serviceAccount.ClientEmail) {
			roles = append(roles, b.Role)
		}
	}

	hasPermission := []string{}
	noPermission := []string{}
	roleList := map[string]string{}
	for _, v := range roles {
		roleList[v] = v
	}
	for _, v := range requiredRoles {
		if _, ok := roleList[v]; ok {
			if v == "roles/owner" || v == "roles/editor" {
				return requiredRoles, []string{}, nil
			}
			hasPermission = append(hasPermission, v)
		} else {
			noPermission = append(noPermission, v)
		}
	}
	return hasPermission, noPermission, nil
}
func contains(s []string, searchterm string) bool {
	i := sort.SearchStrings(s, searchterm)
	return i < len(s) && s[i] == searchterm
}

func CheckCloudFlareDNSPermission(request models.DNS) ([]string, []string, error) {
	api, err := cloudflare.New(request.Credential, request.ProjectId)
	hasPermission := []string{}
	noPermission := []string{}
	if err != nil {
		return hasPermission, noPermission, errors.New("invalid dns credentials")
	}
	verifyBody, err := api.ListAPITokensPermissionGroups(context.Background())
	if err != nil {
		return hasPermission, noPermission, errors.New("invalid dns credentials")
	}
	requiredPermission := map[string]string{
		"82e64a83756745bbbb1c9c2701bf816b": "DNS Read",
		"4755a26eedb94da69e1066d98aa820be": "DNS Write",
		"c8fed203ed3043cba015a93ad1616f1f": "Zone Read",
		"517b21aee92c4d89936c976ba6e4be55": "Zone Settings Read",
		"3030687196b94b638145a3953da2b699": "Zone Settings Write",
		"e6d2666161e84845a636613608cee8d5": "Zone Write",
	}
	permissions := []string{
		"DNS Read",
		"DNS Write",
		"Zone Read",
		"Zone Settings Read",
		"Zone Settings Write",
		"Zone Write",
	}
	roleList := map[string]string{}
	for _, groups := range verifyBody {
		roleList[groups.ID] = groups.ID
	}
	for v := range requiredPermission {
		if _, ok := roleList[v]; ok {
			if v == "roles/owner" || v == "roles/editor" {
				return permissions, []string{}, nil
			}
			hasPermission = append(hasPermission, requiredPermission[v])
		} else {
			noPermission = append(noPermission, requiredPermission[v])
		}
	}

	return hasPermission, noPermission, nil
}
