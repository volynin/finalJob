package api

import (
	"finalJob/pkg/db"
	"net/http"
	"strconv"
	"time"
)

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем параметр id из URL
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	// Преобразуем в int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, "Некорректный идентификатор", http.StatusBadRequest)
		return
	}

	// Получаем задачу из БД
	task, err := db.GetTask(id)
	if err != nil {
		writeError(w, err.Error(), http.StatusNotFound)
		return
	}

	now := time.Now()

	// Если правило повторения отсутствует — удаляем задачу
	if task.Repeat == "" {
		err = db.DeleteTask(id)
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Если задача периодическая — вычисляем следующую дату
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeError(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Обновляем дату задачи в БД
		err = db.UpdateDate(nextDate, id)
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Возвращаем пустой JSON при успешном выполнении
	writeJSON(w, map[string]interface{}{}, http.StatusOK)
}
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем параметр id из URL
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	// Преобразуем в int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, "Некорректный идентификатор", http.StatusBadRequest)
		return
	}

	// Удаляем задачу из БД
	err = db.DeleteTask(id)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем пустой JSON при успешном удалении
	writeJSON(w, map[string]interface{}{}, http.StatusOK)
}
