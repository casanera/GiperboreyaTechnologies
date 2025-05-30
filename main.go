// File: main.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Запрос к homeHandler: %s %s", r.Method, r.URL.Path)
	fmt.Fprintf(w, "Привет от командного проекта! Сервер запущен.")
}

func main() {
	log.Println("Запуск базового HTTP сервера...")
	http.HandleFunc("/", homeHandler)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
		log.Printf("Переменная окружения APP_PORT не установлена. Используется порт по умолчанию: %s", port)
	}

	log.Printf("Сервер (базовый) запускается на http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}
