package constants

const (
	CreateRelease                 = "create-release"
	CreateHelmRelease             = "create-helm-release"
	UpgradeHelmRelease            = "upgrade-helm-release"
	DeleteHelmRelease             = "delete-helm-release"
	ListRevision                  = "list-revision"
	CreateCluster                 = "create-cluster"
	CreateLoadBalancer            = "create-load-balancer"
	DestroyLoadBalancer           = "destroy-load-balancer"
	CreateClusterNamespace        = "zerone-create-cluster"
	FetchPackageStatus            = "fetch-package-status"
	ValidateConfig                = "validate-config"
	FetchLoadBalancerStatus       = "fetch-loadbalancer-status"
	ClusterStatusDrafted          = "drafted"
	ClusterStatusPlanned          = "planned"
	ClusterStatusApplied          = "applied"
	ClusterStatusDestroyed        = "destroyed"
	ClusterStatusPackageInstalled = "package-installed"
	ClusterStatusFailed           = "failed"
	TypeImported                  = "imported"
	VClusterProvision             = "vcluster-provision"
	VClusterProvisioned           = "provisioned"
	ClusterPlan                   = "cluster-plan"
	ClusterPlaning                = "cluster-planing"
	ClusterApplying               = "cluster-applying"
	PackageInstalling             = "package-installing"
	InstallPackage                = "package-install"
	UpdatePackage                 = "package-update"
	UninstallPackage              = "package-uninstall"
	ClusterApply                  = "cluster-apply"
	ClusterDestroy                = "cluster-destroy"
	CreateClusterStatus           = "create-cluster-status"
	RollbackRelease               = "rollback-release"
	UpgradeRelease                = "upgrade-release"
	WhitelistedIPS                = "whitelisted-ips"
	DestroyRelease                = "destroy-release"
	InstallAddOn                  = "install-add-on"
	AddonExternalURL              = "addon-external-url"
	UnInstallAddOn                = "uninstall-add-on"
	UpdateAddOn                   = "update-add-on"
	FetchLogs                     = "fetch-log"
	FetchHelmLogs                 = "fetch-helm-log"
	FetchStorage                  = "fetch-storage"
	CreateStorage                 = "create-storage"
	CreatedTemplateStorage        = "created-template-storage"
	UpdateStorage                 = "update-storage"
	DeleteStorage                 = "delete-storage"
	UpdateEnvironmentState        = "update-environment-state"
	ScheduleEnvironment           = "schedule-environment"
	ScheduleHelmEnvironment       = "schedule-helm-environment"
	UpdateHelmEnvironmentState    = "update-helm-environment-state"
	CheckPackageStatus            = "check-package-status"
	BuildWorkflow                 = "build-workflow"
	StopWorkflow                  = "stop-workflow"
	InstallOperator               = "install-operator"
	UnInstallOperator             = "uninstall-operator"
	ReInstallOperator             = "reinstall-operator"
	StatusOperator                = "status-operator"
	InstallOperatorApp            = "install-operator-app"
	UnInstallOperatorApp          = "uninstall-operator-app"
	UpdateOperatorApp             = "update-operator-app"
	FetchOperatorService          = "fetch-operator-service"
	BuildWorkflowSuccess          = "build-workflow-status"
	CreateClusterSuccess          = "create-cluster-status"
	CollectionPodState            = "pod-state"
	CollectionPodStatus           = "pod-status"
	CollectionReleaseInfo         = "release-info"
	CollectionRevisonInfo         = "revision-info"
	CollectionStorageInfo         = "storage-info"
	CollectionHelmService         = "helm-service-log"
	CollectionWorkflowInfo        = "workflow-info"
	CollectionClusterInfo         = "create-cluster-info"
	CollectionClusterLog          = "create-cluster-log"
	CollectionWorkflowLog         = "workflow-log"
	CollectionEventLog            = "event-log"
	CollectionCronJob             = "cron-job-info"
	CollectionLoadBalancerDetail  = "loadbalancer-detail"
	CollectionInitContainer       = "cron-init-container"
	CollectionCertificateInfo     = "certificate-info"
	CreateInitContainer           = "create-init-container"
	DeleteInitContainer           = "delete-init-container"
	UpdateInitContainer           = "update-init-container"
	FetchInitContainer            = "fetch-init-container"
	CreateCronJob                 = "create-cron-job"
	DeleteCronJob                 = "delete-cron-job"
	UpdateCronJob                 = "update-cron-job"
	FetchCronJob                  = "fetch-cron-job"
	PostCreateEnvironment         = "post-create-environment"
	ActiveDeactiveEnvironment     = "active-deactive-environment"
	ActiveDeactiveHelmEnvironment = "active-deactive-helm-environment"
	ExternalSecret                = "external-secret"
	ExternalSecretSync            = "external-secret-sync"
	ExternalSecretSyncStatus      = "external-secret-sync-status"
	ExternalSecretStatus          = "external-secret-status"
	RunJobNow                     = "create-job-now"
	CollectionCronJobLog          = "cron-job-log"
	CollectionCiMetrics           = "ci-metrics"
	ServiceTypeExternal           = "external"
	ServiceTypeInternal           = "internal"
	GCP                           = "gcp"
	CloudFlare                    = "cloudflare"
	CUSTOM                        = "custom"
	EKS                           = "aws"
	Seen                          = "seen"
	All                           = "all"
	Unseen                        = "unseen"
	TagHelm                       = "helm-env"
	TagEnv                        = "env"
	EnableDisableFileManager      = "enable-disable-filemanager"
	Vault                         = "vault"
	Elastic                       = "elastic"
	Loki                          = "loki"
	Kafka                         = "kafka"
	Cloudwatch                    = "cloudwatch"
	S3                            = "s3"
	ExternalLogging               = "external-logging"
	ErrorQueue                    = "error-queue"
	Dockerfile                    = "Dockerfile"

	EnvironmentExternalURL = "environment-external-url"
)

