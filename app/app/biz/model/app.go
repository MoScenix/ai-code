package model

import "gorm.io/gorm"

type App struct {
	gorm.Model
	Name       string `gorm:"type:varchar(100);Index" json:"name"`
	Cover      string `gorm:"type:varchar(255)" json:"cover"`
	InitPrompt string `gorm:"type:text" json:"initPrompt"`
	Deploykey  string `gorm:"type:varchar(100)" json:"deploykey"`
	UserId     uint   `gorm:"type:int;Index" json:"userId"`
	Priority   int    `gorm:"type:int" json:"priority"`
}
