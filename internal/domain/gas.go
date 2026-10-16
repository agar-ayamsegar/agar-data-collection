package domain

import "time"

type Gas struct {
	ID          string    `json:"id" db:"id"`
	Date        time.Time `json:"date" db:"date" validate:"required"`
	Supplier    string    `json:"supplier" db:"supplier" validate:"required"`
	Part        string    `json:"part" db:"part" validate:"required"`
	MQ137       int       `json:"mq_137" db:"mq_137"`
	MQ135       int       `json:"mq_135" db:"mq_135"`
	MQ6         int       `json:"mq_6" db:"mq_6"`
	Label       string    `json:"label" db:"label" validate:"required"`
	Author      string    `json:"author" db:"author" validate:"required"`
	Baseline137 int       `json:"baseline_137" db:"baseline_137"`
	Baseline135 int       `json:"baseline_135" db:"baseline_135"`
	Temp        int       `json:"temp" db:"temp"`
	Humidity    int       `json:"humidity" db:"humidity"`
}
