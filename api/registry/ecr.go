package registry

import (
	"01cloud-api/api/models"
	"context"
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

type EcrRegistry struct {
	Ecr *ecr.ECR
}

func NewEcr(user models.GitUser) (*EcrRegistry, error) {
	mySession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(user.Region),
		Credentials: credentials.NewStaticCredentialsFromCreds(credentials.Value{
			AccessKeyID:     *aws.String(user.AccessToken),
			SecretAccessKey: *aws.String(user.SecretKey),
		}),
	}))
	svc := ecr.New(mySession)
	_, err := svc.DescribeRepositories(&ecr.DescribeRepositoriesInput{})
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	return &EcrRegistry{Ecr: svc}, nil
}

func EcrLogin(externalLogin models.ExternalLogin) (*models.GitUser, error) {

	mySession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(externalLogin.Region),
		Credentials: credentials.NewStaticCredentialsFromCreds(credentials.Value{
			AccessKeyID:     *aws.String(externalLogin.AccessId),
			SecretAccessKey: *aws.String(externalLogin.SecretKey),
		}),
	}))
	svc := ecr.New(mySession)
	_, err := svc.DescribeRepositories(&ecr.DescribeRepositoriesInput{})
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	gitUser := models.GitUser{}
	gitUser.ServiceUserName = svc.ClientInfo.SigningName
	gitUser.AccessToken = externalLogin.AccessId
	gitUser.SecretKey = externalLogin.SecretKey
	gitUser.Region = externalLogin.Region
	gitUser.ServiceName = externalLogin.Service
	return &gitUser, nil
}

func (ecrClient EcrRegistry) GetRepositories(ctx context.Context, namespace string) ([]RepositoryList, error) {
	repos, err := ecrClient.Ecr.DescribeRepositories(&ecr.DescribeRepositoriesInput{})
	if err != nil {
		return []RepositoryList{}, nil
	}
	var repoNames []RepositoryList
	for _, repo := range repos.Repositories {
		repoNames = append(repoNames, RepositoryList{
			Name: *repo.RepositoryName,
			Uri:  *repo.RepositoryUri,
		})
	}
	return repoNames, nil
}

func (ecrClient EcrRegistry) GetOrganizations(ctx context.Context, limit int) ([]Organization, error) {
	orgs := []Organization{
		{
			ID:   "Default",
			Name: "default",
		},
	}
	return orgs, nil
}

func (ecrClient EcrRegistry) GetTags(ctx context.Context, namespace, repo string, limit int) ([]Tags, error) {
	imgs, err := ecrClient.Ecr.ListImages(&ecr.ListImagesInput{RepositoryName: aws.String(repo)})
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var tagResponse []Tags
	for _, tag := range imgs.ImageIds {
		tagResponse = append(tagResponse, Tags{
			ID:   *tag.ImageTag,
			Name: *tag.ImageTag,
		})
	}
	return tagResponse, nil

}

func (ecrClient EcrRegistry) GetRegistry(ctx context.Context, namespace, repo string) (*Repository, error) {
	src := []string{repo}
	srepo := &ecr.DescribeRepositoriesInput{
		RepositoryNames: aws.StringSlice(src),
	}
	registryData, err := ecrClient.Ecr.DescribeRepositories(srepo)
	if err != nil {
		return nil, err
	}
	ecrRepo := Repository{}
	for _, j := range registryData.Repositories {
		ecrRepo.Name = *j.RepositoryName
		ecrRepo.Uri = *j.RepositoryUri
		ecrRepo.RepositoryArn = *j.RepositoryArn

	}
	return &ecrRepo, nil
}

func (ecrClient EcrRegistry) CreateWebhook(ctx context.Context, namespace, repo, name, url string) error {
	return nil
}

func (ecrClient EcrRegistry) ParseWebhook(body []byte) (*WebhookResponse, error) {
	return nil, nil
}

func (ecrClient EcrRegistry) GetCurrentUser(ctx context.Context) (*User, error) {
	return nil, nil
}
