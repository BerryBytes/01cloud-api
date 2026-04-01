package backup

import pb "01cloud-api/api/backup/proto"

type Backup struct {
	Id              int64   `json:"id"`
	Namespace       string  `json:"namespace"`
	Name            string  `json:"name"`
	EnvironmentType string  `json:"environment_type"`
	Snapshot        string  `json:"snapshot"`
	Preserved       bool    `json:"preserved"`
	Label           []Label `json:"label"`
	Created         int64   `json:"created"`
	Completed       int64   `json:"completed"`
	Restored        int64   `json:"restored"`
	Duration        int64   `json:"duration"`
	Status          string  `json:"status"`
	TriggeredBy     int64   `json:"triggered_by"`
	TriggeredByName string  `json:"triggered_by_name"`
	Remarks         string  `json:"remarks"`
}
type NotifyBackup struct {
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	Resource        string `json:"resource"`
	Status          string `json:"status"`
	Message         string `json:"message"`
	Emails          string `json:"emails"`
	TriggeredBy     string `json:"triggered_by"`
	TriggeredByName string `json:"triggered_by_name"`
	Type            string `json:"type"`
	Created         int64  `json:"created"`
	Duration        int64  `json:"duration"`
	Completed       int64  `json:"completed"`
}

type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (m *Backup) FromGrpc(pb *pb.BackupMessage) {
	m.Id = pb.Id
	m.Namespace = pb.Namespace
	m.Name = pb.Name
	m.EnvironmentType = pb.EnvironmentType
	m.Snapshot = pb.Snapshot
	m.Preserved = pb.Preserved
	m.Created = pb.Created
	m.Restored = pb.Restored
	m.Status = pb.Status
	m.Duration = pb.Duration
	m.TriggeredBy = pb.TriggeredBy
	m.TriggeredByName = pb.TriggeredByName
	m.Remarks = pb.Remarks
	for _, data := range pb.Label {
		label := Label{
			Key:   data.Key,
			Value: data.Value,
		}
		m.Label = append(m.Label, label)
	}
}

func (pbm *Backup) ToGrpc(m *pb.BackupMessage) {
	m.Id = pbm.Id
	m.Namespace = pbm.Namespace
	m.Name = pbm.Name
	m.EnvironmentType = pbm.EnvironmentType
	m.Snapshot = pbm.Snapshot
	m.Preserved = pbm.Preserved
	m.Created = pbm.Created
	m.Restored = pbm.Restored
	m.Status = pbm.Status
	m.Duration = pbm.Duration
	m.TriggeredBy = pbm.TriggeredBy
	m.TriggeredByName = pbm.TriggeredByName
	m.Remarks = pbm.Remarks
	for _, data := range pbm.Label {
		label := pb.Label{
			Key:   data.Key,
			Value: data.Value,
		}
		m.Label = append(m.Label, &label)
	}
}

func (pb *Label) ToGrpc(m *pb.Label) {
	m.Key = pb.Key
	m.Value = pb.Value
}
