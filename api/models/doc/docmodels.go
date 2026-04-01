package doc

import (
	"01cloud-api/api/models"
	"time"
)

type Application struct {
	ID                  uint          `json:"id,omitempty"`
	CreatedAt           time.Time     `json:"createdat,omitempty"`
	Name                string        `json:"name"`
	Project             *Project      `json:"project"`
	ProjectID           uint64        `json:"project_id"`
	Plugin              *Plugin       `json:"plugin"`
	PluginID            uint64        `json:"plugin_id"`
	Cluster             *Cluster      `json:"cluster"`
	ClusterID           uint64        `json:"cluster_id,omitempty"`
	Chart               *Chart        `json:"chart"`
	ChartID             string        `json:"chart_id,omitempty"`
	Owner               *UserResponse `json:"owner"`
	OwnerId             uint64        `json:"owner_id,omitempty"`
	GitRepository       *GitRepo      `json:"git_repository_info"`
	GitUrl              string        `json:"git_repository"`
	GitRepoUrl          *string       `json:"git_repo_url"`
	GitService          string        `json:"git_service"`
	ImageUrl            string        `json:"image_url"`
	ImageNamespace      string        `json:"image_namespace"`
	ImageRepo           string        `json:"image_repo"`
	ImageService        string        `json:"image_service"`
	ServiceType         int           `json:"service_type"`
	Active              bool          `json:"active"`
	OperatorPackageName string        `json:"operator_package_name"`
}
type Chart struct {
	Name            string          `json:"name"`
	Repo            *Repo           `json:"repo,omitempty"`
	RepoID          string          `json:"repo_id"`
	OrganizationID  int64           `json:"organization_id"`
	Description     string          `json:"description"`
	Home            string          `json:"text"`
	Icon            string          `json:"icon"`
	IconContentType string          `json:"icon_content_type"`
	Category        string          `json:"category"`
	ChartVersions   []*ChartVersion `json:"chart_versions"`
}

type ChartVersion struct {
	Chart      *Chart      `json:"chart,omitempty"`
	ChartID    string      `json:"chart_id"`
	Version    string      `json:"version"`
	AppVersion string      `json:"app_version"`
	Digest     string      `json:"digest"`
	URLs       interface{} `json:"urls"`
	Readme     string      `json:"readme"`
	Values     string      `json:"values"`
	Schema     string      `json:"schema"`
}
type Repo struct {
	Name           string      `json:"name"`
	URL            string      `json:"url"`
	RepoType       string      `json:"repo_type"`
	Filter         string      `json:"filter"`
	OrganizationId int64       `json:"organization_id"`
	Authorization  interface{} `json:"authorization"`
}
type GitRepo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Owner    string `json:"owner"`
	HtmlUrl  string `json:"html_url"`
	CloneURL string `json:"clone_url"`
	GitURL   string `json:"git_url"`
}

type MarkAsSeeNotificationRequest struct {
	ID        uint      `json:"id,omitempty"`
	CreatedAt time.Time `json:"createdat,omitempty"`
	Ids       []uint64  `json:"ids"`
	Seen      bool      `json:"seen"`
}

type Authorization struct {
	ID            uint          `json:"id,omitempty"`
	CreatedAt     time.Time     `json:"createdat,omitempty"`
	Email         string        `gorm:"size:255;null;" json:"email"`
	User          *UserResponse `gorm:"foreignkey:UserID;null" json:"user"`
	UserID        uint64        `gorm:"default:0;" json:"user_id"`
	UserRole      *UserRole     `gorm:"foreignkey:UserRoleID" json:"user_role"`
	UserRoleID    uint64        `gorm:"not null;" json:"user_role_id"`
	Project       *Project      `gorm:"foreignkey:ProjectID" json:"project"`
	ProjectID     uint64        `gorm:"not null;" json:"project_id"`
	Application   *Application  `gorm:"foreignkey:ProjectID" json:"application"`
	ApplicationID uint64        `gorm:"not null;default:0" json:"application_id"`
	Environment   *Environment  `gorm:"foreignkey:EnvironmentID;null" json:"environment"`
	EnvironmentID uint64        `gorm:"not null;default:0;" json:"environment_id"`
	Active        bool          `gorm:"not null;" json:"active"`
	Attributes    string        `gorm:"null;" json:"attributes"`
	Group         *Group        `gorm:"foreignkey:GroupID;null" json:"group"`
	GroupID       uint64        `gorm:"default:0;" json:"group_id"`
}

type Cluster struct {
	ID                  uint            `json:"id,omitempty"`
	CreatedAt           time.Time       `json:"createdat,omitempty"`
	Name                string          `json:"name"`
	Context             string          `json:"context"`
	ConfigPath          string          `json:"configPath"`
	Region              string          `json:"region"`
	Provider            string          `json:"provider"`
	Zone                string          `json:"zone"`
	DNS                 *DNS            `json:"dns,omitempty"`
	DNSId               uint64          `json:"dns_id"`
	Labels              string          `json:"labels"`
	Nodes               uint64          `json:"nodes"`
	PvCapacity          uint64          `json:"pv_capacity"`
	Weight              uint32          `json:"weight"`
	ImageRegistry       *ImageRegistry  `json:"image_registry,omitempty"`
	ImageRegistryID     uint64          `json:"image_registry_id"`
	Active              bool            `json:"active"`
	Organization        *Organization   `json:"organization,omitempty"`
	OrganizationID      uint64          `json:"organization_id"`
	TotalMemory         uint64          `json:"total_memory"`
	CloudStorage        interface{}     `json:"cloud_storage"`
	ProvisionPercentage float64         `json:"provision_percentage"`
	ClusterRequest      *ClusterRequest `json:"cluster_request"`
	ClusterRequestID    uint64          `json:"cluster_request_id"`
	Color               string          `json:"color"`
	StorageClass        string          `json:"storage_class"`
}

