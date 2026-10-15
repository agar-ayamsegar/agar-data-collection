package domain

import "time"

type Gas struct {
	ID       string    `json:"id" db:"id"`
	Date     time.Time `json:"date" db:"date"`
	Supplier string    `json:"supplier" db:"supplier"`
	Part     string    `json:"part" db:"part"`
	MQ137    string    `json:"mq_137" db:"mq_137"`
	MQ135    string    `json:"mq_135" db:"mq_135"`
	MQ6      string    `json:"mq_6" db:"mq_6"`
}
