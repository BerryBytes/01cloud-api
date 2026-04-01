package mailer

import (
	"01cloud-api/api/backup"
	"01cloud-api/api/models"
	"01cloud-api/api/notifications"
	"fmt"
	"net/http"
	"os"
	"time"
)

func SendVerification(to string, token string) notifications.SendEmailMessage {
	body := GetMessageBody(VERIFICATION)
	verifyUrl := os.Getenv("HOST_URL") + "/user/verify?token=" + token
	body = replaceMessageBody(body, map[string]string{
		"{{.verifyUrl}}": verifyUrl,
	})
	return SendEmail(to, "Verify Email - 01cloud", body, verifyUrl, "Verify Email")
}

func SendUpdateEmail(to string, projectId uint64, project string, role string, isEnv bool, previousRole string) notifications.SendEmailMessage {
	body := GetMessageBody(UPDATE_ROLE)
	projectUrl := fmt.Sprintf("%s/project/%d", os.Getenv("HOST_URL"), projectId)
	envText := func() string {
		if isEnv {
			return "environment of "
		} else {
			return ""
		}
	}()
	body = replaceMessageBody(body, map[string]string{
		"{{.envText}}": envText,
		"{{.project}}": project,
		"{{.role}}":    role,
	})
	return SendEmail(to, "Project role updated - 01cloud", body, projectUrl, "Open Project")
}

func SendShareEmail(to string, projectId uint64, project string, role string, isEnv bool) notifications.SendEmailMessage {
	body := GetMessageBody(SHARE_PROJECT)
	projectUrl := fmt.Sprintf("%s/project/%d", os.Getenv("HOST_URL"), projectId)
	envText := func() string {
		if isEnv {
			return "environment of "
		} else {
			return ""
		}
	}()
	body = replaceMessageBody(body, map[string]string{
		"{{.envText}}": envText,
		"{{.project}}": project,
		"{{.role}}":    role,
	})
	return SendEmail(to, "Project shared - 01cloud", body, projectUrl, "Open Project")
}

func SendCiStatusEmail(to []string, envId uint64, envName string, status string) notifications.SendEmailMessage {
	body := GetMessageBody(CI_STATUS)
	envUrl := fmt.Sprintf("%s/environment/%d", os.Getenv("HOST_URL"), envId)
	body = replaceMessageBody(body, map[string]string{
		"{{.envName}}": envName,
		"{{.status}}":  status,
	})
	return SendMultipleEmail(to, "01cloud CI Status Update", body, envUrl, "Open Environment")
}

func SendInviteEmail(to string, token string) notifications.SendEmailMessage {
	body := GetMessageBody(INVITE_EMAIL)
	verifyUrl := os.Getenv("HOST_URL") + "/user/invite/" + token
	return SendEmail(to, "Invitation to Join 01cloud", body, verifyUrl, "Join 01cloud")
}

func SendBlockUnblockEmail(to string, status string) notifications.SendEmailMessage {
	subject := "Account Status Update - Blocked by Admin"
	body := GetMessageBody(BLOCK_EMAIL)
	if status == "unblock" {
		subject = "Account Unblocked by Admin"
		body = GetMessageBody(UNBLOCK_EMAIL)

	}
	return SendEmail(to, subject, body, "", "")
}

func SendDeactivateAccountEmail(to string) notifications.SendEmailMessage {
	body := GetMessageBody(DEACTIVATE_ACCOUNT)
	return SendEmail(to, "Confirmation of Account Deactivation", body, "", "")
}

func SendBackupEmailAfterComplete(to []string, notifyBackup *backup.NotifyBackup, env *models.Environment, backupText, statusText string) notifications.SendEmailMessage {
	body := GetMessageBody(BACKUP_EMAIL)
	body = replaceMessageBody(body, map[string]string{
		"{{.backupText}}":     backupText,
		"{{.statusText}}":     statusText,
		"{{.envName}}":        env.Name,
		"{{.projName}}":       env.Application.Project.Name,
		"{{.appName}}":        env.Application.Name,
		"{{.creationTime}}":   getTime(notifyBackup.Created),
		"{{.completionTime}}": getTime(notifyBackup.Completed),
		"{{.totalDuration}}":  fmt.Sprintf("%d seconds", notifyBackup.Duration),
		"{{.triggeredBy}}":    fmt.Sprint(notifyBackup.TriggeredByName),
	})
	envUrl := fmt.Sprintf("%s/environment/%d", os.Getenv("HOST_URL"), env.ID)
	return SendMultipleEmail(to, "Backup Event- 01cloud", body, envUrl, "Open Environment")
}

func SendLinkUnlinkEmail(to string, serviceName string, isLinked bool) notifications.SendEmailMessage {
	body := GetMessageBody(LINK_UNLINK_EMAIL)
	linkedUnlinkedText := func() string {
		if isLinked {
			return "linked"
		} else {
			return "unlinked"
		}
	}()
	body = replaceMessageBody(body, map[string]string{
		"{{.serviceName}}":        serviceName,
		"{{.linkedUnlinkedText}}": linkedUnlinkedText,
	})
	subject := fmt.Sprintf("Account %s - 01cloud", linkedUnlinkedText)
	return SendEmail(to, subject, body, "", "")
}

