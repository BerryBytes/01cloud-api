package mailer

const (
	VERIFICATION      = "verification"
	VERIFICATION_BODY = "Thank you for creating an account with us. In order to ensure that we have a valid email address on file, we need you to verify your email address by clicking the link below:<br/>" +
		"{{.verifyUrl}} <br/>" +
		"Once you have verified your email address, you will be able to access all of the features and services provided by our platform.<br/>" +
		"If you have any questions or concerns, please contact us for assistance.<br/>" +
		"Thank you for choosing our platform and we look forward to serving you.<br/>{{.button}}"
	UPDATE_ROLE        = "update_role"
	UPDATE_ROLE_BODY   = "We are writing to inform you that your role for the {{.envText}}project {{.project}} has been updated to {{.role}}. This change has been made to better align your skills and expertise with the needs of the project.<br/> Please let us know if you have any questions or concerns about this change.{{.button}}"
	SHARE_PROJECT      = "share_project"
	SHARE_PROJECT_BODY = "We are pleased to inform you that you have been authorized to access the {{.envText}}project {{.project}} as {{.role}}. This authorization grants you access to the necessary resources, tools and information to perform your role effectively.<br/>Please let us know if you have any questions or concerns about this authorization.{{.button}}"
	CI_STATUS          = "ci_status"
	CI_STATUS_BODY     = "Author: 01cloud<br/>Environment Name: {{.envName}} <br/>Status:  {{.status}} {{.button}}"
	INVITE_EMAIL       = "invite_email"
	INVITE_EMAIL_BODY  = "We are excited to invite you to join 01cloud, a platform that provides an array of services and features to support your business.<br/>" +
		"By joining 01cloud, you will have access to a variety of tools and resources that will help you achieve your business goals. Our team is dedicated to providing the best possible experience and we are confident that you will find the platform valuable.<br/>" +
		"Please let us know if you have any questions or concerns, we are always here to help.<br/>" +
		"Thank you for considering 01cloud and we look forward to having you on board.<br/>{{.button}}"
	BLOCK_EMAIL      = "block_email"
	BLOCK_EMAIL_BODY = "We are writing to inform you that your account with us has been blocked. This action has been taken due to a violation of our terms of service or suspicious activity on your account. We take the security of our users' information very seriously and will not tolerate any illegal or unauthorized activity on our platform.<br/>" +
		"If you have any questions or concerns about this action, please contact us immediately so that we can investigate and resolve the issue. In the meantime, your account will remain blocked and you will not be able to access any of our services.<br/>" +
		"We apologize for any inconvenience this may cause and appreciate your cooperation in resolving this matter.<br/>{{.button}}"
	UNBLOCK_EMAIL      = "unblock_email"
	UNBLOCK_EMAIL_BODY = "We are writing to inform you that your account with us has been unblocked. After reviewing the situation, it has been determined that the block was made in error or the issue has been resolved.<br/>" +
		"We apologize for any inconvenience this may have caused and appreciate your patience while we resolved the issue.<br/>{{.button}}" +
		"Thank you for your continued support and we look forward to serving you again.<br/>"
	DEACTIVATE_ACCOUNT      = "deactivate_account"
	DEACTIVATE_ACCOUNT_BODY = "We wanted to confirm that we have received your request to deactivate your account with us. We understand that you have chosen to discontinue using our service, and we respect your decision.<br/>" +
		"Please be aware that this action is permanent and cannot be undone. You will no longer have access to any of the services or information associated with your account. Any data or files stored on your account will also be deleted.<br/>" +
		"If you change your mind and wish to reactivate your account in the future, you will need to create a new account.<br/>" +
		"If you have any other concerns or questions, please do not hesitate to reach out to our customer support team at info@01cloud.com.{{.button}}"
	BACKUP_EMAIL               = "backup_email"
	BACKUP_EMAIL_BODY          = "The {{.backupText}} process is {{.statusText}} for the environment {{.envName}}{{.button}}"
	LINK_UNLINK_EMAIL          = "link_unlink_email"
	LINK_UNLINK_EMAIL_BODY     = "Your {{.serviceName}} account has been {{.linkedUnlinkedText}} successfully.{{.button}}"
	SHARE_ORGANIZATION         = "share_organization"
	SHARE_ORGANIZATION_BODY    = "You have been added to organization  '{{.organization}}' as '{{.role}}'{{.button}}"
	UPDATE_ORGANIZATION        = "update_organization"
	UPDATE_ORGANIZATION_BODY   = "Your role has been updated to  '{{.role}}' in  '{{.organization}}' organization<br/> {{.button}}"
	REQUEST_DEMO_TO_ADMIN      = "request_demo_to_admin"
	REQUEST_DEMO_TO_ADMIN_BODY = "The new request for a demo has arrived with the following details: <br/>Firstname:  {{.data.FirstName}}" +
		"<br/>Lastname:  {{.data.LastName}} <br/>Email: {{.data.Email}} <br/>Company:  {{.data.Company}}{{.button}}"
	REQUEST_DEMO_TO_USER          = "request_demo_to_user"
	REQUEST_DEMO_TO_USER_BODY     = "Thank you for contacting us about your interest in 01cloud . Your request is being processed and we will be in touch as soon as possible.{{.button}}"
	LOGIN_DETAIL                  = "login_details"
	LOGIN_DETAIL_BODY             = "You can login to 01cloud with following details: <br/>Email: {{.email}} <br/>Password:  {{.password}}{{.button}}"
	RESET_PASSWORD                = "reset_password"
	RESET_PASSWORD_BODY           = "Click this link to reset your password. {{.button}}"
	OUTSTANDING_PAYMENT           = "outstanding_payment"
	OUTSTANDING_PAYMENT_BODY      = "We wanted to remind you that a payment of ${{.balance}} is still outstanding. Please pay within {{.days}} days for uniterrupted service. {{.button}}"
	SUSPENDED_EMAIL               = "suspended_email"
	SUSPENDED_EMAIL_BODY          = "We are writing to inform you that your service with us has been suspended due to non-payment. Please pay due amount of ${{.balance}} to resume your services. {{.button}}"
	THRESHOLD_LIMIT_EXCEED        = "threshold_limit_exceed"
	THRESHOLD_LIMIT_EXCEED_BODY   = "We wanted to bring to your attention that your billing amount of ${{.totalCost}} has surpassed the payment threshold limit of ${{.thresholdLimit}}. {{.button}}"
	PARTNER_NEWUSER_EMAIL         = "partner_newuser_email"
	PARTNER_NEWUSER_EMAIL_BODY    = "Here is your login credentials for login into 01cloud. <br/>Email: {{.email}} <br/>Password: {{.password}}{{.button}}"
	PARTNER_ORG_NEWUSER           = "partner_org_newuser"
	PARTNER_ORG_NEWUSER_BODY      = "We are  delighted to inform you that you've been successfully added to {{.org}}. We're excited to have you on board and look forward to working together.</br>Here is your login credentials for login into 01cloud. <br/>Email: {{.email}} <br/>Password: {{.password}}{{.button}}"
	PARTNER_ORG_EXISTINGUSER      = "partner_org_existinguser"
	PARTNER_ORG_EXISTINGUSER_BODY = "We are  delighted to inform you that you've been successfully added to {{.org}}. We're excited to have you on board and look forward to working together.{{.button}}"
	PROJECT_TERMINATION           = "project_termination"
	PROJECT_TERMINATION_BODY      = "We are writing to inform you that your {{.projectCount}} projects has been terminated. This action has been taken due to a inactivity of project for {{.days}} days.<br/>" +
		"Please let us know if you have any questions or concerns.<br/>"
)

