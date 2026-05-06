package api

import (
	"encoding/json"
	"finalJob/pkg/db"
	"io"
	"net/http"
	"strconv"
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	limit := 50

	tasks, err := db.Tasks(limit)
	if err != nil {
		writeError(w, "ошибка получения списка задач", http.StatusInternalServerError)
		return
	}

	writeJSON(w, TasksResp{
		Tasks: tasks,
	}, http.StatusOK) // Добавляем третий аргумент — HTTP‑статус
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	writeJSON(w, task, http.StatusOK)
}
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Читаем тело запроса
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, "ошибка чтения запроса или превышен размер", http.StatusBadRequest)
		return
	}

	var task db.Task
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

	// Обновляем задачу в базе данных
	err = db.UpdateTask(&task)
	if err != nil {
		writeError(w, err.Error(), http.StatusNotFound)
		return
	}

	// Возвращаем пустой JSON при успешном обновлении
	writeJSON(w, map[string]interface{}{}, http.StatusOK)
}
