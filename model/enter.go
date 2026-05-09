package model

import (
	"gorm.io/gorm"
)

type Model struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt Time
	UpdatedAt Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ModelC struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt Time
}

type ModelCU struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt Time
	UpdatedAt Time
}