type ClusterDetails struct {
	ID                  uint            `json:"id,omitempty"`
	CreatedAt           time.Time       `json:"createdat,omitempty"`
	Name                string          `json:"name"`
	Context             string          `json:"context"`
	ConfigPath          string          `json:"configPath"`
	Region              string          `json:"region"`
	Provider            string          `json:"provider"`
	Zone                string          `json:"zone"`
	DNS                 *DNS            `json:"dns,omitempty"`
	DNSId               uint64          `json:"dns_id"`
	Labels              string          `json:"labels"`
	Nodes               uint64          `json:"nodes"`
	PvCapacity          uint64          `json:"pv_capacity"`
	Weight              uint32          `json:"weight"`
	ImageRegistry       *ImageRegistry  `json:"image_registry,omitempty"`
	ImageRegistryID     uint64          `json:"image_registry_id"`
	Active              bool            `json:"active"`
	Organization        *Organization   `json:"organization,omitempty"`
	OrganizationID      uint64          `json:"organization_id"`
	TotalMemory         uint64          `json:"total_memory"`
	ProvisionPercentage float64         `json:"provision_percentage"`
	ClusterRequest      *ClusterRequest `json:"cluster_request"`
	ClusterRequestID    uint64          `json:"cluster_request_id"`
	Color               string          `json:"color"`
	ProjectName         string          `json:"project_name"`
	Attributes          string          `json:"attributes"`
	PrometheusServerUrl string          `josn:"prometheus_server_url"`
	CloudStorage        interface{}     `json:"cloud_storage"`
	StorageAccessKey    string          `json:"storage_access_key"`
	StorageSecretKey    string          `json:"storage_secret_key"`
}
type ClusterRequest struct {
	ID                    uint          `json:"id,omitempty"`
	CreatedAt             time.Time     `json:"createdat,omitempty"`
	Name                  string        `json:"cluster_name"`
	Version               string        `json:"cluster_version"`
	Region                string        `json:"region"`
	Zone                  interface{}   `json:"zone"`
	Provider              string        `json:"provider"`
	ProviderName          string        `json:"provider_name"`
	VpcName               string        `json:"vpc_name"`
	NetworkCidr           string        `json:"network_cidr"`
	NetworkPolicy         bool          `json:"network_policy"`
	PvcWriteMany          bool          `json:"pvc_write_many"`
	Active                bool          `json:"active"`
	Status                string        `json:"status"`
	Type                  string        `json:"type"`
	TLS                   string        `json:"tls"`
	NfsDetail             interface{}   `json:"nfs_detail"`
	RegionalCluster       bool          `json:"regional_cluster"`
	RemoveDefaultNodePool bool          `json:"remove_default_node_pool"`
	NodeGroupCount        uint64        `json:"node_group_count"`
	NodeGroupDetail       interface{}   `json:"node_group_detail"`
	Organization          *Organization `json:"organization,omitempty"`
	OrganizationID        uint64        `json:"organization_id"`
	Cluster               *Cluster      `json:"cluster,omitempty"`
	ClusterID             uint64        `json:"cluster_id"`
}
type ClusterRequestDetails struct {
	ID                    uint          `json:"id,omitempty"`
	CreatedAt             time.Time     `json:"createdat,omitempty"`
	Name                  string        `json:"cluster_name"`
	Version               string        `json:"cluster_version"`
	Region                string        `json:"region"`
	Zone                  interface{}   `json:"zone"`
	Provider              string        `json:"provider"`
	ProviderName          string        `json:"provider_name"`
	VpcName               string        `json:"vpc_name"`
	NetworkCidr           string        `json:"network_cidr"`
	NetworkPolicy         bool          `json:"network_policy"`
	PvcWriteMany          bool          `json:"pvc_write_many"`
	Active                bool          `json:"active"`
	Status                string        `json:"status"`
	Type                  string        `json:"type"`
	TLS                   string        `json:"tls"`
	NfsDetail             interface{}   `json:"nfs_detail"`
	RegionalCluster       bool          `json:"regional_cluster"`
	RemoveDefaultNodePool bool          `json:"remove_default_node_pool"`
	NodeGroupCount        uint64        `json:"node_group_count"`
	NodeGroupDetail       interface{}   `json:"node_group_detail"`
	Organization          *Organization `json:"organization,omitempty"`
	OrganizationID        uint64        `json:"organization_id"`
	Cluster               *Cluster      `json:"cluster,omitempty"`
	ClusterID             uint64        `json:"cluster_id"`
	ProjectId             string        `json:"project_id"`
	Credential            string        `json:"credentials"`
	AccessKey             string        `json:"access_key"`
	SecretKey             string        `json:"secret_key"`
	SubnetCidrRange       string        `json:"subnet_cidr_range"`
	ErrorMessage          interface{}   `json:"error_message"`
}

