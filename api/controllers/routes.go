package controllers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"01cloud-api/api/auth"

	"01cloud-api/api/middlewares"
	"01cloud-api/api/websocket"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (server *Server) initializeRoutes(base_url string) {
	server.Router.Use(middlewares.CORS)
	server.Router.Use(middlewares.WithLogging)
	server.Router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL(base_url+"/docs/swagger.yaml"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("#swagger-ui"),
	))

	server.serveMetrics()

	server.Router.PathPrefix("/docs").Handler(http.StripPrefix("/docs", http.FileServer(http.Dir("./docs"))))
	server.Router.PathPrefix("/uploads").Handler(http.StripPrefix("/uploads", middlewares.HideFolders(http.FileServer(http.Dir("/data/uploads")))))

	server.setJson("/", server.Home, "GET")
	server.UserRoutes()
	server.SecurityScannerRoutes()
	server.CronJobRoutes()
	server.InitContainerRoutes()
	server.DnsRoutes()
	server.ImageRegistryRoutes()
	server.GroupRoutes()
	server.PluginCategoryRoutes()
	server.CIRoutes()
	server.NotificationRoutes()
	server.OrganizationPlanRoutes()
	server.HelmEnvironmentRoutes()
	server.CreateClusterRoutes()
	server.OrganizationRoutes()
	server.AddonsRoutes()
	server.LoadbalancerRoutes()
	server.InviteRoutes()
	server.StorageRoutes()
	server.BackupRoutes()
	server.OperatorRoutes()
	server.BasicRoutes()
	server.SetRoutes("/ticket", "TICKET_SERVER_URL")
	server.SetRoutes("/payment", "PAYMENT_SERVER_URL")
	server.SetRoutes("/helm", "HELM_CD_SERVER_URL")
	server.SetRoutes("/monitoring", "MONITORING_SERVER_URL")
}

