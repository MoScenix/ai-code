package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
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

type AppProQuery struct {
	q      *AppQuery
	rdb    *redis.Client
	prefix string
}

func NewAppProQuery(ctx context.Context, db *gorm.DB, rdb *redis.Client) *AppProQuery {
	return &AppProQuery{
		q:      NewAppQuery(ctx, db),
		rdb:    rdb,
		prefix: "ai-code",
	}
}

func (p *AppProQuery) keyApp(id uint) string {
	return fmt.Sprintf("%s_app_%d", p.prefix, id)
}

func (p *AppProQuery) keyUserApps(userId uint) string {
	return fmt.Sprintf("%s_user_apps_%d", p.prefix, userId)
}

func (p *AppProQuery) GetAppById(id uint) (App, error) {
	if p.rdb != nil {
		if val, err := p.rdb.Get(p.q.ctx, p.keyApp(id)).Result(); err == nil && val != "" {
			var a App
			if json.Unmarshal([]byte(val), &a) == nil {
				return a, nil
			}
		}
	}

	a, err := p.q.GetAppById(id)
	if err != nil {
		return App{}, err
	}

	if p.rdb != nil {
		if b, e := json.Marshal(a); e == nil {
			_ = p.rdb.Set(p.q.ctx, p.keyApp(id), b, time.Hour).Err()
		}
	}
	return a, nil
}
func (p *AppProQuery) GetAppsByUserId(userId uint) ([]App, error) {
	if p.rdb != nil {
		if val, err := p.rdb.Get(p.q.ctx, p.keyUserApps(userId)).Result(); err == nil && val != "" {
			var list []App
			if json.Unmarshal([]byte(val), &list) == nil {
				return list, nil
			}
		}
	}

	list, err := p.q.GetAppsByUserId(userId)
	if err != nil {
		return nil, err
	}

	if p.rdb != nil {
		if b, e := json.Marshal(list); e == nil {
			_ = p.rdb.Set(p.q.ctx, p.keyUserApps(userId), b, time.Hour).Err()
		}
	}
	return list, nil
}
func (p *AppProQuery) UpdateApp(id uint, app App) error {
	err := p.rdb.Del(p.q.ctx, p.keyApp(id)).Err()
	if err != nil {
		return err
	}
	err = p.rdb.Del(p.q.ctx, p.keyUserApps(app.UserId)).Err()
	if err != nil {
		return err
	}
	return p.q.UpdateApp(id, app)
}

func (p *AppProQuery) CreateApp(app App) (App, error) {
	err := p.rdb.Del(p.q.ctx, p.keyUserApps(app.UserId)).Err()
	if err != nil {
		return App{}, err
	}
	err = p.rdb.Del(p.q.ctx, p.keyUserApps(app.UserId)).Err()
	if err != nil {
		return App{}, err
	}
	created, err := p.q.CreateApp(app)
	if err != nil {
		return App{}, err
	}
	return created, nil
}

func (p *AppProQuery) DeleteApp(id uint) error {
	app, err := p.q.GetAppById(id)
	if err != nil {
		return err
	}
	err = p.rdb.Del(p.q.ctx, p.keyApp(id)).Err()
	if err != nil {
		return err
	}
	err = p.rdb.Del(p.q.ctx, p.keyUserApps(app.UserId)).Err()
	if err != nil {
		return err
	}
	return p.q.DeleteApp(id)
}

func (p *AppProQuery) ListApp(page uint32, userId uint, name string, pageSize uint32) ([]App, error) {
	return p.q.ListApp(page, userId, name, pageSize)
}
func (p *AppProQuery) CountApp(userId uint, name string) (int64, error) {
	return p.q.CountApp(userId, name)
}