type CronJob struct {
	ID                         uint      `json:"id,omitempty"`
	CreatedAt                  time.Time `json:"createdat,omitempty"`
	EnvironmentID              uint
	Image                      string        `json:"image"`
	Name                       string        `json:"name"`
	RestartPolicy              *string       `json:"restart_policy"`
	ConcurrentPolicy           *string       `json:"concurrent_policy"`
	Command                    string        `json:"command"`
	Labels                     interface{}   `json:"labels"`
	Schedule                   string        `json:"schedule"`
	StartingDeadlineSeconds    *int64        `json:"starting_deadline_seconds"`
	FailedJobsHistoryLimit     *int32        `json:"failed_jobs_history_limit"`
	SuccessfulJobsHistoryLimit *int32        `json:"successful_job_history_limit"`
	Suspend                    *bool         `json:"suspend"`
	User                       *UserResponse `json:"user"`
	UserID                     uint64        `json:"user_id"`
}
type LoadBalancer struct {
	ID           uint      `json:"id,omitempty"`
	CreatedAt    time.Time `json:"createdat,omitempty"`
	Name         string    `json:"name"`
	CustomDomain string    `json:"custom_domain"`
	Cluster      *Cluster  `json:"cluster,omitempty"`
	ClusterID    uint64    `json:"cluster_id"`
	Project      *Project  `json:"project,omitempty"`
	ProjectID    uint64    `json:"project_id"`
}
type UserInvite struct {
	ID          uint      `json:"id,omitempty"`
	CreatedAt   time.Time `json:"createdat,omitempty"`
	UpdatedAt   time.Time `json:"updatedat,omitempty"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email,omitempty"`
	Company     string    `json:"company"`
	Designation string    `json:"designation"`
	EmailSent   bool      `json:"email_sent"`
	Remarks     string    `json:"remarks"`
}
type UserRole struct {
	ID          uint      `json:"id,omitempty"`
	CreatedAt   time.Time `json:"createdat,omitempty"`
	Name        string    `json:"name"`
	Code        uint64    `json:"code"`
	Description string    `json:"description"`
	Active      bool      `json:"active"`
}

type CronImage struct {
	ID        uint      `json:"id,omitempty"`
	CreatedAt time.Time `json:"createdat,omitempty"`
	Name      string    `json:"name"`
	ImageName string    `json:"image_name"`
	Version   string    `json:"version"`
	Active    bool      `json:"active"`
}

type InitContainer struct {
	ID            uint      `json:"id,omitempty"`
	CreatedAt     time.Time `json:"createdat,omitempty"`
	EnvironmentID uint
	Image         string `json:"image"`
	Name          string `json:"name"`
	Command       string `json:"command"`
}

type SuccessResponse struct {
	ID        uint      `json:"id,omitempty"`
	CreatedAt time.Time `json:"createdat,omitempty"`
	Message   string    `json:"message"`
}
type ObjectResponse struct {
	ID        uint      `json:"id,omitempty"`
	CreatedAt time.Time `json:"createdat,omitempty"`
}

type NotificationCountResponse struct {
	ID         uint      `json:"id,omitempty"`
	CreatedAt  time.Time `json:"createdat,omitempty"`
	Message    string    `json:"message"`
	ShowBubble bool      `json:"show_bubble"`
	Count      string    `json:"count"`
}

type LoadBalancerStatusResponse struct {
	ID        uint      `json:"id,omitempty"`
	CreatedAt time.Time `json:"createdat,omitempty"`
	Namespace string
	Status    string
	Service   map[string]interface{}
}

type RedeployRequest struct {
	ID           uint                   `json:"id,omitempty"`
	CreatedAt    time.Time              `json:"createdat,omitempty"`
	Version      map[string]interface{} `json:"version"`
	OtherVersion map[string]interface{} `json:"other_version"`
}

type CIRequest struct {
	ID                uint      `json:"id,omitempty"`
	CreatedAt         time.Time `json:"createdat,omitempty"`
	EnvironmentId     int64
	Namespace         string
	ConfigPath        string
	Name              string
	ImageRepoUsername string
	ImageRepoPassword string
	ImageRepoService  string
	ImageRepoProject  string
	GitUrl            string
	GitBranch         string
	GitUserName       string
	GitAccessToken    string
	RepositoryImage   RepositoryImage
	PluginUrl         string
	Author            string
	CommitMessage     string
	BaseImage         string
	BaseTag           string
}

type DNS struct {
	ID             uint          `json:"id,omitempty"`
	CreatedAt      time.Time     `json:"createdat,omitempty"`
	Name           string        `json:"name"`
	Provider       string        `json:"provider"`
	ProjectId      string        `json:"project_id"`
	Region         string        `json:"region"`
	ZoneID         string        `json:"zone_id"`
	TLS            string        `json:"tls"`
	BaseDomain     string        `json:"base_domain"`
	Active         bool          `json:"active"`
	Organization   *Organization `json:"organization,omitempty"`
	OrganizationID uint64        `json:"organization_id"`
}