func (server *Server) BasicRoutes() {
	server.setAuth("/subscription/{id}", server.UpdateSubscription, "PUT")
	server.setAuth("/subscription/{id}", server.DeleteSubscription, "DELETE")
	server.setAuth("/subscription/{id}", server.GetSubscription, "GET")
	server.setAuth("/subscription", server.CreateSubscription, "POST")
	server.setAuth("/subscriptions", server.GetSubscriptions, "GET")
	server.setAuth("/subscriptions/user/{id}", server.GetSubscriptionsByUser, "GET")
	server.setAuth("/sidebar", server.SidebarApi, "GET")

	server.setAdmin("/plugin/{id}", server.UpdatePlugin, "PUT")
	server.setAdmin("/plugin/{id}", server.DeletePlugin, "DELETE")
	server.setAuth("/plugin/{id}", server.GetPlugin, "GET")
	server.setAuth("/plugins", server.GetPlugins, "GET")
	server.setAuth("/plugins/active", server.GetAllActivePlugins, "GET")
	server.setAuth("/plugin", server.CreatePlugin, "POST")

	server.setAuth("/plugin/{id}/versions", server.GetPluginVersionByPluginId, "GET")
	server.setAuth("/plugin/{id}/latest-version", server.GetLatestPluginVersionByPluginId, "GET")
	server.setAuth("/plugin-version/{id}", server.UpdatePluginVersion, "PUT")
	server.setAuth("/plugin-version/{id}", server.DeletePluginVersion, "DELETE")
	server.setAuth("/plugin-version/{id}", server.GetPluginVersion, "GET")
	server.setAuth("/plugin-version/{id}/settings", server.GetPluginVersionSetting, "GET")
	server.setAuth("/plugin-version/{id}/config", server.GetPluginVersionConfig, "GET")
	server.setAuth("/plugin/{id}/add-ons", server.GetAddOns, "GET")
	server.setAuth("/plugin-version", server.CreatePluginVersion, "POST")
	server.setAuth("/plugin-versions", server.GetPluginVersions, "GET")

	server.setAuth("/resource/{id}", server.UpdateResource, "PUT")
	server.setAuth("/resource/{id}", server.DeleteResource, "DELETE")
	server.setAuth("/resource/{id}", server.GetResource, "GET")
	server.setAuth("/resource", server.CreateResource, "POST")
	server.setAuth("/resources", server.GetResources, "GET")

	server.setAdmin("/cron-image/{id}", server.UpdateCronImage, "PUT")
	server.setAdmin("/cron-image/{id}", server.DeleteCronImage, "DELETE")
	server.setAuth("/cron-image/{id}", server.GetCronImage, "GET")
	server.setAdmin("/cron-image", server.CreateCronImage, "POST")
	server.setAuth("/cron-images", server.GetCronImages, "GET")

	server.setAuth("/admin/plugins", server.GetAllPlugins, "GET")
	server.setAdmin("/admin/clusters", server.GetClustersForAdmin, "GET")
	server.setAuth("/admin/subscriptions", server.GetSubscriptionsForAdmin, "GET")
	server.setAuth("/admin/subscription", server.CreateSubscriptionForAdmin, "POST")
	server.setAdmin("/admin/plugin-versions", server.GetPluginVersionsForAdmin, "GET")
	server.setAdmin("/admin/cluster/{id}", server.GetClusterWithDetail, "GET")
	server.setAuth("/admin/resources", server.GetResourcesForAdmin, "GET")
	server.setAuth("/admin/dns", server.GetDnssForAdmin, "GET")
	server.setAdmin("/admin/dashboard", server.Dashboard, "GET")
	server.setAdmin("/admin/file/{filename}", server.PutAdminConfig, "PUT")
	server.setAdmin("/admin/shadow/{uid}", server.ShadowUser, "POST")
	server.setAdmin("/admin/file/{filename}", server.GetAdminConfig, "GET")
	server.setAdmin("/admin/email-templates", server.ListEmailTemplate, "GET")
	server.setAdmin("/admin/email-templates/{template}", server.GetEmailTemplate, "GET")
	server.setAdmin("/admin/email-templates/{template}", server.SetEmailTemplate, "PUT")
	server.setJson("/external/subscriptions", server.GetSubscriptionsForPartner, "GET")
	server.setJson("/external/plugins", server.GetPluginsForPartner, "GET")

	server.setAuth("/cluster/{id}", server.UpdateCluster, "PUT")
	server.setAdmin("/cluster/{id}", server.DeleteCluster, "DELETE")
	server.setAdmin("/cluster/{id}", server.GetCluster, "GET")
	server.setAdmin("/cluster", server.CreateCluster, "POST")
	server.setAuth("/cluster/{id}/{pipeline}/{task}/{step}", server.GetClusterPipelineLog, "GET")
	server.setAuth("/cluster/{id}/change-storage", server.AddStorageDetail, "POST")
	server.setAdmin("/clusters", server.GetClusters, "GET")
	server.setAuth("/cluster/{id}/environments", server.GetClusterEnvironments, "GET")
	server.setAuth("/regions", server.GetRegions, "GET")
	server.setAuth("/get-regions", server.GetAllRegions, "GET")
	server.setAuth("/get-clusters", server.GetAllClusters, "GET")
	server.setAdmin("/admin/org/cluster/{oid}", server.GetClustersByOrgForAdmin, "GET")
	server.setAdmin("/admin/user/{uid}/quotas", server.UpdateUserQuotas, "PUT")

	server.setAuth("/role/{id}", server.UpdateUserRole, "PUT")
	server.setAuth("/role/{id}", server.DeleteUserRole, "DELETE")
	server.setAuth("/role/{id}", server.GetUserRole, "GET")
	server.setAuth("/role", server.CreateUserRole, "POST")
	server.setAuth("/roles", server.GetUserRoles, "GET")

	server.setAuth("/activity/{id}", server.UpdateActivity, "PUT")
	server.setAuth("/activity/{id}", server.DeleteActivity, "DELETE")
	server.setAuth("/activity/{id}", server.GetActivity, "GET")
	server.setAuth("/activity", server.CreateActivity, "POST")
	server.setAuth("/project/{pid}/activities", server.GetProjectActivities, "GET")
	server.setAuth("/project/{pid}/activation", server.ActiveDeactiveProject, "POST")

	server.setAuth("/project/{pid}/users", server.GetUsersInProject, "GET")
	server.setAuth("/project/{pid}/user", server.CreateAuthorization, "POST")
	server.setAuth("/project/{pid}/variables", server.GetGlobalVariables, "GET")
	server.setAuth("/project/{pid}/user/{id}", server.UpdateAuthorization, "PUT")
	server.setAuth("/project/{pid}/user/{id}", server.DeleteAuthorization, "DELETE")
	server.setJson("/project/valid", server.CheckProjectValidity, "GET")
	server.setAuth("/environment/{eid}/users", server.GetUsersInEnv, "GET")
	server.setAuth("/environment/{eid}/user", server.CreateAuthorization, "POST")
	server.setAuth("/environment/{eid}/user/{id}", server.UpdateAuthorization, "PUT")
	server.setAuth("/environment/{eid}/user/{id}", server.DeleteAuthorization, "DELETE")
	server.setAuth("/role/{model}/{id}", server.GetAuthMod, "GET")
	server.setAuth("/rename/{model}/{id}", server.RenameModel, "PUT")
	server.setAuth("/getbyname/{model}/{name}", server.GetByNameController, "GET")
	server.setAuth("/authorization/{id}", server.UpdateAuthorization, "PUT")
	server.setAuth("/authorization/{id}", server.DeleteAuthorization, "DELETE")
	server.setAuth("/authorization/{id}", server.GetAuthorization, "GET")
	server.setAuth("/authorization", server.CreateAuthorization, "POST")
	server.setAuth("/authorizations", server.GetAuthorizations, "GET")

	server.setAuth("/project/{id}", server.UpdateProject, "PUT")
	server.setAuth("/project/{id}", server.DeleteProject, "DELETE")
	server.setAuth("/project/{id}", server.GetProject, "GET")
	server.setAuth("/project/{id}/resource", server.GetResourceUsed, "GET")
	server.setAuth("/project", server.CreateProject, "POST")
	server.setAuth("/projects", server.GetProjects, "GET")
	server.setAuth("/project/{pid}/available-storage", server.GetAvailableStorageInProject, "GET")
	server.setAuth("/datatransfer", server.GetDataTransferUsed, "GET")
	server.setAdmin("/user/{id}/projects", server.GetProjectOfUserOnly, "GET")
	server.setAdmin("/admin/project/{oid}", server.GetProjectByOrganizationForAdmin, "GET")

	server.setAuth("/application/{id}", server.UpdateApplication, "PUT")
	server.setAuth("/application/{id}", server.DeleteApplication, "DELETE")
	server.setAuth("/application/{id}", server.GetApplication, "GET")
	server.setAuth("/application/{id}/available-resource", server.GetAvailableResource, "GET")
	server.setAuth("/application", server.CreateApplication, "POST")
	server.setAuth("/project/{id}/applications", server.GetApplicationsByProject, "GET")
	server.setAdmin("/project/{id}/admin-app", server.GetApplicationsByProjectForAdmin, "GET")

	server.setAuth("/environment", server.CreateEnvironment, "POST")
	server.setAuth("/environment/{id}", server.UpdateEnvironment, "PUT")
	server.setAuth("/environment/{id}", server.DeleteEnvironment, "DELETE")
	server.setAuth("/environment/{id}", server.GetEnvironment, "GET")
	server.setAuth("/environment/{id}/launch", server.ReLaunchEnvironment, "GET")
	server.setAuth("/environment/{id}/stop", server.StopEnvironment, "POST")
	server.setAuth("/environment/{id}/start", server.StartEnvironment, "POST")
	server.setAuth("/environment/{id}/schedule", server.ScheduleEnvironment, "POST")
	server.setAuth("/environment/{id}/clone", server.CloneEnvironment, "POST")
	server.setAuth("/environment/{id}/schedule/logs", server.GetScheduleEnvironmentLogs, "GET")
	server.setAuth("/environment/{id}/revision-fetch", server.RevisionEnvironment, "POST")
	server.setAuth("/environment/{id}/revision-list", server.GetRevisionEnvironment, "GET")
	server.setAuth("/environment/{id}/re-deploy", server.RedeployEnvironment, "POST")
	server.setAuth("/environment/{id}/rollback", server.RollbackEnvironment, "POST")
	server.setAuth("/environment/{id}/change-tag", server.ChangeTagEnvironment, "POST")
	server.setAuth("/environment/{id}/change-branch", server.ChangeBranchEnvironment, "POST")
	server.setAuth("/environment/{id}/rerun-ci", server.ReRunCICD, "GET")
	server.setJson("/webhook/{eid}/{userId}", server.WebhookTrigger, "POST")
	server.setJson("/webhook/registry/{eid}/{userId}", server.Webhook, "POST")
	server.setAuth("/environment/{id}/stop-ci", server.StopCIBuild, "GET")
	server.setAuth("/environment/{id}/ci-trigger", server.TriggerCi, "POST")
	server.setAuth("/environment/{id}/ci-trigger", server.GetCiTriggerConfig, "GET")
	server.setAuth("/cd-config", server.GetCDStrategyConfig, "GET")
	server.setAuth("/application/{id}/environments", server.GetEnvironmentsByApplication, "GET")
	server.setAuth("/application/{id}/helm-environments", server.GetHelmEnvironmentsByApplication, "GET")
	server.setAuth("/application/{id}/admin-env", server.GetEnvironmentsByApplicationAdmin, "GET")
	server.setAuth("/environment/{id}/variables", server.GetEnvironmentVariables, "GET")
	server.setAuth("/environment/{id}/variables", server.SetEnvironmentVariables, "POST")
	server.setAuth("/environment/{id}/dns", server.SetDNS, "POST")
	server.setAuth("/environment/{id}/status", server.GetEnvironmentStatus, "GET")
	server.setAuth("/environment/{id}/workflow", server.GetWorkflows, "GET")
	server.setAuth("/environment/{id}/workflow-log", server.GetWorkflowLog, "GET")
	server.setAuth("/environment/{id}/{pipeline}/{task}/{step}", server.GetCIPipelineStepLog, "GET")
	server.setAuth("/environment/{id}/external-secret", server.RetryExternalSecret, "PUT")
	server.setAuth("/environment/{id}/external-secret-sync", server.SyncExternalSecret, "PUT")
	server.setAuth("/environment/{id}/external-secret-log", server.GetExternalSecretActivityLog, "GET")
	server.setAuth("/environment/{id}/external-logging", server.UpdateExternalLogging, "PUT")
	server.setAuth("/project/{id}/insights", server.GetEnvironmentInsights, "GET")
	server.setAuth("/environment/{id}/insights", server.GetEnvironmentInsights, "POST")
	server.setAuth("/environment/{id}/overview", server.GetEnvironmentOverview, "POST")
	server.setAuth("/environment/{id}/hpa-insight", server.GetHpaGraph, "POST")
	server.setAuth("/environment/{id}/fetch-logs", server.FetchLogs, "GET")
	server.setAuth("/environment/{id}/fetch-state", server.FetchEnvironmentState, "GET")
	server.setAuth("/environment/{id}/package-status", server.FetchEnvironmentPackageStatus, "GET")
	server.setAuth("/environment/{id}/activity-log", server.GetEnvironmentActivityLog, "GET")
	server.setAuth("/environment/{id}/state", server.GetEnvironmentState, "GET")
	server.setAuth("/environment/{id}/pods", server.GetPodList, "GET")
	server.setAuth("/environment/{id}/enable-disable-filemanager", server.EnableDisableFileManager, "POST")
	server.setAuth("/environment/{id}/ip-whitelist", server.UpdateEnvironmentWhiteListedIP, "PUT")
	server.setAuth("/environment/{id}/external-url", server.UpdateEnvironmentExternalURL, "PUT")

	server.setAuth("/search", server.Search, "GET")
	server.setAuth("/search/user", server.SearchUser, "GET")

	server.setAuth("/upload", server.UploadFile, "POST")
	server.setAuth("/upload-gcs", server.UploadFile, "POST")

	server.setAuth("/external/connections", server.GetGitServices, "GET")
	server.setAuth("/external/connect/git", server.ConnectToGit, "POST")
	server.setAuth("/external/connect/registry", server.ConnectToRegistry, "POST")
	server.setAuth("/external/revoke/{id}", server.RevokeGitToken, "DELETE")

	server.setAuth("/git/connect", server.ConnectToGit, "POST")
	server.setAuth("/git/revoke/{id}", server.RevokeGitToken, "DELETE")
	server.setAuth("/git/connections", server.GetGitServices, "GET")
	server.setAuth("/git/repos", server.GetRepos, "POST")
	server.setAuth("/git/branches", server.GetBranches, "POST")
	server.setAuth("/git/organizations", server.GetOrganizations, "POST")
	server.setAuth("/git/organization/{name}", server.GetReposByOrganization, "POST")
	server.setAuth("/git/repos/{name}", server.GetReposByOrganization, "POST")

	server.setAuth("/git/repo/{service}", server.GetRepos, "POST")
	server.setAuth("/git/branch/{service}", server.GetBranches, "POST")
	server.setAuth("/git/org/{service}", server.GetOrganizations, "POST")

	server.setAuth("/registry/connect", server.ConnectToRegistry, "POST")
	server.setAuth("/registry/org/{service}", server.GetRegistryOrganizations, "POST")
	server.setAuth("/registry/repos/{service}", server.GetRegistryRepos, "POST")
	server.setAuth("/registry/repo/{service}", server.GetRegistryDetails, "POST")
	server.setAuth("/registry/repo/tags/{service}", server.GetRegistryRepoTags, "POST")
	server.setAuth("/registry/{eid}/settings", server.SaveSettingForRegistryRepo, "POST")
	server.setAuth("/public/file/{filename}", server.GetPublicConfig, "GET")
	server.setAdmin("/public/file/{filename}", server.PutPublicConfig, "PUT")

	server.Router.HandleFunc("/ws", middlewares.SetMiddlewareAuthenticationLegacy(server.Cache, websocket.ServeWs))
	server.Router.HandleFunc("/ws-server", middlewares.SetServerMiddlewareAuthentication(websocket.ServeWs))
	server.Router.HandleFunc("/joblog", middlewares.SetServerMiddlewareAuthentication(server.StoreJobLogs))
	server.Router.HandleFunc("/activity/audit", middlewares.SetServerMiddlewareAuthentication(server.CreateActivity)).Methods("POST")
	server.Router.HandleFunc("/environment/{id}/after-clone", middlewares.SetServerMiddlewareAuthentication(server.AfterClone)).Methods("POST")
	server.Router.HandleFunc("/environment/{id}/start-stop", middlewares.SetServerMiddlewareAuthentication(server.ScheduleStartStopEnvironment)).Methods("POST")
	server.Router.HandleFunc("/project/{id}/deactivate", middlewares.SetServerMiddlewareAuthentication(server.DeactiveProject)).Methods("POST")
	server.Router.HandleFunc("/environment/{id}/schedule-backup", middlewares.SetServerMiddlewareAuthentication(server.ScheduleBackup)).Methods("POST")
	server.Router.HandleFunc("/helm-environment/{id}/start-stop", middlewares.SetServerMiddlewareAuthentication(server.ScheduleStartStopHelmEnvironment)).Methods("POST")
	server.Router.HandleFunc("/project/terminate", middlewares.SetServerMiddlewareAuthentication(server.TerminateProject)).Methods("POST")
	server.Router.HandleFunc("/project/{id}/all-resources", middlewares.SetServerMiddlewareAuthentication(server.deleteAllProjectResource)).Methods("DELETE")
	server.Router.HandleFunc("/publish-queue", middlewares.SetServerMiddlewareAuthentication(server.publishMessage)).Methods("POST")
}

