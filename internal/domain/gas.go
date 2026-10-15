package domain

import "time"

type Gas struct {
	ID       string    `json:"id" db:"id"`
	Date     time.Time `json:"date" db:"date" validate:"required"`
	Supplier string    `json:"supplier" db:"supplier" validate:"required"`
	Part     string    `json:"part" db:"part" validate:"required"`
	MQ137    int       `json:"mq_137" db:"mq_137" validate:"required"`
	MQ135    int       `json:"mq_135" db:"mq_135" validate:"required"`
	MQ6      int       `json:"mq_6" db:"mq_6" validate:"required"`
	Label    string    `json:"label" db:"label" validate:"required"`
}