func SendShareOrganization(to string, organization string, role string, orgId int, name string) notifications.SendEmailMessage {
	body := GetMessageBody(SHARE_ORGANIZATION)
	body = replaceMessageBody(body, map[string]string{
		"{{.organization}}": organization,
		"{{.role}}":         role,
	})
	orgUrl := fmt.Sprintf("%s/organization/%d", os.Getenv("HOST_URL"), orgId)
	return SendEmail(to, "Added to Organization - 01cloud", body, orgUrl, "Open Organization")
}

func SendUpdateOrganization(to string, organization string, role string, orgId int, name string) notifications.SendEmailMessage {
	body := GetMessageBody(UPDATE_ORGANIZATION)
	body = replaceMessageBody(body, map[string]string{
		"{{.role}}":         role,
		"{{.organization}}": organization,
	})
	orgUrl := fmt.Sprintf("%s/organization/%d", os.Getenv("HOST_URL"), orgId)
	return SendEmail(to, "Role updated in Organization - 01cloud", body, orgUrl, "Open Organization")
}

func SendRequestDemoEmailToAdmin(data *models.UserInvite) notifications.SendEmailMessage {
	body := GetMessageBody(REQUEST_DEMO_TO_ADMIN)
	body = replaceMessageBody(body, map[string]string{
		"{{.data.FirstName}}": data.FirstName,
		"{{.data.LastName}}":  data.LastName,
		"{{.data.Email}}":     data.Email,
		"{{.data.Company}}":   data.Company,
	})
	to := os.Getenv("ADMIN_EMAIL")
	return SendEmail(to, "Demo Request- 01cloud", body, "", "")
}

func SendRequestDemoEmailToUser(data *models.UserInvite) notifications.SendEmailMessage {
	body := GetMessageBody(REQUEST_DEMO_TO_USER)
	to := data.Email
	return SendEmail(to, "Request Submitted - 01cloud", body, "", "")
}

func SendLoginDetailToUser(to string, password string) notifications.SendEmailMessage {
	body := GetMessageBody(LOGIN_DETAIL)
	body = replaceMessageBody(body, map[string]string{
		"{{.email}}":    to,
		"{{.password}}": password,
	})
	loginUrl := fmt.Sprint(os.Getenv("HOST_URL"))
	return SendEmail(to, "Login Details - 01cloud", body, loginUrl, "Open 01Cloud")
}

func SendPartnerNewUserEmail(to string, envId uint64, password string) notifications.SendEmailMessage {
	body := GetMessageBody(PARTNER_NEWUSER_EMAIL)
	envUrl := fmt.Sprintf("%s/environment/%d", os.Getenv("HOST_URL"), envId)
	body = replaceMessageBody(body, map[string]string{
		"{{.email}}":    to,
		"{{.password}}": password,
	})
	return SendEmail(to, "Partner - 01cloud", body, envUrl, "Go to Environment")
}

func SendPartnerOrgNewUserEmail(to string, envId uint64, password, orgName string) notifications.SendEmailMessage {
	body := GetMessageBody(PARTNER_ORG_NEWUSER)
	envUrl := fmt.Sprintf("%s/environment/%d", os.Getenv("HOST_URL"), envId)
	body = replaceMessageBody(body, map[string]string{
		"{{.org}}":      orgName,
		"{{.email}}":    to,
		"{{.password}}": password,
	})
	return SendEmail(to, "Partner - 01cloud", body, envUrl, "Go to Environment")
}

func SendPartnerOrgExistingUserEmail(to string, envId uint64, orgName string) notifications.SendEmailMessage {
	body := GetMessageBody(PARTNER_ORG_EXISTINGUSER)
	envUrl := fmt.Sprintf("%s/environment/%d", os.Getenv("HOST_URL"), envId)
	body = replaceMessageBody(body, map[string]string{
		"{{.org}}": orgName,
	})
	return SendEmail(to, "Partner - 01cloud", body, envUrl, "Go to Environment")
}

func SendProjectTerminationEmail(to string, projectCount, days int) notifications.SendEmailMessage {
	body := GetMessageBody(PROJECT_TERMINATION)
	body = replaceMessageBody(body, map[string]string{
		"{{.projectCount}}": fmt.Sprint(projectCount),
		"{{.days}}":         fmt.Sprint(days),
	})
	return SendEmail(to, "Project Terminated - 01cloud", body, "", "")
}

type sendMail struct{}

type SendMailer interface {
	SendResetPassword(string, string) (notifications.SendEmailMessage, *EmailResponse)
}

var (
	SendMail SendMailer = &sendMail{}
)

type EmailResponse struct {
	Status   int
	RespBody string
}

func (s *sendMail) SendResetPassword(ToUser string, Token string) (notifications.SendEmailMessage, *EmailResponse) {
	body := GetMessageBody(RESET_PASSWORD)
	forgotUrl := os.Getenv("HOST_URL") + "/resetpassword/" + Token
	return SendEmail(ToUser, "Forgot password - 01cloud", body, forgotUrl, "Reset Password"), &EmailResponse{
		Status:   http.StatusOK,
		RespBody: "Success, Please click on the link provided in your email",
	}
}

func getTime(data int64) string {
	unixTimestamp := data
	unixTime := time.Unix(unixTimestamp, 0)
	return fmt.Sprint(unixTime)
}