const OperatorBaseUrl = "https://operatorhub.io/api/"

var (
	GcpRequiredPermission = []string{
		"roles/compute.networkAdmin",
		"roles/compute.securityAdmin",
		"roles/file.editor",
		"roles/container.admin",
		"roles/iam.serviceAccountUser",
		"roles/resourcemanager.projectIamAdmin",
		"roles/serviceusage.serviceUsageAdmin",
	}
	AwsRequiredPermission = []string{
		"autoscaling:*",
		"ec2:*",
		"iam:*",
		"logs:*",
		"eks:DescribeCluster",
		"kms:CreateKey",
		"kms:DescribeKey",
	}
	AwsLoggerPermission = []string{
		"s3:*",
	}
	CloudwatchLoggerPermission = []string{
		"logs:*",
	}
	GCPLoggerPermission = []string{
		"roles/logging.logWriter",
		"roles/storage.objectCreator",
		"roles/logging.bucketWriter",
		"roles/logging.viewer",
	}
)

var (
	PackagesNamespaces = map[string]string{
		"sealed-secret":        "zerone-sealed-secrets",
		"cert-manager":         "zerone-cert-manager",
		"tekton":               "tekton-pipelines",
		"contour":              "zerone-projectcontour",
		"prometheus":           "zerone-monitoring",
		"lb-controller":        "zerone-lb-controller",
		"dns-controller":       "zerone-dns-controller",
		"tekton-controller":    "tekton-pipelines",
		"image-del-controller": "zerone-image-del-controller",
		"zerone-jobs":          "zerone-jobs",
		"reloader":             "zerone-reloader",
		"velero":               "velero",
		"olm":                  "olm",
		"secret-patcher":       "zerone-secret-patcher",
		"external-secrets":     "zerone-external-secret",
		"external-logging":     "zerone-external-logging",
	}
)
