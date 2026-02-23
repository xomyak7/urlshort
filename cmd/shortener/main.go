// пакеты исполняемых приложений должны называться main
package main

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"log"

	// "log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/xomyak7/urlshort/internal/config"
)

// Простое хранилище в памяти (для демонстрации)
var urlStorage = make(map[string]string)
var localhost = "http://localhost:8080"
var baseAddr = "http://localhost:8080/"

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
    cfg := config.Load()
    localhost = "http://" + cfg.Host
    baseAddr  = "http://" + cfg.BaseAddr

    r := chi.NewRouter()

    // Используем chi для маршрутизации - обратите внимание на {id} в пути
    r.Get("/{id}", handleGet)  // GET запросы на /{id}
    r.Post("/", handlePost)    // POST запросы на /

    log.Printf("Server starting on %s", cfg.Host)
    return http.ListenAndServe(cfg.Host, r)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
    // Получаем id из параметров маршрута chi (а не из r.URL.Path)
    id := chi.URLParam(r, "id")
    
    // Проверяем, что ID не пустой
    if id == "" {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    // Формируем ключ с ведущим слешем
    key := "/" + id

    // Ищем оригинальный URL в хранилище
    originalURL, exists := urlStorage[key]
    if !exists {
        http.Error(w, "ID not found", http.StatusNotFound) // Используем 404, а не 400
        return
    }

    // Выполняем редирект - сначала устанавливаем заголовок, потом статус
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
    shortURL := baseAddr + shortID

    // Устанавливаем заголовки и статус
    w.Header().Set("Content-Type", "text/plain")
    w.WriteHeader(http.StatusCreated)
    w.Write([]byte(shortURL))
}