func (server *Server) setJson(path string, next func(http.ResponseWriter, *http.Request), method string) {
	server.Router.HandleFunc(path, middlewares.SetMiddlewareJSON(next)).Methods(method, "OPTIONS")
}

func (server *Server) setAuth(path string, next func(http.ResponseWriter, *http.Request), method string) {
	server.setJson(path, middlewares.SetMiddlewareAuthentication(server.Cache, next), method)
}

func (server *Server) setAdmin(path string, next func(http.ResponseWriter, *http.Request), method string) {
	server.setJson(path, middlewares.SetAdminMiddlewareAuthentication(server.Cache, next), method)
}

// SecurityScannerRoutes
//
// this function is responsible for setting up the security scanner apis
func (server *Server) SecurityScannerRoutes() {
	// env scann routes
	server.setAuth("/scanner/{id}/plugins", server.PluginList, "GET")
	server.setAuth("/scanner/{id}/scan", server.ScanRequest, "POST")
	server.setAuth("/scanner/{id}/reports", server.GetAllScanReports, "GET")

	// cluster scan routes
	server.setAuth("/scanner/cluster-plugins", server.ClusterPluginList, "GET")
	server.setAuth("/scanner/{id}/cluster-scan", server.ClusterScanRequest, "POST")
	server.setAuth("/scanner/{id}/cluster-reports", server.GetClusterAllScanReports, "GET")

	// common routes
	server.setAuth("/scanner/{id}/reports/{reportId}", server.GetScanReport, "GET")
}

