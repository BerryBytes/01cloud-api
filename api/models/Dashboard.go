package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type IDashboard interface {
	Dashboard(db *gorm.DB) (Count, error)
}

type DashboardRepo struct {
}

type dataList struct {
	Month time.Time `json:"month,omitempty"`
	Total int       `json:"total"`
}

type transaction struct {
	Month   time.Time `json:"month"`
	Payable float64   `json:"payable"`
	Paid    float64   `json:"paid"`
}

type countData struct {
	Total           int           `json:"total"`
	CurrentMonth    int           `json:"current_month"`
	LastMonth       int           `json:"last_month"`
	Data            []dataList    `json:"data,omitempty"`
	TransactionData []transaction `json:"transaction_data,omitempty"`
}

type resource struct {
	Memory uint64 `json:"memory"`
	Core   uint64 `json:"core"`
}

type subscriptionList []struct {
	Name    string `json:"name"`
	Count   int    `json:"count"`
	Price   int    `json:"price"`
	Revenue int    `json:"revenue"`
}

type Count struct {
	Projects      countData        `json:"projects"`
	Applications  countData        `json:"applications"`
	Environments  countData        `json:"environments"`
	Users         countData        `json:"users"`
	Clusters      countData        `json:"clusters"`
	Plugins       countData        `json:"plugins"`
	Resources     resource         `json:"resources"`
	Transactions  countData        `json:"transactions"`
	Subscriptions subscriptionList `json:"subscriptions"`
}

func NewDashboard() IDashboard {
	return &DashboardRepo{}
}

func (d *DashboardRepo) Dashboard(db *gorm.DB) (Count, error) {
	interval := "11 months"
	count := Count{}
	var paymentdue, balance float64
	db.Raw("select subscriptions.name,subscriptions.price, count(*) as count, sum(price) as revenue from subscriptions join projects on projects.subscription_id = subscriptions.id where projects.deleted_at is NULL and subscriptions.organization_id=0 group by subscriptions.name,subscriptions.price Order by price asc").Scan(&count.Subscriptions)
	db.Raw("select date_trunc('month', date) as month,sum(credit) as payable,sum(debit) as paid from payment_histories where  date>=current_date - ?::interval group by 1 order by month", "12 months").Scan(&count.Transactions.TransactionData)
	db.Raw("select date_trunc('month', created_at) as month, count(*) as total from projects where created_at>=current_date -  ?::interval group by 1 order by month", interval).Scan(&count.Projects.Data)
	db.Model(&Project{}).Count(&count.Projects.Total)
	db.Model(&Project{}).Where("date_trunc('month', created_at) = date_trunc('month', current_date)").Count(&count.Projects.CurrentMonth)
	db.Model(&Project{}).Where("created_at >= date_trunc('month', current_date - interval '1' month) and created_at < date_trunc('month', current_date)").Count(&count.Projects.LastMonth)
	db.Model(&Application{}).Count(&count.Applications.Total)
	db.Model(&Application{}).Where("date_trunc('month', created_at) = date_trunc('month', current_date)").Count(&count.Applications.CurrentMonth)
	db.Model(&Application{}).Where("created_at >= date_trunc('month', current_date - interval '1' month) and created_at < date_trunc('month', current_date)").Count(&count.Applications.LastMonth)
	db.Raw("select date_trunc('month', created_at) as month, count(*) as total from applications where created_at>=current_date - ?::interval group by 1 order by month", interval).Scan(&count.Applications.Data)
	db.Model(&Environment{}).Count(&count.Environments.Total)
	db.Model(&Environment{}).Where("date_trunc('month', created_at) = date_trunc('month', current_date)").Count(&count.Environments.CurrentMonth)
	db.Model(&Environment{}).Where("created_at >= date_trunc('month', current_date - interval '1' month) and created_at < date_trunc('month', current_date)").Count(&count.Environments.LastMonth)
	db.Raw("select date_trunc('month', created_at) as month, count(*) as total from environments where created_at>=current_date - ?::interval group by 1 order by month", interval).Scan(&count.Environments.Data)
	db.Model(&Cluster{}).Count(&count.Clusters.Total)
	db.Model(&Cluster{}).Where("date_trunc('month', created_at) = date_trunc('month', current_date)").Count(&count.Clusters.CurrentMonth)
	db.Model(&Cluster{}).Where("created_at >= date_trunc('month', current_date - interval '1' month) and created_at < date_trunc('month', current_date)").Count(&count.Clusters.LastMonth)
	db.Model(&User{}).Count(&count.Users.Total)
	db.Model(&User{}).Where("date_trunc('month', created_at) = date_trunc('month', current_date)").Count(&count.Users.CurrentMonth)
	db.Model(&User{}).Where("created_at >= date_trunc('month', current_date - interval '1' month) and created_at < date_trunc('month', current_date)").Count(&count.Users.LastMonth)
	db.Raw("select date_trunc('month', created_at) as month, count(*) as total from users where created_at>=current_date - ?::interval group by 1 order by month", interval).Scan(&count.Users.Data)
	db.Model(&Plugin{}).Count(&count.Plugins.Total)
	db.Model(&Plugin{}).Where("date_trunc('month', created_at) = date_trunc('month', current_date)").Count(&count.Plugins.CurrentMonth)
	db.Model(&Plugin{}).Where("created_at >= date_trunc('month', current_date - interval '1' month) and created_at < date_trunc('month', current_date)").Count(&count.Plugins.LastMonth)
	db.Raw("select sum(memory) as memory,sum(cores) as core from resources join environments on environments.resource_id=resources.id where environments.deleted_at is null and resources.organization_id=0 ").First(&count.Resources)
	db.Raw("select sum(balance) as paymentdue from payment_histories where date_trunc('month',current_date)=date_trunc('month', date)  and balance<0").Scan(&paymentdue)
	db.Raw("select sum(balance) as balance from payment_histories where date_trunc('month',current_date)=date_trunc('month', date)  and balance>0").Scan(&balance)
	// db.Raw("select sum(credit) as credit,sum(debit) as debit from payment_histories where date_trunc('month',current_date)= date_trunc('month', date)").Scan(&count.Transactions.CurrentMonth)
	// db.Raw("select sum(credit) as credit,sum(debit) as debit from payment_histories where created_at >= date_trunc('month', current_date - interval '1' month) and created_at < date_trunc('month', current_date)").Scan(&count.Transactions.LastMonth)
	return count, nil
}
