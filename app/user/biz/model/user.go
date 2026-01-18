package model

import (
	"context"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Name         string `gorm:"type:varchar(50);not null;index"`
	PasswordHash string `gorm:"type:varchar(100);not null"`
	UserAccount  string `gorm:"type:varchar(50);not null;uniqueIndex"`

	UserAvatar  string `gorm:"type:varchar(100);not null;default:''"`
	UserProfile string `gorm:"type:varchar(100);default:''"`
	UserRole    string `gorm:"type:varchar(10);not null;default:'user'"`
}

func (User) TableName() string {
	return "user"
}

type UserQuery struct {
	ctx context.Context
	db  *gorm.DB
}

func NewUserQuery(ctx context.Context, db *gorm.DB) *UserQuery {
	return &UserQuery{
		ctx: ctx,
		db:  db,
	}
}
func (q *UserQuery) GetUserById(id int) (User, error) {
	user := User{}
	err := q.db.WithContext(q.ctx).Model(&User{}).Where("id = ?", id).First(&user).Error
	return user, err
}
func (q *UserQuery) GetUserByAccount(account string) (User, error) {
	user := User{}
	err := q.db.WithContext(q.ctx).Model(&User{}).Where("user_account = ?", account).First(&user).Error
	return user, err
}
func (q *UserQuery) UpdateUser(id uint, user User) error {
	err := q.db.WithContext(q.ctx).Model(&User{}).Where("id = ?", id).Updates(user).Error
	return err
}
func (q *UserQuery) CreateUser(user User) (User, error) {
	err := q.db.WithContext(q.ctx).Model(&User{}).Create(&user).Error
	return user, err
}
func (q *UserQuery) DeleteUser(id int) error {
	err := q.db.WithContext(q.ctx).Model(&User{}).Where("id = ?", id).Delete(&User{}).Error
	return err
}
func (q *UserQuery) ListUser(page uint32, username, useraccount string, page_size uint32) ([]User, error) {
	var users []User
	err := q.db.WithContext(q.ctx).Model(&User{}).Where("name like ? and user_account like ?", username+"%", useraccount+"%").Limit(int(page_size)).Offset(int(page_size * (page - 1))).Find(&users).Error
	return users, err
}
func (q *UserQuery) CountUser(username, useraccount string) (int64, error) {
	var count int64
	err := q.db.WithContext(q.ctx).Model(&User{}).Where("name like ? and user_account like ?", username+"%", useraccount+"%").Count(&count).Error
	return count, err
}