type DNSDetails struct {
	ID             uint          `json:"id,omitempty"`
	CreatedAt      time.Time     `json:"createdat,omitempty"`
	Name           string        `json:"name"`
	Provider       string        `json:"provider"`
	ProjectId      string        `json:"project_id"`
	Region         string        `json:"region"`
	ZoneID         string        `json:"zone_id"`
	TLS            string        `json:"tls"`
	BaseDomain     string        `json:"base_domain"`
	Active         bool          `json:"active"`
	Organization   *Organization `json:"organization,omitempty"`
	OrganizationID uint64        `json:"organization_id"`
	Credential     string        `json:"credentials"`
	AccessKey      string        `json:"access_key"`
	SecretKey      string        `json:"secret_key"`
}

type ImageRegistry struct {
	ID             uint          `json:"id,omitempty"`
	CreatedAt      time.Time     `json:"createdat,omitempty"`
	Name           string        `json:"name"`
	Service        string        `json:"service"`
	Provider       string        `json:"provider"`
	ProjectName    string        `json:"project_name"`
	UserName       string        `json:"user_name"`
	Active         bool          `json:"active"`
	Organization   *Organization `json:"organization,omitempty"`
	OrganizationID uint64        `json:"organization_id"`
}

type ImageRegistryDetails struct {
	ID             uint          `json:"id,omitempty"`
	CreatedAt      time.Time     `json:"createdat,omitempty"`
	Name           string        `json:"name"`
	Service        string        `json:"service"`
	Provider       string        `json:"provider"`
	ProjectName    string        `json:"project_name"`
	UserName       string        `json:"user_name"`
	Active         bool          `json:"active"`
	Organization   *Organization `json:"organization,omitempty"`
	OrganizationID uint64        `json:"organization_id"`
	Password       string        `json:"password"`
	Credentials    interface{}   `json:"credentials"`
}
type VariableResponse struct {
	ID              uint                `json:"id,omitempty"`
	CreatedAt       time.Time           `json:"createdat,omitempty"`
	SystemVariables map[string]string   `json:"system_variables"`
	UserVariables   []map[string]string `json:"user_variables"`
}

type InsightResponse struct {
	ID           uint                   `json:"id,omitempty"`
	CreatedAt    time.Time              `json:"createdat,omitempty"`
	TotalCpu     int                    `json:"total_cpu"`
	TotalMemory  int                    `json:"total_memory"`
	MemoryUsages int                    `json:"memory_usages"`
	TotalPv      int                    `json:"total_pv"`
	CpuUsages    int                    `json:"cpu_usages"`
	DataTransfer map[string]interface{} `json:"data_transfer"`
}

type HpaInsightResponse struct {
	ID                uint      `json:"id,omitempty"`
	CreatedAt         time.Time `json:"createdat,omitempty"`
	CurrentMinReplica int       `json:"current_min_replica"`
	CurrentMaxReplica int       `json:"current_max_replica"`
	MinReplica        int       `json:"min_replica"`
	MaxReplica        int       `json:"max_replica"`
	Desired           int       `json:"desired"`
	Running           int       `json:"running"`
}

type Environment struct {
	ID                 uint              `json:"id,omitempty"`
	CreatedAt          time.Time         `json:"createdat,omitempty"`
	Name               string            `json:"name"`
	Application        *Application      `json:"application"`
	ApplicationID      uint64            `json:"application_id"`
	Resource           *Resource         `json:"resource"`
	ResourceID         uint64            `json:"resource_id"`
	PluginVersion      *PluginVersion    `json:"plugin_version"`
	PluginVersionID    uint64            `json:"plugin_version_id"`
	Replicas           uint16            `json:"replicas"`
	GitUrl             string            `json:"git_url"`
	GitRepository      *GitRepo          `json:"git_repository_info"`
	GitBranch          string            `json:"git_branch"`
	ImageTag           string            `json:"image_tag"`
	ImageUrl           string            `json:"image_url"`
	ServiceType        int               `json:"service_type"` // template/git/image
	Variables          interface{}       `json:"variables"`
	Version            interface{}       `json:"version"`
	OtherVersion       interface{}       `json:"other_version"`
	UserVariables      interface{}       `json:"user_variables"`
	Active             bool              `json:"active"`
	ApplyImmediately   bool              `json:"apply_immediately"`
	Attributes         interface{}       `json:"attributes"`
	RepositoryImage    *RepositoryImage  `json:"repository_image"`
	CiRequest          *CIRequest        `json:"ci_request"`
	AutoScaler         interface{}       `json:"auto_scaler"`
	Storage            []*Storage        `gorm:"foreignkey:EnvironmentID;association_foreignkey:ID"`
	CronJob            []*CronJob        `gorm:"foreignkey:EnvironmentID;association_foreignkey:ID"`
	InitContainers     []*InitContainer  `gorm:"foreignkey:EnvironmentID;association_foreignkey:ID"`
	LoadBalancer       *LoadBalancer     `gorm:"foreignkey:LoadBalancerID" json:"load_balancer" `
	DeploymentStrategy interface{}       `sql:"json" json:"deployment_strategy"`
	LoadBalancerID     uint64            `json:"load_balancer_id"`
	Parent             *Environment      `json:"parent" `
	ParentID           uint64            `json:"parent_id"`
	Action             string            `json:"action"`
	Scripts            *models.Script    `json:"scripts"`
	OperatorPayload    interface{}       `json:"operator_payload"`
	Schedules          interface{}       `json:"schedules"`
	FileManagerEnabled *time.Time        `json:"file_manager_enabled"`
	CloneEnvironment   *CloneEnvironment `json:"clone_environment"`
	ExternalSecret     interface{}       `json:"external_secret"`
	ErrorMessage       interface{}       `json:"error_message"`
	ExternalLogging    interface{}       `json:"external_logging"`
	Setting            interface{}       `json:"setting"`
	ExternalURL        bool              `json:"external_url"`
	WhitelistedIPs     string            `json:"whitelisted_ips"`
}

