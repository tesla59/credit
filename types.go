package main

type User struct {
	User_id    int `gorm:"primaryKey;autoIncrement:false"`
	Chat_id    int64 `gorm:"primaryKey;autoIncrement:false"`
	First_name string
	Last_name  string
	Username   string
	Credit     int
}