func (server *Server) UserRoutes() {
	server.setJson("/user/login", server.Login, "POST")
	server.setJson("/user/login/external", server.LoginExternal, "POST")
	server.setJson("/user/login/sso", server.CreateNewSSOCode, "POST")
	server.setJson("/user/login/sso", server.CheckSSOStatus, "GET")
	server.setAuth("/user/login/sso", server.SubmitSSOCode, "PUT")
	server.setJson("/user/verify/{token}", server.VerifyEmail, "GET")
	server.setJson("/user/resend-verification", server.ResendVerificationEmail, "POST")
	server.setJson("/user/forgot-password", server.ForgotPassword, "POST")
	server.setJson("/user/reset-password", server.ResetPassword, "POST")
	server.setAuth("/user/change-password", server.ChangePassword, "POST")
	server.setJson("/user/logout", server.Logout, "POST")
	server.setJson("/user/logout-all", server.LogoutAll, "POST")
	server.setAuth("/user/sessions", server.GetAllSessionByUserId, "GET")
	server.setAuth("/user/session/{id}/deactivate", server.DeactivateSession, "POST")
	server.setAuth("/user/token", server.CreateToken, "POST")
	server.setAuth("/user/tokens", server.GetTokens, "GET")
	server.setAuth("/user/tokens", server.RevokeAllToken, "DELETE")
	server.setAuth("/user/token/{id}", server.RevokeSingleToken, "DELETE")

	server.setAuth("/profile", server.UpdateProfile, "PUT")
	server.setAuth("/profile", server.GetProfile, "GET")
	server.setJson("/user/new", server.CreateAuth0User, "GET")
	server.setJson("/user/register", server.CreateUser, "POST")
	server.setAuth("/user/{id}", server.GetUser, "GET")
	server.setAuth("/user/{id}", server.UpdateUser, "PUT")
	server.setJson("/partners/campaign", server.CreateCampaignData, "POST")
	server.setAdmin("/user/{id}", server.DeleteUser, "DELETE")
	server.setAdmin("/user/{id}/block", server.BlockUnBlockAccount, "GET")
	server.setAdmin("/user/{id}/change-admin-status", server.ChangeAdminStatus, "GET")
	server.setAuth("/user/account/deactivate", server.DeactivateAccount, "GET")
	server.setAdmin("/users", server.GetUsers, "GET")

}