type CloneEnvironment struct {
	Name                  string `json:"name"`
	EnvId                 int64  `json:"id"`
	OldEnvId              int64  `json:"old_env_id"`
	IsResource            bool   `json:"is_resource"`
	IsUserPermission      bool   `json:"is_user_permission"`
	IsCronJob             bool   `json:"is_cron_job"`
	IsCIConfig            bool   `json:"is_ci_config"`
	IsCDConfig            bool   `json:"is_cd_config"`
	IsBackupSetting       bool   `json:"is_backup_setting"`
	IsAddon               bool   `json:"is_addon"`
	IsPVC                 bool   `json:"is_pvc"`
	IsBuildAndRunScript   bool   `json:"is_build_and_run_scrpt"`
	IsHPASetting          bool   `json:"is_hpa_setting"`
	IsEnvironmentVariable bool   `json:"is_environment_variable"`
	IsScheduler           bool   `json:"is_scheduler"`
}

type Token struct {
	ID         uint       `json:"id,omitempty"`
	CreatedAt  time.Time  `json:"createdat,omitempty"`
	UserId     uint64     `json:"user_id"`
	Name       string     `json:"name,omitempty"`
	Token      string     `json:"token,omitempty"`
	ExpiryDate *time.Time `json:"expiry_date,omitempty"`
}

type Group struct {
	ID             uint            `json:"id,omitempty"`
	CreatedAt      time.Time       `json:"createdat,omitempty"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Organization   *Organization   `json:"organization,omitempty"`
	OrganizationID uint64          `json:"organization_id"`
	Members        []*UserResponse `json:"members"`
}

type Organization struct {
	ID                 uint                   `json:"id"`
	CreatedAt          time.Time              `json:"createdat,omitempty"`
	Name               string                 `json:"name"`
	Description        string                 `json:"description"`
	Domain             string                 `json:"domain"`
	Image              string                 `json:"image"`
	User               *UserResponse          `json:"user"`
	UserID             uint64                 `json:"user_id"`
	OrganizationPlan   *OrganizationPlan      `json:"organization_plan"`
	OrganizationPlanID uint64                 `json:"organization_plan_id"`
	Plugins            []*Plugin              `json:"plugins"`
	Members            []*OrganizationMembers `json:"members"`
}
type UserRoleEnum int

const (
	Admin UserRoleEnum = 1 + iota
	Members
)

type OrganizationMembers struct {
	ID             uint          `json:"id,omitempty"`
	CreatedAt      time.Time     `json:"createdat,omitempty"`
	User           *UserResponse `json:"user"`
	UserID         uint64        `json:"user_id"`
	Organization   *Organization `json:"organization"`
	OrganizationID uint64        `json:"organization_id"`
	UserRole       UserRoleEnum  `json:"user_role"`
}

type OrganizationPlan struct {
	ID         uint      `json:"id,omitempty"`
	CreatedAt  time.Time `json:"createdat,omitempty"`
	Name       string    `json:"name"`
	Cluster    uint32    `json:"cluster"`
	Memory     uint32    `json:"memory"`
	Cores      uint32    `json:"cores"`
	NoOfUser   uint32    `json:"no_of_user"`
	Price      uint32    `json:"price"`
	Weight     uint32    `json:"weight"`
	Attributes string    `json:"attributes"`
	Active     bool      `json:"active"`
}
type Notification struct {
	Id                      int64  `json:"id"`
	OrganizationName        string `json:"organization_name"`
	OrganizationID          int64  `json:"organization_id"`
	ApplicationName         string `json:"application_name"`
	ApplicationID           int64  `json:"application_id"`
	EnvironmentName         string `json:"environment_name"`
	EnvironmentID           int64  `json:"environment_id"`
	EnvironmentResourceName string `json:"environment_resource_name"`
	EnvironmentResourceID   int64  `json:"environment_resource_id"`
	ProjectName             string `json:"project_name"`
	ProjectID               int64  `json:"project_id"`
	Scope                   string `json:"scope"`
	ScopeID                 int64  `json:"scope_id"`
	Action                  string `json:"action"`
	Type                    string `json:"type"`
	Body                    string `json:"body"`
	SendBy                  string `json:"send_by"`
	User                    int64  `json:"user"`
	UserName                string `json:"user_name"`
	TriggeredBy             int64  `json:"triggered_by"`
	TriggeredByName         string `json:"triggered_by_name"`
	Seen                    bool   `json:"seen"`
	CreateTimestamp         int64  `json:"create_timestamp"`
	UpdateTimestamp         int64  `json:"update_timestamp"`
}

type Plugin struct {
	ID               uint              `json:"id,omitempty"`
	CreatedAt        time.Time         `json:"createdat,omitempty"`
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	SourceUrl        string            `json:"source_url"`
	Image            string            `json:"image"`
	Active           bool              `json:"active"`
	SupportCi        bool              `json:"support_ci"`
	IsManagedService bool              `json:"is_managed_service"`
	MinCpu           uint64            `json:"min_cpu"`
	MinMemory        uint64            `json:"min_memory"`
	IsAddOn          bool              `json:"is_add_on"`
	AddOns           []*Plugin         `json:"add_ons"`
	Categories       []*PluginCategory `json:"plugin_category"`
	Attributes       string            `json:"attributes"`
	ServiceDetail    interface{}       `json:"service_detail"`
}

type PluginCategory struct {
	ID          uint      `json:"id,omitempty"`
	CreatedAt   time.Time `json:"createdat,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsAddOn     bool      `json:"is_add_on"`
	Plugins     []*Plugin `json:"plugins"`
}

