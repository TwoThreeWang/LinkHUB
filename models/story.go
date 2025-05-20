package models

import "time"

type Story struct {
	ID          int       `json:"id" gorm:"column:id;primaryKey"`
	Title       string    `json:"title" gorm:"column:title;type:varchar(100)"`
	URL         string    `json:"url" gorm:"column:url;type:varchar(1024)"`
	Time        time.Time `json:"time" gorm:"column:time"`
	By          string    `json:"by" gorm:"column:by;type:varchar(50)"`
	Descendants int       `json:"descendants" gorm:"column:descendants"`
	Content     string    `json:"content" gorm:"column:content;type:text"`
	Summary     string    `json:"summary" gorm:"column:summary;type:text"`
}