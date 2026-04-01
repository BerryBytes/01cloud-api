package controllers

import (
	"01cloud-api/api/backup"
	"01cloud-api/api/mailer"
	"01cloud-api/api/models"
	"01cloud-api/api/notifications"
	"01cloud-api/api/responses"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (server *Server) NotifyBackup(w http.ResponseWriter, r *http.Request) {
	notifyBackup := backup.NotifyBackup{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = json.Unmarshal(body, &notifyBackup)
	if err != nil {
		log.Error("can't unmarshall notifyBackup:: ", err)
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	log.Infof("notify backup message received :: %v", notifyBackup)
	environment := &models.Environment{}
	envId := splitNamespace(notifyBackup.Namespace)
	if envId == 0 {
		log.Error("environment id can't be empty")
		responses.ERROR(w, http.StatusBadRequest, errors.New("environment id can't be empty"))
		return
	}
	environmentReceived, err := environment.Find(server.DB, uint64(envId))
	if err != nil {
		log.Error(err)
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	triggeredBy, err := strconv.ParseInt(notifyBackup.TriggeredBy, 10, 64)
	if err != nil {
		log.Error(err)
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	err = server.checkNofityConditions(environmentReceived, server.NotifyClient, &notifyBackup, uint(triggeredBy))
	if err != nil {
		log.Error(err)
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) checkNofityConditions(envirenment *models.Environment, notifyClient *grpc.ClientConn, notifyBackup *backup.NotifyBackup, triggeredBy uint) error {
	//this is the condition for success case
	if notifyBackup.Status == "Completed" {
		if notifyBackup.Resource == "Backup" {
			err := SendBackupEmail(notifyClient, notifyBackup, envirenment, "backup", "completed")
			if err != nil {
				return err
			}
			err = server.StoreActivityLog(notifyBackup.Namespace, "backup completed", "Normal", "backup", envirenment.Name)
			if err != nil {
				return err
			}
		} else {
			StoreNotification(triggeredBy, envirenment, notifyClient, "completed")
			err := SendBackupEmail(notifyClient, notifyBackup, envirenment, "restoration", "completed")
			if err != nil {
				return err
			}
			err = server.StoreActivityLog(notifyBackup.Namespace, "restoration completed", "Normal", "restore", envirenment.Name)
			if err != nil {
				return err
			}
		}
	} else if notifyBackup.Status == "Failed" {
		if notifyBackup.Resource == "Backup" {
			err := SendBackupEmail(notifyClient, notifyBackup, envirenment, "backup", "failed")
			if err != nil {
				return err
			}
			err = server.StoreActivityLog(notifyBackup.Namespace, "backup failed", "Error", "backup", envirenment.Name)
			if err != nil {
				return err
			}
		} else {
			StoreNotification(triggeredBy, envirenment, notifyClient, "failed")
			err := SendBackupEmail(notifyClient, notifyBackup, envirenment, "restoration", "failed")
			if err != nil {
				return err
			}
			err = server.StoreActivityLog(notifyBackup.Namespace, "restoration failed", "Error", "restore", envirenment.Name)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func splitNamespace(namespace string) uint {
	res := strings.Split(namespace, "-")
	if len(res) >= 3 {
		envId, _ := strconv.ParseInt(res[len(res)-1], 10, 64)
		return uint(envId)
	}
	return 0
}

func (server *Server) StoreActivityLog(namespace, message, typ, reason, name string) error {
	event := v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Now(),
			Name:              name,
			Namespace:         namespace,
		},
		InvolvedObject: v1.ObjectReference{
			Kind:      "Backup",
			Namespace: namespace,
			Name:      name,
		},
		Message: message,
		Type:    typ,
		Reason:  reason,
	}
	return server.StoreClient.Event().StoreEventData(&event)
}

func StoreNotification(id uint, envirenment *models.Environment, notifyClient *grpc.ClientConn, statusText string) {
	notifyInfo := notifications.BasicInfoNotification{
		ApplicationName: envirenment.Application.Name,
		ApplicationID:   int64(envirenment.Application.ID),
		EnvironmentID:   int64(envirenment.ID),
		EnvironmentName: envirenment.Name,
		ProjectName:     envirenment.Application.Project.Name,
		ProjectID:       int64(envirenment.Application.Project.ID),
		OrganizationID:  int64(envirenment.Application.Project.OrganizationId),
	}
	if envirenment.Application.Project.Organization != nil {
		notifyInfo.OrganizationName = envirenment.Application.Project.Organization.Name
	}
	notify(notifyClient, notifyInfo, "backup-restored", fmt.Sprintf("restoration %s for environment %s", statusText, envirenment.Name), "info", int64(id))
}

func SendBackupEmail(notifyClient *grpc.ClientConn, notifyBackup *backup.NotifyBackup, environment *models.Environment, backupText, statusText string) error {
	emailsArray := strings.Split(notifyBackup.Emails, ",")
	if len(notifyBackup.Emails) > 0 {
		backupEmail := mailer.SendBackupEmailAfterComplete(emailsArray, notifyBackup, environment, backupText, statusText)
		return notifications.NotifyEmail(notifyClient, &backupEmail)
	}
	return nil
}
