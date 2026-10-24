package domain

import "time"

type Gas struct {
	ID            string    `json:"id" db:"id"`
	Date          time.Time `json:"date" db:"date" validate:"required"`
	Supplier      string    `json:"supplier" db:"supplier" validate:"required"`
	ChickenID     string    `json:"chicken_id" db:"chicken_id" validate:"required"`
	Label         string    `json:"label" db:"label" validate:"required,oneof=FRESH UNFRESH UNKNOWN"`
	Author        string    `json:"author" db:"author" validate:"required"`
	WeightG       float64   `json:"weight_g" db:"weight_g" validate:"required"`
	HoursSinceBuy float64   `json:"hours_since_buy" db:"hours_since_buy" validate:"required"`
	Part          string    `json:"part" db:"part" validate:"required,oneof=WHOLE BREAST THIGH"`
	IntervalMS    int       `json:"interval_ms" db:"interval_ms"`
	VRef          float64   `json:"vref" db:"vref" validate:"required,eq=1.1|eq=5"`
	LidClosedAt   int       `json:"lid_closed_at_s" db:"lid_closed_at_s" validate:"required"`
	Baseline137   []float64 `json:"baseline_137" db:"baseline_137"`
	Baseline135   []float64 `json:"baseline_135" db:"baseline_135"`
	MQ137         []float64 `json:"mq_137" db:"mq_137"`
	MQ135         []float64 `json:"mq_135" db:"mq_135"`
	Temp          []float64 `json:"temp" db:"temp"`
	Humidity      []float64 `json:"humidity" db:"humidity"`
	Notes         []string  `json:"notes" db:"notes"`
	TasteNotes    string    `json:"taste_notes" db:"taste_notes"`
	Texture       string    `json:"texture" db:"texture" validate:"omitempty,oneof=BOUNCY SOFT MUSHY"`
	Color         string    `json:"color" db:"color" validate:"omitempty,oneof=PINK_PEACH PEACH_YELLOWISH PINK_YELLOWISH PALE GREYISH"`
	SmellLevel    int       `json:"smell_level" db:"smell_level" validate:"omitempty,min=1,max=5"`
}