func (server *Server) CronJobRoutes() {
	server.setAuth("/environment/{id}/cronjob", server.CreateCronJob, "POST")
	server.setAuth("/environment/{id}/cronjob", server.GetCronJobs, "GET")
	server.setAuth("/environment/{id}/cronjob/{cid}", server.UpdateCronJob, "PUT")
	server.setAuth("/environment/{id}/cronjob/{cid}", server.GetCronJob, "GET")
	server.setAuth("/environment/{id}/cronjob/{cid}", server.DeleteCronJob, "DELETE")
	server.setAuth("/environment/{id}/cronjob-status", server.GetCronJobStatus, "GET")
	server.setAuth("/environment/{id}/cronjob-fetch", server.RequestCronJobStatus, "GET")
	server.setAuth("/environment/{id}/cronjob/{cid}/run", server.CreateJobNow, "GET")
	server.setAuth("/environment/{id}/cronjob/{cid}/logs", server.GetCronJobsLogs, "GET")

}

func (server *Server) InitContainerRoutes() {
	server.setAuth("/environment/{id}/init-container", server.CreateInitContainer, "POST")
	server.setAuth("/environment/{id}/init-container", server.GetInitContainers, "GET")
	server.setAuth("/environment/{id}/init-container/{cid}", server.UpdateInitContainer, "PUT")
	server.setAuth("/environment/{id}/init-container/{cid}", server.GetInitContainer, "GET")
	server.setAuth("/environment/{id}/init-container/{cid}", server.DeleteInitContainer, "DELETE")
}

func (server *Server) DnsRoutes() {
	server.setAuth("/dns/{id}", server.UpdateDns, "PUT")
	server.setAuth("/dns/{id}", server.DeleteDns, "DELETE")
	server.setAuth("/dns/{id}", server.GetDns, "GET")
	server.setAuth("/dns", server.CreateDns, "POST")
	server.setAuth("/dns", server.GetDnsList, "GET")
}
func (server *Server) ImageRegistryRoutes() {
	server.setAuth("/registry", server.CreateImageRegistry, "POST")
	server.setAuth("/registries", server.GetImageRegistrys, "GET")
	server.setAuth("/registry/aws-region", server.GetRegion, "GET")
	server.setAuth("/registry-config", server.PostRegistryConfig, "POST")
	server.setAuth("/registry-config", server.GetRegistryConfig, "GET")
	server.setAuth("/registry/{id}", server.UpdateImageRegistry, "PUT")
	server.setAuth("/registry/{id}", server.GetImageRegistry, "GET")
	server.setAuth("/registry/{id}", server.DeleteImageRegistry, "DELETE")
}

