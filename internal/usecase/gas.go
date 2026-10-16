package usecase

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/agar-ayamsegar/agar-data-collection/internal/domain"
	"github.com/agar-ayamsegar/agar-data-collection/internal/repository/mongo"
	"github.com/agar-ayamsegar/agar-data-collection/internal/utils/config"
	"github.com/google/uuid"
)


var ErrInvalidFile = errors.New("invalid csv file")

const dateLayout = "2006-01-02"

var requiredColumns = []string{"date", "supplier", "part", "label", "author"}

var textColumns = []string{"supplier", "part", "label", "author"}

var intColumns = []string{
	"mq_137", "mq_135", "mq_6",
	"baseline_137", "baseline_135",
	"temp", "humidity",
}

type IGasUsecase interface {
	CreateGas(ctx context.Context, gas domain.Gas) (domain.Gas, error)
	ListGas(ctx context.Context, author string) ([]domain.Gas, error)
	ImportCSV(ctx context.Context, src io.Reader) (int, error)
}

type GasUsecase struct {
	app  *config.AppConfig
	repo mongo.IGasRepository
}

func NewGasUsecase(app *config.AppConfig, repo mongo.IGasRepository) IGasUsecase {
	return &GasUsecase{
		app:  app,
		repo: repo,
	}
}

func (g *GasUsecase) CreateGas(ctx context.Context, gas domain.Gas) (domain.Gas, error) {
	log := g.app.Logger.WithField("method", "CreateGas")

	err := g.repo.InsertLog(ctx, gas)
	if err != nil {
		log.WithError(err).Error("failed to call db insert gas")
		return domain.Gas{}, err
	}

	log.WithField("id", gas.ID).Info("creating the gas successfully")

	return gas, nil
}

func (g *GasUsecase) ListGas(ctx context.Context, author string) ([]domain.Gas, error) {
	log := g.app.Logger.WithField("method", "ListGasByAuthor")

	gas, err := g.repo.ListLogByAuthor(ctx, author)
	if err != nil {
		log.WithError(err).Error("failed to call db list gas by author")
		return nil, err
	}

	return gas, nil
}


func (g *GasUsecase) ImportCSV(ctx context.Context, src io.Reader) (int, error) {
	log := g.app.Logger.WithField("method", "ImportCSV")

	reader := csv.NewReader(src)
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		log.WithError(err).Error("failed to read csv header")
		return 0, fmt.Errorf("%w: cannot read header: %v", ErrInvalidFile, err)
	}

	columns := make(map[string]int)
	for i, name := range header {
		columns[strings.ToLower(strings.TrimSpace(name))] = i
	}

	if missing := missingColumns(columns); len(missing) > 0 {
		log.WithField("missing", missing).Error("csv is missing required columns")
		return 0, fmt.Errorf("%w: missing required columns: %s", ErrInvalidFile, strings.Join(missing, ", "))
	}

	var batch []domain.Gas
	var problems []string

	line := 1

	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		line++

		if err != nil {
			problems = append(problems, fmt.Sprintf("row %d: %v", line, err))
			continue
		}

		gas, err := parseGasRow(record, columns)
		if err != nil {
			problems = append(problems, fmt.Sprintf("row %d: %v", line, err))
			continue
		}

		batch = append(batch, gas)
	}

	if len(problems) > 0 {
		log.WithField("problems", len(problems)).Warn("csv rejected, nothing inserted")
		return 0, fmt.Errorf("%w: %s", ErrInvalidFile, strings.Join(problems, "; "))
	}

	if len(batch) == 0 {
		return 0, fmt.Errorf("%w: file has a header but no data rows", ErrInvalidFile)
	}

	if err := g.repo.InsertMany(ctx, batch); err != nil {
		log.WithError(err).Error("failed to bulk insert gas")
		return 0, err
	}

	log.WithField("inserted", len(batch)).Info("csv import finished")

	return len(batch), nil
}

func missingColumns(columns map[string]int) []string {
	var missing []string

	for _, name := range requiredColumns {
		if _, ok := columns[name]; !ok {
			missing = append(missing, name)
		}
	}

	return missing
}

func parseGasRow(record []string, columns map[string]int) (domain.Gas, error) {
	field := func(name string) string {
		i, ok := columns[name]
		if !ok || i >= len(record) {
			return ""
		}

		return strings.TrimSpace(record[i])
	}

	date, err := time.Parse(dateLayout, field("date"))
	if err != nil {
		return domain.Gas{}, fmt.Errorf("column date: %q is not a date like 2026-01-15", field("date"))
	}

	for _, name := range textColumns {
		if field(name) == "" {
			return domain.Gas{}, fmt.Errorf("column %s: must not be empty", name)
		}
	}

	numbers := make(map[string]int, len(intColumns))

	for _, name := range intColumns {
		raw := field(name)
		if raw == "" {
			continue
		}

		value, err := strconv.Atoi(raw)
		if err != nil {
			return domain.Gas{}, fmt.Errorf("column %s: %q is not a whole number", name, raw)
		}

		numbers[name] = value
	}

	return domain.Gas{
		ID:          uuid.New().String(),
		Date:        date,
		Supplier:    field("supplier"),
		Part:        field("part"),
		Label:       field("label"),
		Author:      field("author"),
		MQ137:       numbers["mq_137"],
		MQ135:       numbers["mq_135"],
		MQ6:         numbers["mq_6"],
		Baseline137: numbers["baseline_137"],
		Baseline135: numbers["baseline_135"],
		Temp:        numbers["temp"],
		Humidity:    numbers["humidity"],
	}, nil
}
