package main

type User struct {
	UserID    int   `gorm:"primaryKey;autoIncrement:false"`
	ChatID    int64 `gorm:"primaryKey;autoIncrement:false"`
	FirstName string
	LastName  string
	UserName  string
	Credit    int
}
