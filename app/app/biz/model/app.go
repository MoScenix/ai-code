package model

import (
	"context"

	"gorm.io/gorm"
)

type App struct {
	gorm.Model
	Name         string    `gorm:"type:varchar(100);Index" json:"name"`
	Cover        string    `gorm:"type:varchar(255)" json:"cover"`
	InitPrompt   string    `gorm:"type:text" json:"initPrompt"`
	Deploykey    string    `gorm:"type:varchar(100)" json:"deploykey"`
	UserId       uint      `gorm:"type:int;Index" json:"userId"`
	Priority     int       `gorm:"type:int" json:"priority"`
	DeployedTime string    `gorm:"type:varchar(50)" json:"deployedTime"`
	Messages     []Message `gorm:"foreignKey:AppId" json:"messages"`
}
type AppQuery struct {
	ctx context.Context
	db  *gorm.DB
}

func NewAppQuery(ctx context.Context, db *gorm.DB) *AppQuery {
	return &AppQuery{
		ctx: ctx,
		db:  db,
	}
}

func (q *AppQuery) GetAppById(id uint) (App, error) {
	app := App{}
	err := q.db.WithContext(q.ctx).
		Model(&App{}).
		Where("id = ?", id).
		First(&app).Error
	return app, err
}

func (q *AppQuery) GetAppsByUserId(userId uint) ([]App, error) {
	var apps []App
	err := q.db.WithContext(q.ctx).
		Model(&App{}).
		Where("user_id = ?", userId).
		Order("priority desc, id desc").
		Find(&apps).Error
	return apps, err
}

func (q *AppQuery) UpdateApp(id uint, app App) error {
	return q.db.WithContext(q.ctx).
		Model(&App{}).
		Where("id = ?", id).
		Updates(app).Error
}

func (q *AppQuery) CreateApp(app App) (App, error) {
	err := q.db.WithContext(q.ctx).
		Model(&App{}).
		Create(&app).Error
	return app, err
}

func (q *AppQuery) DeleteApp(id uint) error {
	return q.db.WithContext(q.ctx).
		Model(&App{}).
		Where("id = ?", id).
		Delete(&App{}).Error
}

func (q *AppQuery) ListApp(page uint32, userId uint, name string, pageSize uint32) ([]App, error) {
	var apps []App

	tx := q.db.WithContext(q.ctx).Model(&App{})

	if userId != 0 {
		tx = tx.Where("user_id = ?", userId)
	}
	if name != "" {
		tx = tx.Where("name LIKE ?", name+"%")
	}

	err := tx.Order("priority desc, id desc").
		Limit(int(pageSize)).
		Offset(int(pageSize * (page - 1))).
		Find(&apps).Error

	return apps, err
}
func (q *AppQuery) CountApp(userId uint, name string) (int64, error) {
	var count int64

	tx := q.db.WithContext(q.ctx).Model(&App{})

	if userId != 0 {
		tx = tx.Where("user_id = ?", userId)
	}
	if name != "" {
		tx = tx.Where("name LIKE ?", name+"%")
	}

	err := tx.Count(&count).Error
	return count, err
}
