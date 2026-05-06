package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"finalJob/pkg/db"
)

// writeJSON отправляет JSON‑ответ с указанным HTTP‑статусом
func writeJSON(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Ошибка кодирования JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// writeError отправляет JSON‑ответ с сообщением об ошибке и указанным HTTP‑статусом
func writeError(w http.ResponseWriter, message string, statusCode int) {
	writeJSON(w, map[string]string{"error": message}, statusCode)
}

func checkDate(task *db.Task) error {
	now := time.Now()

	// Если дата пустая, берём сегодняшнюю
	if task.Date == "" {
		task.Date = now.Format(dateFormat)
		return nil
	}

	// Проверяем корректность формата даты
	t, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		return errors.New("некорректный формат даты")
	}

	// Если правило повторения есть, вычисляем следующую дату
	if task.Repeat != "" {
		startTrunc := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		nowTrunc := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		// Если дата уже прошла, вычисляем следующую
		if startTrunc.Before(nowTrunc) {
			next, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return err
			}
			task.Date = next
		} else if startTrunc.Equal(nowTrunc) {
			// Если дата сегодня, оставляем как есть (задача на сегодня)
			task.Date = t.Format(dateFormat)
		}
	} else {
		// Если правила нет, проверяем, что дата не в прошлом
		if !afterNow(t, now) {
			task.Date = now.Format(dateFormat)
		}
	}

	return nil
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Читаем тело запроса с ограничением размера (1 MB)
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, "ошибка чтения запроса или превышен размер", http.StatusBadRequest)
		return
	}

	var task db.Task
	// Десериализуем JSON в структуру
	if err := json.Unmarshal(body, &task); err != nil {
		writeError(w, "ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	// Проверяем обязательное поле title
	if task.Title == "" {
		writeError(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	// Проверяем и корректируем дату
	if err := checkDate(&task); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Добавляем задачу в базу данных
	id, err := db.AddTask(&task)
	if err != nil {
		writeError(w, "ошибка добавления задачи в базу", http.StatusInternalServerError)
		return
	}

	// Возвращаем идентификатор созданной записи
	writeJSON(w, map[string]int64{"id": id}, http.StatusOK)
}
