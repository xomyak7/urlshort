// пакеты исполняемых приложений должны называться main
package main

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	// "log"
	"net/http"
	"strings"
)

// Простое хранилище в памяти (для демонстрации)
var urlStorage = make(map[string]string)
const localhost = "http://localhost:8080"

// Генерация ID путем хэширования и кодирования
func generateShortID(originalURL string) string {
    // Создаем хеш от URL
    hash := sha256.Sum256([]byte(originalURL))
    
    // Берем первые 8 байт хеша и кодируем в base64
    // base64.URLEncoding использует безопасные для URL символы (- и _ вместо + и /)
    shortID := base64.URLEncoding.EncodeToString(hash[:8])
    
    // Убираем возможные символы = в конце
    shortID = strings.TrimRight(shortID, "=")
    
    return shortID
}

// функция main вызывается автоматически при запуске приложения
func main() {
    if err := run(); err != nil {
        panic(err)
    }
}

// функция run будет полезна при инициализации зависимостей сервера перед запуском
func run() error {
    return http.ListenAndServe(`:8080`, http.HandlerFunc(webhook))
}

// функция webhook — обработчик HTTP-запроса
func webhook(w http.ResponseWriter, r *http.Request) {
    // 1. Важно! Закрываем тело запроса по окончании работы функции,
    //    чтобы освободить ресурсы сетевого соединения.
    defer r.Body.Close()
    
    // // разрешаем только POST-запросы
    // w.WriteHeader(http.StatusMethodNotAllowed)

    switch r.Method {
        case http.MethodGet:
            handleGet(w, r)
        case http.MethodPost:
            handlePost(w, r)
        default:
            http.Error(w, "" /*"Method not allowed" */, http.StatusBadRequest)
    }
}

func handleGet(w http.ResponseWriter, r *http.Request) {
    // Получаем id из пути
    id := r.URL.Path

    // Проверяем, что ID не пустой и нет дополнительных слешей
    if id == "" {
        http.Error(w, "" /*"Invalid ID" */, http.StatusBadRequest)
		return
    }

    // Ищем оригинальный URL в хранилище
    originalURL, exists := urlStorage[id]
    if !exists {
        http.Error(w, "" /*"ID not found" */, http.StatusBadRequest)
		return
    }

    // Выполняем редирект
    w.Header().Set("Location", originalURL)
    w.WriteHeader(http.StatusTemporaryRedirect)
}

func handlePost(w http.ResponseWriter, r *http.Request) {
    // Проверяем Content-Type
    if r.Header.Get("Content-Type") != "text/plain" {
        http.Error(w, "" /*"Content-Type must be text/plain" */, http.StatusBadRequest)
        return
    }

    // Читаем тело запроса
    bodyBytes, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "" /*"Error reading body" */, http.StatusBadRequest)
        return
    }

    originalURL := string(bodyBytes)
    if originalURL == "" {
        http.Error(w, "" /*"URL cannot be empty" */, http.StatusBadRequest)
        return
    }

    // Генерируем короткий ID (в реальном приложении здесь должна быть более сложная логика)
	// Для примера используем фиксированный ID
	shortID := /* "/EwHXdJfB" // */"/" + generateShortID(originalURL)

    // Сохраняем соответствие URL и его сокращения
    urlStorage[shortID] = originalURL

    // Формируем сокращенный URL
    shortURL := localhost + shortID

    // Устанавливаем заголовки и статус
    w.Header().Set("Content-Type", "text/plain")
    w.WriteHeader(http.StatusCreated)
    w.Write([]byte(shortURL))
}