package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"finalJob/pkg/api"
	"finalJob/pkg/db"
)

func getPort() string {
	portFromEnv := os.Getenv("TODO_PORT")
	if portFromEnv != "" {
		return portFromEnv
	}
	return "7540"
}

func main() {
	// Инициализируем базу данных
	err := db.InitDB("scheduler.db")
	if err != nil {
		log.Fatal("Ошибка инициализации базы данных:", err)
	}

	// Откладываем закрытие соединения с БД до завершения программы
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Ошибка при закрытии соединения с БД: %v", closeErr)
		}
	}()

	// Регистрируем API‑обработчики
	api.Init()

	// Получаем порт для запуска сервера
	port := getPort()

	// Настраиваем статический файловый сервер для веб‑интерфейса
	webDir := "./web"
	if _, err := os.Stat(webDir); err == nil {
		fileServer := http.FileServer(http.Dir(webDir))
		http.Handle("/", http.StripPrefix("/", fileServer))
		fmt.Printf("Веб‑интерфейс доступен из директории: %s\n", webDir)
	} else {
		fmt.Printf("Директория %s не найдена, статические файлы не будут доступны\n", webDir)
	}

	// Выводим информацию о запуске
	fmt.Printf("Сервер запущен на http://localhost:%s\n", port)

	// Запускаем HTTP‑сервер
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("Ошибка при запуске сервера:", err)
	}
}