type PluginVersion struct {
	ID          uint        `json:"id,omitempty"`
	CreatedAt   time.Time   `json:"createdat,omitempty"`
	Plugin      *Plugin     `json:"plugin,omitempty"`
	PluginID    uint64      `json:"plugin_id"`
	Version     string      `json:"version"`
	Url         string      `json:"url"`
	ChangeLogs  string      `json:"change_logs"`
	Attributes  string      `json:"attributes"`
	ReleaseDate time.Time   `json:"released"`
	Active      bool        `json:"active"`
	Versions    interface{} `json:"versions"`
}
type ClusterScopeEnum int

const (
	SHARED_SCOPE ClusterScopeEnum = iota
	ORGANIZATION_SCOPE
)

type Project struct {
	ID             uint             `json:"id,omitempty"`
	CreatedAt      time.Time        `json:"createdat,omitempty"`
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	ProjectCode    string           `json:"project_code"`
	Tags           interface{}      `json:"tags"`
	Region         string           `json:"region"`
	Active         bool             `json:"active"`
	ClusterScope   ClusterScopeEnum `json:"cluster_scope"`
	Subscription   *Subscription    `json:"subscription,omitempty"`
	SubscriptionID uint64           `json:"subscription_id"`
	Image          string           `json:"image,omitempty"`
	User           *UserResponse    `json:"user,omitempty"`
	UserID         uint64           `json:"user_id"`
	Organization   *Organization    `json:"organization,omitempty"`
	OrganizationId uint64           `json:"organization_id"`
	Logging        interface{}      `json:"logging"`
	Monitoring     interface{}      `json:"monitoring"`
	DedicatedLb    bool             `json:"dedicated_lb"`
	Variables      interface{}      `json:"variables"`
	SubsUpdated    *time.Time       `json:"subscription_updated"`
}

type ResourceUsedResponse struct {
	ID        uint      `json:"id,omitempty"`
	CreatedAt time.Time `json:"createdat,omitempty"`
	Memory    int64     `json:"memory"`
	Apps      int64     `json:"apps"`
	Core      int64     `json:"core"`
	Disk      float64   `json:"disk"`
}

type AvailableResourceResponse struct {
	ID        uint      `json:"id,omitempty"`
	CreatedAt time.Time `json:"createdat,omitempty"`
	Memory    int64     `json:"memory"`
	Core      int64     `json:"cpu"`
	Disk      float64   `json:"disk"`
}

type Resource struct {
	ID             uint          `json:"id,omitempty"`
	CreatedAt      time.Time     `json:"createdat,omitempty"`
	Name           string        `json:"name"`
	Cores          uint64        `json:"cores"`
	Memory         uint64        `json:"memory"`
	Active         bool          `json:"active"`
	Weight         uint32        `json:"weight"`
	Attributes     string        `json:"attributes"`
	Organization   *Organization `json:"organization,omitempty"`
	OrganizationID uint64        `json:"organization_id"`
}

type Operator struct {
	ID                        uint        `json:"id,omitempty"`
	CreatedAt                 time.Time   `json:"createdat,omitempty"`
	Name                      string      `json:"name"`
	PackageName               string      `gorm:"unique" json:"packageName"`
	DisplayName               string      `json:"displayName"`
	Provider                  string      `json:"provider"`
	ThumbUrl                  string      `json:"thumbUrl"`
	Version                   string      `json:"version"`
	VersionForCompare         string      `json:"versionForCompare"`
	K8sMinVersion             string      `json:"k8sMinVersion"`
	K8sMaxVersion             string      `json:"k8sMaxVersion"`
	Replaces                  string      `json:"replaces"`
	CapabilityLevel           string      `json:"capabilityLevel"`
	Repository                string      `json:"repository"`
	Description               string      `json:"description"`
	ContainerImage            string      `json:"containerImage"`
	Channel                   string      `json:"channel"`
	GlobalOperator            bool        `json:"globalOperator"`
	Active                    bool        `json:"active"`
	Channels                  interface{} `json:"channels"`
	Links                     interface{} `json:"links"`
	CustomResourceDefinitions interface{} `json:"customResourceDefinitions"`
	Categories                interface{} `json:"categories"`
	Keywords                  interface{} `json:"keywords"`
	OperaterCreatedAt         string      `json:"createdAt"`
}

