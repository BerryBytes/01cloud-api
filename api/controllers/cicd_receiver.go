package controllers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"01cloud-api/api/mailer"
	"01cloud-api/api/models"
	"01cloud-api/api/notifications"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/helper"
	"01cloud-api/api/utils/storage"
	"01cloud-api/api/utils/tfgenerator"
	"01cloud-api/api/websocket"

	"github.com/ghodss/yaml"

	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/parnurzeal/gorequest"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	mstore "github.com/berrybytes/01cloud-store/model"
)

const (
	Succeeded = "Succeeded"
	Failed    = "Failed"
)

func (server *Server) OnCompleteCi(ciData interface{}) error {
	data := models.CIResponse{}
	err := json.Unmarshal(ciData.([]byte), &data)
	if err != nil {
		log.Error(err)
		return err
	}
	if data.Status == Succeeded || data.Status == Failed {
		log.Info("status succeeded")
		workflow, err := server.StoreClient.CIWOrkflow().GetWorkflow(data.Name)
		if err != nil {
			log.Error("Workflow info not found")
			return err
		}
		if workflow == nil {
			log.Error("Workflow info not found")
			return err
		}
		log.Debug("workflow :: ", workflow)
		jsonbody, err := json.Marshal(workflow)
		if err != nil {
			log.Error(err)
			return err
		}

		meta := models.WorkflowMetadata{}
		if err := json.Unmarshal(jsonbody, &meta); err != nil {
			log.Error(err)
			return err
		}
		environment := models.Environment{}
		env, err := environment.Find(server.DB, uint64(meta.CIRequest.EnvironmentId))
		if err != nil {
			log.Error("error finding environment:: ", err)
			return err
		}
		attributes := map[string]interface{}{
			"repository_name": meta.CIRequest.RepositoryImage.Name,
			"repository_tag":  meta.CIRequest.RepositoryImage.Tag,
		}
		attrs, _ := json.Marshal(attributes)
		env.Attributes.RawMessage = attrs
		_, err = env.Update(server.DB)
		if err != nil {
			return err
		}
		env.RepositoryImage = &meta.CIRequest.RepositoryImage
		env.CiRequest = &meta.CIRequest
		env.CloneEnvironment = data.CloneEnv
		env.Action = "Deploying"
		wsConn, mutex := websocket.WebsocketConn(fmt.Sprintf("env-%d", env.ID))
		server.sendNotification(env, &data)
		if data.Status == Succeeded {
			if data.IsCustomCI {
				websocket.EmitStatusMessage(Succeeded, helper.GetNamespace(env), "env", wsConn, mutex)
				saving := &mstore.EnvironmentState{
					Namespace: helper.GetNamespace(env),
					PodName:   "env",
					Label:     "env",
					Status:    Succeeded,
					LabelKey:  "env",
				}
				err = server.StoreClient.PodState().StorePodState(0, saving)
				if err != nil {
					return err
				}
				return nil
			}
			err = queue.Publish(constants.CreateRelease, env)
			if err != nil {
				return err
			}
		} else {
			websocket.EmitStatusMessage(Failed, helper.GetNamespace(env), "env", wsConn, mutex)
			saving := &mstore.EnvironmentState{
				Namespace: helper.GetNamespace(env),
				PodName:   "env",
				Label:     "env",
				Status:    Failed,
				LabelKey:  "env",
			}
			err = server.StoreClient.PodState().StorePodState(0, saving)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (server *Server) OnReceiveClusterStatus(clusterData interface{}) error {
	data := models.CIResponse{}
	dat := clusterData.([]byte)
	err := json.Unmarshal(dat, &data)
	if err != nil {
		log.Error(err)
		return err
	}
	meta, err := server.StoreClient.ClusterWOrkflow().GetCreateClusterWorkflow(data.Name)
	if err != nil {
		log.Error("Workflow not found")
		return err
	}
	cr := models.ClusterRequest{}
	cr.ID = uint(meta.CIRequest.ID)
	log.Debug("Workflow status is ", data.Status)
	if data.Status == Succeeded {
		log.Info("Success Status")
		if meta == nil {
			log.Error("Workflow not found")
			return err
		}

		if data.Type == constants.ClusterPlan {
			cr.Status = constants.ClusterStatusPlanned
		} else if data.Type == constants.ClusterApply {
			cr.Status = constants.ClusterStatusApplied
		} else if data.Type == constants.VClusterProvision {
			cr.Status = constants.VClusterProvisioned
		} else if data.Type == constants.ClusterDestroy {
			cr.Status = constants.ClusterStatusDestroyed
		} else if data.Type == constants.InstallPackage {
			cr.Status = constants.ClusterStatusPackageInstalled
		} else if data.Type == constants.UninstallPackage {
			cr.Status = constants.ClusterStatusApplied
		}
		_, err := clusterRequestInterface.ChangeStatus(server.DB, &cr)
		if err != nil {
			log.Error(err)
		}
		crs, err := clusterRequestInterface.FindWithDns(server.DB, uint64(meta.CIRequest.ID))
		if err == nil {
			if data.Type == constants.ClusterApply {
				crs.ID = uint(meta.CIRequest.ID)
				bucketName := tfgenerator.GetBucketName(crs.Organization)
				objectName := tfgenerator.GetObjectName(crs)
				configPath, err := storage.DownloadKubeConfigFile(server.StorageClient, bucketName, objectName, crs.Name)
				if err != nil {
					log.Error(err)
				}
				crs.Cluster.ConfigPath = configPath
				_, err = clusterInterface.Update(server.DB, *crs.Cluster)
				if err != nil {
					log.Error(err)
				}
			}
			if data.Type == constants.InstallPackage {
				crs.Cluster.Active = true

				if crs.Cluster.DNS != nil && len(crs.Cluster.DNS.BaseDomain) > 0 {
					baseDomain := crs.Cluster.DNS.BaseDomain[:len(crs.Cluster.DNS.BaseDomain)-1]
					baseDomain = fmt.Sprintf("%s.%s", helper.GetClusterNamespace(crs.Cluster), baseDomain)
					if checkIsPackageInstalled("prometheus-operator", data.HelmFileString) {
						crs.Cluster.PrometheusServerUrl = fmt.Sprintf("https://%s.%s", "zerone-monitoring", baseDomain)
					}
				}

				_, err = clusterInterface.Update(server.DB, *crs.Cluster)
				if err != nil {
					log.Error(err)
				}
			}
		}
		log.Debug("Cluster Request Model", crs, err)
	} else if data.Status == Failed {
		log.Info("Status Failed")
		// _, err := cr.Find(server.DB, uint64(cr.ID))
		// if err != nil {
		// 	log.Error(err)
		// }
		log.Debug("cr status :: ", cr.Status)
		if data.Type == constants.ClusterPlan {
			cr.Status = constants.ClusterStatusFailed
		} else if data.Type == constants.ClusterApply {
			cr.Status = constants.ClusterStatusPlanned
		} else if data.Type == constants.InstallPackage {
			cr.Status = constants.ClusterStatusApplied
		} else if data.Type == constants.UninstallPackage {
			cr.Status = constants.ClusterStatusPackageInstalled
		} else if data.Type == constants.ClusterDestroy {
			cr.Status = constants.ClusterStatusApplied
		}
		if cr.Status != "" {
			_, err := clusterRequestInterface.ChangeStatus(server.DB, &cr)
			if err != nil {
				log.Error(err)
			}
		}

	}
	return nil
}

func checkIsPackageInstalled(name string, helmfileString string) bool {
	helmFile := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(helmfileString), &helmFile)
	if err != nil {
		log.Error(err)
		return false
	}
	log.Debug("helmFile :: ", helmFile)
	if v, ok := helmFile["releases"]; ok {
		packages := v.([]interface{})
		for _, packageYaml := range packages {
			if packageYaml.(map[string]interface{})["name"].(string) == name {
				return true
			}
		}
	}
	return false
}

func (server *Server) sendNotification(env *models.Environment, data *models.CIResponse) {
	config, err := ciConfigRepo.FindByEnvironment(server.DB, env.ID)
	if err != nil {
		log.Error(err)
		return
	}
	types := "info"
	if data.Status != Succeeded {
		types = "error"
	}
	notifyInfo := notifications.BasicInfoNotification{
		ApplicationName:         env.Application.Name,
		ApplicationID:           int64(env.Application.ID),
		OrganizationID:          int64(env.Application.Project.OrganizationId),
		EnvironmentName:         env.Name,
		EnvironmentID:           int64(env.ID),
		EnvironmentResourceName: env.Resource.Name,
		EnvironmentResourceID:   int64(env.Resource.ID),
	}
	if env.Application.Project.Organization != nil {
		notifyInfo.OrganizationName = env.Application.Project.Organization.Name
	}
	notifyCI(server.NotifyClient, notifyInfo, "build", fmt.Sprintf("environment %s build %s", env.Name, data.Status), types)
	if config.EventType == models.Normal && data.Status != Succeeded {
		return
	} else if config.EventType == models.Error && data.Status == Succeeded {
		return
	}
	attachment1 := slack.Attachment{}
	attachment1.AddField(slack.Field{Title: "Author", Value: "01cloud"}).AddField(slack.Field{Title: "Status", Value: data.Status}).AddField(slack.Field{Title: "Environment Name", Value: env.Name})
	attachment1.AddAction(slack.Action{Type: "button", Text: "Open Environment", Url: fmt.Sprintf("%s/environment/%d", os.Getenv("HOST_URL"), env.ID), Style: "primary"})
	payload := slack.Payload{
		Text:        "01cloud CI Status Update",
		Username:    "01cloud",
		Attachments: []slack.Attachment{attachment1},
	}

	if config.SlackWebhookUrl != "" && config.SlackNotification {
		err := slack.Send(config.SlackWebhookUrl, "", payload)
		if err != nil {
			log.Error(err)
		}
	}
	if config.WebhookUrl != "" && config.WebhookNotification {
		request := gorequest.New().Proxy("")
		_, _, err := request.
			Post(config.WebhookUrl).
			RedirectPolicy(func(req gorequest.Request, via []gorequest.Request) error {
				return fmt.Errorf("incorrect token (redirection)")
			}).
			AppendHeader("token", config.WebhookToken).
			Send(payload).
			End()

		if err != nil {
			log.Error(err)
		}
	}

	emails := strings.Split(config.Emails, ",")
	if len(emails) > 0 && config.EmailNotification {
		ciStatusEmail := mailer.SendCiStatusEmail(emails, uint64(env.ID), env.Name, data.Status)
		err = notifications.NotifyEmail(server.NotifyClient, &ciStatusEmail)
		if err != nil {
			log.Error(err)
		}
	}

}

func (server *Server) testNotification(eid uint, source string) {
	config, err := ciConfigRepo.FindByEnvironment(server.DB, eid)
	if err != nil {
		log.Error(err)
		return
	}
	attachment1 := slack.Attachment{}
	attachment1.AddField(slack.Field{Title: "Author", Value: "01cloud"}).AddField(slack.Field{Title: "Status", Value: "Testing"}).AddField(slack.Field{Title: "Environment Name", Value: config.Environment.Name})
	attachment1.AddAction(slack.Action{Type: "button", Text: "Open Environment", Url: fmt.Sprintf("%s/environment/%d", os.Getenv("HOST_URL"), eid), Style: "primary"})
	payload := slack.Payload{
		Text:        "01cloud CI Status Update",
		Username:    "01cloud",
		Attachments: []slack.Attachment{attachment1},
	}

	if source == "slack" && config.SlackWebhookUrl != "" && config.SlackNotification {
		err := slack.Send(config.SlackWebhookUrl, "", payload)
		if err != nil {
			log.Error(err)
		}

	}
	if source == "webhook" && config.WebhookUrl != "" && config.WebhookNotification {
		request := gorequest.New().Proxy("")
		_, _, err := request.
			Post(config.WebhookUrl).
			RedirectPolicy(func(req gorequest.Request, via []gorequest.Request) error {
				return fmt.Errorf("incorrect token (redirection)")
			}).
			AppendHeader("token", config.WebhookToken).
			Send(payload).
			End()

		if err != nil {
			log.Error(err)
		}
	}

	emails := strings.Split(config.Emails, ",")
	if source == "email" && len(emails) > 0 && config.EmailNotification {
		ciStatusEmail := mailer.SendCiStatusEmail(emails, uint64(eid), config.Environment.Name, "Testing")
		err = notifications.NotifyEmail(server.NotifyClient, &ciStatusEmail)
		if err != nil {
			log.Error(err)
		}
	}

}

func notifyCI(conn *grpc.ClientConn, info notifications.BasicInfoNotification, action, body, types string) {
	_, err := PublishNotificationBase(conn, &notifications.Notification{
		ApplicationName:         info.ApplicationName,
		ApplicationID:           info.ApplicationID,
		EnvironmentName:         info.EnvironmentName,
		EnvironmentID:           info.EnvironmentID,
		EnvironmentResourceName: info.EnvironmentResourceName,
		EnvironmentResourceID:   info.EnvironmentResourceID,
		OrganizationName:        info.OrganizationName,
		OrganizationID:          info.OrganizationID,
		Scope:                   "environment",
		Action:                  action,
		Type:                    types,
		Body:                    body,
		SendBy:                  "api",
	})
	if err != nil {
		log.Error(err)
	}
}
