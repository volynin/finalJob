package api

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const dateFormat = "20060102"

func afterNow(date, now time.Time) bool {
	dateTrunc := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	nowTrunc := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return dateTrunc.After(nowTrunc)
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("пустое правило повторения")
	}

	start, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return "", errors.New("некорректный формат даты dstart")
	}

	parts := strings.Split(repeat, " ")
	rule := parts[0]

	switch rule {
	case "d":
		if len(parts) != 2 {
			return "", errors.New("неверный формат правила d: требуется указать количество дней")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days <= 0 || days > 400 {
			return "", errors.New("некорректное количество дней для правила d")
		}

		startTrunc := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
		nowTrunc := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		// Если начальная дата строго после now, возвращаем её + days
		if startTrunc.After(nowTrunc) {
			nextDate := startTrunc.AddDate(0, 0, days)
			return nextDate.Format(dateFormat), nil
		}

		// Если даты равны, следующая дата — через days дней
		if startTrunc.Equal(nowTrunc) {
			nextDate := nowTrunc.AddDate(0, 0, days)
			return nextDate.Format(dateFormat), nil
		}

		// Вычисляем разницу в днях
		daysDiff := int(nowTrunc.Sub(startTrunc).Hours() / 24)

		// Количество полных периодов, прошедших с начальной даты
		fullPeriods := daysDiff / days

		// Следующая дата — начальная дата + (fullPeriods + 1) * days
		nextDate := startTrunc.AddDate(0, 0, (fullPeriods+1)*days)

		return nextDate.Format(dateFormat), nil

	case "y":
		if len(parts) != 1 {
			return "", errors.New("неверный формат правила y")
		}
		date := start
		for {
			nextDate := date.AddDate(1, 0, 0)
			// Корректировка для високосных дат: если 29.02 стало 28.02
			if nextDate.Month() == date.Month() && nextDate.Day() < date.Day() {
				// Устанавливаем последний день месяца
				lastDay := time.Date(nextDate.Year(), nextDate.Month()+1, 0, 0, 0, 0, 0, nextDate.Location())
				nextDate = lastDay
			}
			// Если полученная дата строго больше now, возвращаем её
			if afterNow(nextDate, now) {
				log.Printf("NextDate (y): now=%s, dstart=%s → next=%s",
					now.Format(dateFormat), date.Format(dateFormat), nextDate.Format(dateFormat))
				return nextDate.Format(dateFormat), nil
			}
			date = nextDate
		}

	default:
		return "", errors.New("неподдерживаемый формат правила повторения")
	}
}

// nextDateHandler обрабатывает запросы к /api/nextdate
func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	var now time.Time
	var err error

	if nowStr != "" {
		now, err = time.Parse(dateFormat, nowStr)
		if err != nil {
			http.Error(w, "некорректный формат параметра now", http.StatusBadRequest)
			return
		}
	} else {
		now = time.Now()
	}

	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(nextDate))
}