func (server *Server) GroupRoutes() {
	server.setAuth("/groups", server.CreateGroup, "POST")
	server.setAuth("/groups", server.GetGroups, "GET")
	server.setAuth("/groups/{gid}", server.GetGroup, "GET")
	server.setAuth("/groups/{gid}", server.UpdateGroup, "PUT")
	server.setAuth("/groups/{gid}", server.DeleteGroup, "DELETE")
	server.setAuth("/groups/{gid}/members", server.AddMembersToGroup, "POST")
	server.setAuth("/groups/{gid}/members", server.DeleteMembersFromGroup, "DELETE")
}
func (server *Server) PluginCategoryRoutes() {
	server.setAuth("/plugin-category", server.CreatePluginCategory, "POST")
	server.setAuth("/plugin-category", server.GetPluginCategorys, "GET")
	server.setAuth("/plugin-category/{pid}", server.GetPluginCategory, "GET")
	server.setAuth("/plugin-category/{pid}", server.UpdatePluginCategory, "PUT")
	server.setAuth("/plugin-category/{pid}", server.DeletePluginCategory, "DELETE")
	server.setAuth("/plugin-category/{pid}/category", server.AddPluginsToPluginCategory, "POST")
	server.setAuth("/plugin-category/{pid}/category", server.DeletePluginFromPluginCategory, "DELETE")
}

func (server *Server) CIRoutes() {
	server.setAuth("/environment/{id}/ci-metrics", server.GetMetrics, "GET")
	server.setAuth("/environment/{id}/workflow/{wid}", server.DeleteWorkflow, "DELETE")
	server.setAuth("/environment/{id}/build-images", server.GetBuildImages, "GET")
	server.setAuth("/environment/{id}/test/{source}", server.TestNotification, "GET")
}

func (server *Server) NotificationRoutes() {
	server.setAuth("/notification/publish", server.PublishNotification, "POST")
	server.setAuth("/notification/{id}", server.UpdateNotification, "PUT")
	server.setAuth("/notification/{id}", server.DeleteNotification, "DELETE")
	server.setAuth("/notification", server.CreateNotification, "POST")
	server.setAuth("/notifications/seen-unseen", server.UpdateMultipleNotification, "POST")
	server.setAuth("/notifications", server.GetNotifications, "GET")
	server.setAuth("/notification/get-unseen-count", server.GetNotificationsCount, "GET")
	server.setAuth("/notification/mark-all-as-read", server.MarkAllAsRead, "POST")
}

func (server *Server) OrganizationRoutes() {
	server.setAuth("/organization", server.CreateOrganization, "POST")
	server.setAuth("/organizations", server.GetOrganizationsList, "GET")
	server.setAdmin("/admin/organizations", server.GetOrganizationsListForAdmin, "GET")
	server.setAuth("/organization", server.GetOrganization, "GET")
	server.setAuth("/organization/{oid}/activities", server.GetOrganizationActivities, "GET")
	server.setAdmin("/admin/organization/{id}", server.GetOrganizationForAdmin, "GET")
	server.setAuth("/organization", server.UpdateOrganization, "PUT")
	server.setAuth("/organization", server.DeleteOrganization, "DELETE")
	server.setAuth("/organization/plugin/{id}", server.AddPluginToOrganization, "GET")
	server.setAuth("/organization/plugin/{id}", server.RemovePluginFromOrganization, "DELETE")
	server.setAuth("/organization/add-plugins", server.AddMultiplePluginToOrganization, "POST")
	server.setAuth("/organization/remove-plugins", server.RemoveMultiplePluginFromOrganization, "POST")
	server.setAuth("/organization/members", server.AddMembersToOrganization, "POST")
	server.setAuth("/organization/members", server.UpdateMembersFromOrganization, "PUT")
	server.setAuth("/organization/members", server.DeleteMembersFromOrganization, "DELETE")
	server.setAuth("/organization/{oid}/switch", server.SwitchOrganization, "GET")
}

func (server *Server) OrganizationPlanRoutes() {
	server.setAuth("/organizationPlans", server.GetOrganizationPlans, "GET")
	server.setAuth("/organizationPlan/{id}", server.GetOrganizationPlan, "GET")
	server.setAdmin("/organizationPlan", server.CreateOrganizationPlan, "POST")
	server.setAdmin("/admin/organizationPlans", server.GetOrganizationPlansForAdmin, "GET")
	server.setAdmin("/organizationPlan/{id}", server.UpdateOrganizationPlan, "PUT")
	server.setAdmin("/organizationPlan/{id}", server.DeleteOrganizationPlan, "DELETE")
}
func (server *Server) HelmEnvironmentRoutes() {
	server.setAuth("/helm-environment", server.CreateHelmEnvironment, "POST")
	server.setAuth("/helm-environment/{id}", server.UpdateHelmEnvironment, "PUT")
	server.setAuth("/helm-environment/{id}", server.DeleteHelmEnvironment, "DELETE")
	server.setAuth("/helm-environment/{id}", server.GetHelmEnvironment, "GET")
	server.setAuth("/helm-environment/{id}/status", server.GetHelmEnvironmentStatus, "GET")
	server.setAuth("/helm-environment/{id}/launch", server.ReLaunchHelmEnvironment, "GET")
	server.setAuth("/helm-environment/{id}/stop", server.StopHelmEnvironment, "POST")
	server.setAuth("/helm-environment/{id}/start", server.StartHelmEnvironment, "POST")
	server.setAuth("/helm-environment/{id}/re-deploy", server.RedeployHelmEnvironment, "POST")
	server.setAuth("/helm-environment/{id}/insights", server.GetHelmEnvironmentInsights, "POST")
	server.setAuth("/helm-environment/{id}/overview", server.GetHelmEnvironmentOverview, "POST")
	server.setAuth("/helm-environment/{id}/fetch-logs", server.FetchHelmLogs, "GET")
	server.setAuth("/helm-environment/{id}/fetch-state", server.FetchHelmEnvironmentState, "GET")
	server.setAuth("/helm-environment/{id}/activity-log", server.GetHelmEnvironmentActivityLog, "GET")
	server.setAuth("/helm-environment/{id}/state", server.GetHelmEnvironmentState, "GET")
	server.setAuth("/helm-environment/{id}/pods", server.GetHelmPodList, "GET")
	server.setAuth("/helm-environment/{id}/schedule", server.ScheduleHelmEnvironment, "POST")
	server.setAuth("/helm-environment/{id}/schedule/logs", server.GetScheduleHelmEnvironmentLogs, "GET")
}