type OperatorRequest struct {
	ID                  uint            `json:"id,omitempty"`
	CreatedAt           time.Time       `json:"createdat,omitempty"`
	PackageName         string          `json:"package_name"`
	CsvValue            string          `json:"csv_value"`
	Cluster             *Cluster        `json:"cluster,omitempty"`
	ClusterRequest      *ClusterRequest `json:"cluster_request,omitempty"`
	ClusterRequestID    uint64          `json:"cluster_request_id"`
	InstallPlanApproval string          `json:"install_plan_approval"`
	Channel             string          `json:"channel"`
	GlobalOperator      bool            `json:"global_operator"`
	InstallationMode    string          `json:"installation_mode"`
	OperatorDetails     interface{}     `json:"operator_details"`
	Organization        *Organization   `json:"organization,omitempty"`
	OrganizationID      uint64          `json:"organization_id"`
}
type Storage struct {
	ID            uint          `json:"id,omitempty"`
	CreatedAt     time.Time     `json:"createdat,omitempty"`
	EnvironmentID uint          `json:"environment_id"`
	Name          string        `json:"name"`
	VolumeName    *string       `json:"volume_name"`
	AccessModes   *string       `json:"access_modes"`
	MountPath     *string       `json:"mount_path"`
	StorageType   *string       `json:"storage_type"`
	AttachedTo    *string       `json:"attached_to"`
	Capacity      uint64        `json:"capacity"`
	UsedStorage   *uint64       `json:"used_storage"`
	Active        bool          `json:"active"`
	User          *UserResponse `json:"user,omitempty"`
	UserID        uint          `json:"user_id"`
}

type Activity struct {
	ID             uint          `json:"id,omitempty"`
	CreatedAt      time.Time     `json:"createdat,omitempty"`
	User           *UserResponse `json:"user"`
	UserID         uint          `json:"user_id"`
	Project        *Project      `json:"project"`
	ProjectID      uint          `json:"project_id"`
	Application    *Application  `json:"application"`
	ApplicationID  uint          `json:"application_id"`
	Environment    *Environment  `json:"environment"`
	EnvironmentID  uint          `json:"environment_id"`
	OrganizationID uint          `json:"organization_id"`
	Action         string        `json:"action"`
	Module         string        `json:"module"`
	Active         bool          `json:"active"`
	Remarks        string        `json:"remarks"`
	Extras         string        `json:"extras"`
}

type HelmEnvironment struct {
	ID               uint          `json:"id,omitempty"`
	CreatedAt        time.Time     `json:"createdat,omitempty"`
	Name             string        `json:"name"`
	Application      *Application  `json:"application" `
	ApplicationID    uint64        `json:"application_id"`
	ChartVersion     *ChartVersion `json:"chart_version"`
	ChartVersionID   string        `json:"chart_version_id"`
	Values           string        `json:"values"`
	Version          interface{}   `json:"version"`
	Active           bool          `json:"active"`
	ApplyImmediately bool          `json:"apply_immediately"`
	CiRequest        *CIRequest    `json:"ci_request"`
	Action           string        `json:"action"`
	Scripts          interface{}   `json:"scripts"`
	Schedules        interface{}   `json:"schedules"`
}

type CiConfig struct {
	ID                  uint         `json:"id,omitempty"`
	CreatedAt           time.Time    `json:"createdat,omitempty"`
	WebhookUrl          string       `json:"webhook_url"`
	SlackWebhookUrl     string       `json:"slack_webhook_url"`
	Emails              string       `json:"emails"`
	HookId              string       `json:"hook_id"`
	Events              string       `json:"events"`
	EmailNotification   bool         `json:"email_notification"`
	SlackNotification   bool         `json:"slack_notification"`
	WebhookNotification bool         `json:"webhook_notification"`
	Environment         *Environment `json:"environment,omitempty"`
	EnvironmentID       uint         `json:"environment_id"`
}

type RepositoryImage struct {
	ID            uint      `json:"id,omitempty"`
	CreatedAt     time.Time `json:"createdat,omitempty"`
	Name          string
	Repository    string
	Tag           string
	CommitMessage map[string]string
}

type Subscription struct {
	ID             uint          `json:"id,omitempty"`
	CreatedAt      time.Time     `json:"createdat,omitempty"`
	Name           string        `json:"name"`
	Apps           uint32        `json:"apps"`
	DiskSpace      uint32        `json:"disk_space"`
	Memory         uint32        `json:"memory"`
	Cores          uint32        `json:"cores"`
	DataTransfer   uint32        `json:"data_transfer"`
	Price          uint32        `json:"price"`
	Weight         uint32        `json:"weight"`
	Active         bool          `json:"active"`
	Organization   *Organization `json:"organization,omitempty"`
	OrganizationID uint64        `json:"organization_id"`
	CiBuild        uint32        `json:"ci_build"`
	Attributes     string        `json:"attributes"`
	CronJob        uint64        `json:"cron_job"`
	Backups        uint64        `json:"backups"`
	ResourceList   interface{}   `json:"resource_list"`
	LoadBalancer   uint64        `json:"load_balancer"`
	PriceList      interface{}   `json:"price_list"`
	Validity       uint64        `json:"validity"`
}

