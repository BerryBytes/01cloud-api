package notifications

import (
	notification "01cloud-api/api/notifications/proto"
	pb "01cloud-api/api/notifications/proto"
)

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

type BasicInfoNotification struct {
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
}

type SendEmailMessage struct {
	User    []string `json:"user"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	Caption string   `json:"caption"`
	Link    string   `json:"link"`
}

func (m *Notification) FromGrpc(pb *pb.Message) {
	m.Body = pb.Body
	m.User = pb.User
	m.UserName = pb.UserName
	m.SendBy = pb.SendBy
	m.Scope = pb.Scope
	m.ScopeID = pb.ScopeId
	m.Action = pb.Action
	m.Type = pb.Type
	m.ApplicationID = pb.ApplicationId
	m.ApplicationName = pb.ApplicationName
	m.EnvironmentID = pb.EnvironmentId
	m.EnvironmentName = pb.EnvironmentName
	m.EnvironmentResourceID = pb.EnvironmentResourceId
	m.EnvironmentResourceName = pb.EnvironmentResourceName
	m.ProjectID = pb.ProjectId
	m.ProjectName = pb.ProjectName
	m.OrganizationID = pb.OrganizationId
	m.TriggeredBy = pb.TriggeredBy
	m.TriggeredByName = pb.TriggeredByName
	m.OrganizationName = pb.OrganizationName
	m.UpdateTimestamp = pb.UpdateTimestamp
	m.CreateTimestamp = pb.CreateTimestamp
	m.Seen = pb.Seen
	m.Id = pb.Id
}

func (m *Notification) ToGrpc(pb *pb.Message) {
	pb.Body = m.Body
	pb.User = m.User
	pb.UserName = m.UserName
	pb.SendBy = m.SendBy
	pb.Scope = m.Scope
	pb.ScopeId = m.ScopeID
	pb.Action = m.Action
	pb.Type = m.Type
	pb.ApplicationName = m.ApplicationName
	pb.ApplicationId = m.ApplicationID
	pb.EnvironmentId = m.EnvironmentID
	pb.EnvironmentName = m.EnvironmentName
	pb.EnvironmentResourceId = m.EnvironmentResourceID
	pb.EnvironmentResourceName = m.EnvironmentResourceName
	pb.OrganizationId = m.OrganizationID
	pb.OrganizationName = m.OrganizationName
	pb.ProjectId = m.ProjectID
	pb.ProjectName = m.ProjectName
	pb.TriggeredBy = m.TriggeredBy
	pb.TriggeredByName = m.TriggeredByName
	pb.UpdateTimestamp = m.UpdateTimestamp
	pb.CreateTimestamp = m.CreateTimestamp
	pb.Seen = m.Seen
	pb.Id = m.Id
}

func (m *SendEmailMessage) ToGrpc(pb *notification.EmailMessage) {
	pb.User = m.User
	pb.Subject = m.Subject
	pb.Body = m.Body
	pb.Link = m.Link
	pb.Caption = m.Caption
}
