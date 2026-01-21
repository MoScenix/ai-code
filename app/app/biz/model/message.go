package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	Content string `gorm:"type:text" json:"content"`
	Role    string `gorm:"type:varchar(50)" json:"role"`
	AppId   uint   `gorm:"type:int;index" json:"appId"`
}
type MessageQuery struct {
	ctx context.Context
	db  *gorm.DB
}

func NewMessageQuery(ctx context.Context, db *gorm.DB) *MessageQuery {
	return &MessageQuery{
		ctx: ctx,
		db:  db,
	}
}
func (q *MessageQuery) GetMessageById(id uint) (Message, error) {
	msg := Message{}
	err := q.db.WithContext(q.ctx).
		Model(&Message{}).
		Where("id = ?", id).
		First(&msg).Error
	return msg, err
}
func (q *MessageQuery) CreateMessage(msg Message) (Message, error) {
	err := q.db.WithContext(q.ctx).
		Model(&Message{}).
		Create(&msg).Error
	return msg, err
}

func (q *MessageQuery) DeleteMessageById(id uint) error {
	return q.db.WithContext(q.ctx).
		Model(&Message{}).
		Where("id = ?", id).
		Delete(&Message{}).Error
}
func (q *MessageQuery) ListMessagesByAppId(appId uint, limit int, lastCreateTime *time.Time) ([]Message, error) {
	var msgs []Message

	db := q.db.WithContext(q.ctx).
		Model(&Message{}).
		Where("app_id = ?", appId)
	if lastCreateTime != nil && !lastCreateTime.IsZero() {
		db = db.Where("created_at < ?", *lastCreateTime)
	}

	err := db.
		Order("created_at desc").
		Limit(limit).
		Find(&msgs).Error

	return msgs, err
}

func (q *MessageQuery) Count(appId uint) (int64, error) {
	var count int64
	err := q.db.WithContext(q.ctx).
		Model(&Message{}).
		Where("app_id = ?", appId).
		Count(&count).Error
	return count, err
}