type User struct {
	ID             uint          `json:"id,omitempty"`
	CreatedAt      time.Time     `json:"createdat,omitempty"`
	FirstName      string        `json:"first_name,omitempty"`
	LastName       string        `json:"last_name"`
	Email          string        `json:"email,omitempty"`
	Company        string        `json:"company,omitempty"`
	Designation    string        `json:"designation,omitempty"`
	EmailVerified  bool          `json:"email_verified"`
	Active         bool          `json:"active"`
	IsAdmin        bool          `json:"is_admin,omitempty"`
	Image          string        `json:"image,omitempty"`
	AddressUpdated bool          `json:"address_updated,omitempty"`
	Quotas         models.Quotas `json:"quotas"`
	UsedDemo       bool          `json:"used_demo"`
	Reference      string        `json:"reference"`
}

type UserResponse struct {
	ID        uint   `json:"id,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name"`
	Email     string `json:"email,omitempty"`
	Image     string `json:"image,omitempty"`
}

type Backup struct {
	ID        uint      `json:"id,omitempty"`
	CreatedAt time.Time `json:"createdat,omitempty"`
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Snapshot  string    `json:"snapshot"`
	Preserved bool      `json:"preserved"`
	Created   int64     `json:"created"`
	Restored  int64     `json:"restored"`
	Duration  int64     `json:"duration"`
	Status    string    `json:"status"`
}

type UserLogin struct {
	ID        uint      `json:"id,omitempty"`
	CreatedAt time.Time `json:"createdat,omitempty"`
	Email     string
	Password  string
}

type UserLoginResponse struct {
	ID        uint      `json:"id,omitempty"`
	CreatedAt time.Time `json:"createdat,omitempty"`
	Token     string
	User      User
}

type UserRegister struct {
	ID          uint      `json:"id,omitempty"`
	CreatedAt   time.Time `json:"createdat,omitempty"`
	FirstName   string
	LastName    string
	Email       string
	Password    string
	Designation string
	Company     string
}

type ExternalLogin struct {
	ID          uint      `json:"id,omitempty"`
	CreatedAt   time.Time `json:"createdat,omitempty"`
	ServiceName string
	ServiceCode string
}

type EmailRequest struct {
	ID        uint      `json:"id,omitempty"`
	CreatedAt time.Time `json:"createdat,omitempty"`
	Email     string
}

type GetUsersRequest struct {
	ID            uint      `json:"id,omitempty"`
	CreatedAt     time.Time `json:"createdat,omitempty"`
	Search        string
	Page          int
	Size          int
	SortColumn    string `json:"sort-column"`
	SortDirection string `json:"sort-direction"`
}

type ChangePassword struct {
	ID             uint      `json:"id,omitempty"`
	CreatedAt      time.Time `json:"createdat,omitempty"`
	Password       string
	NewPassword    string
	RetypePassword string
}
type ResetPassword struct {
	ID             uint      `json:"id,omitempty"`
	CreatedAt      time.Time `json:"createdat,omitempty"`
	Token          string
	NewPassword    string
	RetypePassword string
}

type PackageRequestBody []PackageRequest
type PackageRequest struct {
	ID          uint                     `json:"id,omitempty"`
	CreatedAt   time.Time                `json:"createdat,omitempty"`
	Name        string                   `json:"name"`
	Namespace   string                   `json:"namespace"`
	Chart       string                   `json:"chart"`
	RequiredDNS bool                     `json:"required_dns"`
	Set         []map[string]interface{} `json:"set,omitempty"`
	Needs       []string                 `json:"needs,omitempty"`
}

type PackageYaml struct {
	ID          uint                     `json:"id,omitempty"`
	CreatedAt   time.Time                `json:"createdat,omitempty"`
	Name        string                   `json:"name"`
	Namespace   string                   `json:"namespace,omitempty"`
	Chart       string                   `json:"chart"`
	RequiredDNS bool                     `json:"-"`
	Set         []map[string]interface{} `json:"set,omitempty"`
	Needs       []string                 `json:"needs,omitempty"`
}

type GitUser struct {
	ID              uint          `json:"id,omitempty"`
	CreatedAt       time.Time     `json:"createdat,omitempty"`
	UserID          uint64        `json:"user_id"`
	GitUserID       uint64        `json:"git_user_id"`
	ServiceUserName string        `json:"service_user_name"`
	ServiceName     string        `json:"service_name"`
	ServiceUrl      string        `json:"service_url"`
	User            *UserResponse `json:"user"`
	Region          string        `json:"region"`
	Active          bool          `json:"active"`
	IsOauth         bool          `json:"is_oauth"`
}

type KVResponse struct {
	Key   string `json:"name"`
	Value string `json:"value"`
}

type RenameRequest struct {
	Name string `json:"name"`
}

type SecurityScanRequest struct {
	// plugin name to use for scanning
	Name string `json:"name"`
}
