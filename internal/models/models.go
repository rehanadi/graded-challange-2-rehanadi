package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        string    `gorm:"type:uuid;primary_key" json:"id"`
	Username  string    `gorm:"unique;not null" json:"username"`
	Password  string    `gorm:"not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

type Book struct {
	ID            string    `gorm:"type:uuid;primary_key" json:"id"`
	Title         string    `gorm:"not null" json:"title"`
	Author        string    `gorm:"not null" json:"author"`
	PublishedDate time.Time `json:"published_date"`
	Status        string    `gorm:"not null;default:'Available'" json:"status"`
	UserID        *string   `gorm:"type:uuid" json:"user_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (b *Book) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}

type BorrowedBook struct {
	ID           string     `gorm:"type:uuid;primary_key" json:"id"`
	BookID       string     `gorm:"type:uuid;not null" json:"book_id"`
	Book         Book       `gorm:"foreignKey:BookID" json:"book"`
	UserID       string     `gorm:"type:uuid;not null" json:"user_id"`
	User         User       `gorm:"foreignKey:UserID" json:"user"`
	BorrowedDate time.Time  `gorm:"not null" json:"borrowed_date"`
	ReturnDate   *time.Time `json:"return_date"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (bb *BorrowedBook) BeforeCreate(tx *gorm.DB) error {
	if bb.ID == "" {
		bb.ID = uuid.New().String()
	}
	return nil
}
