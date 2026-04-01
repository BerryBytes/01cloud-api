package models

import "github.com/jinzhu/gorm"

type LoadBalancerInterface interface {
	Save(db *gorm.DB, data *LoadBalancer) (*LoadBalancer, error)
	Find(db *gorm.DB, pid uint64) (*LoadBalancer, error)
	FindAllByProject(db *gorm.DB, pid uint64) (*[]LoadBalancer, error)
	Update(db *gorm.DB, data *LoadBalancer) (*LoadBalancer, error)
	Delete(db *gorm.DB, uid uint64) (int64, error)
}
type LoadBalancerType struct {
}

func NewLoadBalancer() LoadBalancerInterface {
	return &LoadBalancerType{}
}