func (server *Server) CreateClusterRoutes() {
	server.setAuth("/create-cluster", server.CreateClusterRequest, "POST")
	server.setAuth("/create-cluster", server.GetCreateClusterRequests, "GET")
	server.setAuth("/create-cluster/{id}", server.GetClusterRequest, "GET")
	server.setAuth("/create-cluster/{id}", server.UpdateClusterRequest, "PUT")
	server.setAuth("/create-cluster/vcluster", server.CreateVClusterRequest, "POST")
	server.setAuth("/create-cluster/vcluster/validate", server.ValidateVClusterToken, "GET")
	server.setAuth("/create-cluster/{id}/label-color", server.UpdateLabelAndColor, "PUT")
	server.setAuth("/create-cluster/{id}/apply", server.ApplyTerraform, "GET")
	server.setAuth("/create-cluster/{id}/cancel", server.CancelPlan, "GET")
	server.setAuth("/create-cluster/{id}/change-active-status", server.EnableDisableCluster, "POST")
	server.setAuth("/create-cluster/{id}/destroy", server.DestroyClusterRequest, "DELETE")
	server.setAuth("/create-cluster/{id}/workflows", server.GetClusterRequestWorkflows, "GET")
	server.setAuth("/create-cluster/{id}/workflow-log", server.GetClusterRequestWorkflowLog, "GET")
	server.setAuth("/create-cluster/{id}/download-files", server.DownloadTfScripts, "GET")

	server.setAuth("/create-cluster/{id}/install-package", server.InstallPackage, "POST")
	server.setAuth("/create-cluster/{id}/uninstall-package", server.UnInstallPackage, "POST")
	server.setAuth("/create-cluster/{id}/package-status", server.CheckPackageStatus, "GET")
	server.setAuth("/create-cluster/{id}/insights", server.GetClusterOverview, "POST")
	server.setAuth("/import-cluster", server.ImportCluster, "POST")
	server.setAuth("/package-config", server.GetPackageConfig, "GET")
	server.setAuth("/package-config", server.UpdatePackageConfig, "POST")
	//server.setAuth("/import-cluster", server.InstallPackage, "POST")
	server.setAuth("/create-cluster/{id}", server.DeleteClusterRequest, "DELETE")
	server.setAuth("/create-cluster/{id}/validate", server.ValidateKubeconfig, "GET")
	server.setAuth("/create-cluster-config", server.GetClusterCreationConfig, "GET")
	server.setAuth("/check-permissions", server.CheckPermission, "POST")
	server.setAuth("/check-dns-permissions", server.ValidateDNSPermission, "POST")
	server.setAuth("/get-permissions", server.GetPermission, "GET")

}
func (server *Server) LoadbalancerRoutes() {
	server.setAuth("/loadbalancer", server.CreateLoadBalancer, "POST")
	server.setAdmin("/loadbalancers", server.GetLoadBalancersForAdmin, "GET")
	server.setAdmin("/loadbalancer/{id}/environment", server.GetEnvironmentByLoadBalancer, "GET")
	server.setAuth("/loadbalancer/{id}", server.GetLoadBalancer, "GET")
	server.setAuth("/project/{id}/loadbalancers", server.GetLoadBalancerByProject, "GET")
	server.setAuth("/loadbalancer/{id}", server.DeleteLoadBalancer, "DELETE")
	server.setAuth("/loadbalancer/{id}/fetch-status", server.FetchLoadBalancerStatus, "GET")
	server.setAuth("/loadbalancer/{id}/status", server.GetLoadBalancerStatus, "GET")
}

func (server *Server) AddonsRoutes() {
	server.setAuth("/environment/{id}/addons", server.GetAddonEnvironments, "GET")
	server.setAuth("/environment/{id}/addons/{aid}", server.GetAddonEnvironment, "GET")
	server.setAuth("/environment/{id}/addons", server.InstallAddOn, "POST")
	server.setAuth("/environment/{id}/addons/{aid}", server.UpdateAddonsEnvironment, "PUT")
	server.setAuth("/environment/{id}/addons/{aid}/external-url", server.UpdateAddonsEnvironmentExternalURL, "PUT")
	server.setAuth("/environment/{id}/addons/{aid}", server.UnInstallAddOn, "DELETE")
	server.setAuth("/environment/{id}/addons-status", server.GetAddOnState, "GET")
}