var EMAIL_DATA = map[string]string{
	VERIFICATION:             VERIFICATION_BODY,
	UPDATE_ROLE:              UPDATE_ROLE_BODY,
	SHARE_PROJECT:            SHARE_PROJECT_BODY,
	CI_STATUS:                CI_STATUS_BODY,
	INVITE_EMAIL:             INVITE_EMAIL_BODY,
	BLOCK_EMAIL:              BLOCK_EMAIL_BODY,
	UNBLOCK_EMAIL:            UNBLOCK_EMAIL_BODY,
	DEACTIVATE_ACCOUNT:       DEACTIVATE_ACCOUNT_BODY,
	BACKUP_EMAIL:             BACKUP_EMAIL_BODY,
	LINK_UNLINK_EMAIL:        LINK_UNLINK_EMAIL_BODY,
	SHARE_ORGANIZATION:       SHARE_ORGANIZATION_BODY,
	UPDATE_ORGANIZATION:      UPDATE_ORGANIZATION_BODY,
	REQUEST_DEMO_TO_ADMIN:    REQUEST_DEMO_TO_ADMIN_BODY,
	REQUEST_DEMO_TO_USER:     REQUEST_DEMO_TO_USER_BODY,
	RESET_PASSWORD:           RESET_PASSWORD_BODY,
	OUTSTANDING_PAYMENT:      OUTSTANDING_PAYMENT_BODY,
	SUSPENDED_EMAIL:          SUSPENDED_EMAIL_BODY,
	THRESHOLD_LIMIT_EXCEED:   THRESHOLD_LIMIT_EXCEED_BODY,
	LOGIN_DETAIL:             LOGIN_DETAIL_BODY,
	PARTNER_NEWUSER_EMAIL:    PARTNER_NEWUSER_EMAIL_BODY,
	PARTNER_ORG_NEWUSER:      PARTNER_ORG_NEWUSER_BODY,
	PARTNER_ORG_EXISTINGUSER: PARTNER_ORG_EXISTINGUSER_BODY,
}