func (server *Server) OperatorRoutes() {
	server.setAuth("/operators", server.GetOperators, "GET")
	server.setAuth("/operator", server.GetOperator, "GET")
	server.setAdmin("/admin/operators", server.GetOperatorsForAdmin, "GET")
	server.setAdmin("/admin/operator/sync-all", server.SyncAllOperator, "POST")
	server.setAdmin("/admin/operator/sync", server.SyncOperator, "POST")
	server.setAdmin("/admin/operator/{packageName}", server.UpdateOperator, "PUT")
	server.setAdmin("/admin/operator/enable-disable", server.EnableDisableOperator, "POST")
	server.setAuth("/cluster/{id}/operator/install", server.InstallOperator, "POST")
	server.setAuth("/cluster/{id}/operator/{oid}/reinstall", server.ReInstallOperator, "POST")
	server.setAuth("/cluster/{id}/operator/status", server.StatusOperator, "GET")
	server.setAuth("/cluster/{id}/operator/{oid}", server.UnInstallOperator, "DELETE")
	server.setAuth("/cluster/{id}/operators", server.GetOrganizationOperators, "GET")
	server.setAuth("/cluster/{id}/operator/{oid}/activity-log", server.GetOperatorActivityLog, "GET")
	server.setAuth("/environment/{id}/operator-service", server.FetchOperatorEnvironmentService, "GET")
}

func (server *Server) InviteRoutes() {
	server.setJson("/user-invite-request", server.UserRequestDemo, "POST")
	server.setAdmin("/user-invite-validate/{id}", server.UserValidateInvite, "POST")
	server.setJson("/user-invite-register/{token}", server.SignupInvite, "POST")
	server.setJson("/user-invite-register/{token}", server.GetUserInvite, "GET")
	server.setAdmin("/user-invite", server.CreateUserInvite, "POST")
	server.setAdmin("/user-invite/{id}/resend", server.ResendUserInvite, "POST")
	server.setAdmin("/user-invites", server.GetUserInvites, "GET")
	server.setAdmin("/user-invite/{id}", server.UpdateUserInvite, "PUT")
	server.setAdmin("/user-invite/{id}", server.DeleteUserInvite, "DELETE")
}

func (server *Server) StorageRoutes() {
	server.setAuth("/environment/{id}/storage-fetch", server.StorageFetch, "GET")
	server.setAuth("/environment/{id}/storage", server.ListStorageInEnv, "GET")
	server.setAuth("/environment/{id}/storage", server.CreateStorageEnvironment, "POST")
	server.setAuth("/environment/{id}/storage/{sid}", server.UpdateStorageEnvironment, "PUT")
	server.setAuth("/environment/{id}/storage/{sid}", server.DeleteStorageEnvironment, "DELETE")
}

func (server *Server) BackupRoutes() {
	server.setAuth("/environment/{id}/backup", server.ListBackup, "GET")
	server.setAuth("/environment/{id}/backup", server.CreateBackup, "POST")
	server.setAuth("/environment/{id}/restores", server.RestoreBackupList, "GET")
	server.setAuth("/environment/{id}/backup/{bid}/restore", server.RestoreBackup, "POST")
	server.setAuth("/environment/{id}/backup/{bid}/preserve", server.PreserveBackup, "POST")
	server.setAuth("/environment/{id}/backup/{bid}/label", server.AddLabelBackup, "POST")
	server.setAuth("/environment/{id}/backup/{bid}", server.DeleteBackup, "DELETE")
	server.setAuth("/environment/{id}/backup/setting", server.GetSettingBackup, "GET")
	server.setAuth("/environment/{id}/backup/setting", server.SaveSettingBackup, "PUT")

	server.setAuth("/helm-environment/{id}/backup", server.ListHelmBackup, "GET")
	server.setAuth("/helm-environment/{id}/backup", server.CreateHelmBackup, "POST")
	server.setAuth("/helm-environment/{id}/restores", server.RestoreHelmBackupList, "GET")
	server.setAuth("/helm-environment/{id}/backup/{bid}/restore", server.RestoreHelmBackup, "POST")
	server.setAuth("/helm-environment/{id}/backup/{bid}/preserve", server.PreserveHelmBackup, "POST")
	server.setAuth("/helm-environment/{id}/backup/{bid}", server.DeleteHelmBackup, "DELETE")
	server.setAuth("/helm-environment/{id}/backup/setting", server.GetSettingHelmBackup, "GET")
	server.setAuth("/helm-environment/{id}/backup/setting", server.SaveSettingHelmBackup, "PUT")
	server.Router.HandleFunc("/backup/notify", middlewares.SetServerMiddlewareAuthentication(server.NotifyBackup)).Methods("POST")

}

func (server *Server) SetRoutes(path string, envValue string) {
	server.Router.PathPrefix(path).HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		url := fmt.Sprintf("%s%s", os.Getenv(envValue), r.RequestURI)
		proxyReq, err := http.NewRequest(r.Method, url, bytes.NewReader(body))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadGateway)
			return
		}
		copyHeader(proxyReq.Header, r.Header)
		userId, orgId, _ := auth.ExtractTokenID(r)
		proxyReq.Header.Set("x-org-id", fmt.Sprint(orgId))
		if userId != 0 {
			userGotten, _ := userInterface.FindUserByID(server.DB, userId)
			proxyReq.Header.Set("x-user-id", fmt.Sprint(userId))
			if userGotten != nil && userGotten.IsAdmin {
				proxyReq.Header.Set("x-user-role", "ADMIN")
			}
		}
		httpClient := http.Client{}
		resp, err := httpClient.Do(proxyReq)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		response, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		copyHeader(rw.Header(), resp.Header)
		rw.WriteHeader(resp.StatusCode)
		_, _ = rw.Write(response)
	})
}
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func (server *Server) serveMetrics() {
	server.Router.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	server.Router.Handle("/metrics", promhttp.Handler())
}
