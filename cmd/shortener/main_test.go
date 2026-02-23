package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	
	"github.com/go-chi/chi/v5"
)

// TestHandlePost тестирует обработчик POST-запросов
func TestHandlePost(t *testing.T) {
	// Очищаем хранилище перед каждым тестом
	urlStorage = make(map[string]string)

	tests := []struct {
		name           string
		contentType    string
		body           string
		expectedStatus int
		expectedBody   string
		checkStorage   bool
	}{
		{
			name:           "Successful URL shortening",
			contentType:    "text/plain",
			body:           "https://practicum.yandex.ru/",
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/QrPnX5IUXSU", // Заметьте, здесь уже есть слеш в начале
			checkStorage:   true,
		},
		{
			name:           "Wrong Content-Type",
			contentType:    "application/json",
			body:           "https://practicum.yandex.ru/",
			expectedStatus: http.StatusBadRequest,
			checkStorage:   false,
		},
		{
			name:           "Empty body",
			contentType:    "text/plain",
			body:           "",
			expectedStatus: http.StatusBadRequest,
			checkStorage:   false,
		},
		{
			name:           "Missing Content-Type",
			contentType:    "",
			body:           "https://practicum.yandex.ru/",
			expectedStatus: http.StatusBadRequest,
			checkStorage:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Очищаем хранилище перед каждым тестом
			urlStorage = make(map[string]string)
			
			// Создаем POST-запрос
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			// Создаем ResponseRecorder для записи ответа
			w := httptest.NewRecorder()

			// Вызываем обработчик
			handlePost(w, req)

			// Проверяем статус-код
			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Проверяем тело ответа для успешных запросов
			if tt.expectedStatus == http.StatusCreated {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}

				// Проверяем, что тело ответа не пустое и содержит ожидаемый ID
				if !strings.Contains(string(body), tt.expectedBody) {
					t.Errorf("Expected body to contain %q, got %q", tt.expectedBody, string(body))
				}

				// Проверяем Content-Type заголовок
				contentType := resp.Header.Get("Content-Type")
				if contentType != "text/plain" {
					t.Errorf("Expected Content-Type: text/plain, got %q", contentType)
				}

				// Проверяем, что данные сохранились в storage
				if tt.checkStorage {
					if len(urlStorage) == 0 {
						t.Error("Expected storage to have 1 entry, got 0")
					}
				}
			}
		})
	}
}

// TestHandleGet тестирует обработчик GET-запросов
func TestHandleGet(t *testing.T) {
	// Очищаем и заполняем хранилище тестовыми данными
	urlStorage = make(map[string]string)
	urlStorage["/QrPnX5IUXSU"] = "https://practicum.yandex.ru/"
	urlStorage["/abc123"] = "https://google.com/"

	tests := []struct {
		name              string
		id                string
		expectedStatus    int
		expectedLocation  string
	}{
		{
			name:              "Successful redirect",
			id:                "QrPnX5IUXSU",
			expectedStatus:    http.StatusTemporaryRedirect,
			expectedLocation:  "https://practicum.yandex.ru/",
		},
		{
			name:              "Another successful redirect",
			id:                "abc123",
			expectedStatus:    http.StatusTemporaryRedirect,
			expectedLocation:  "https://google.com/",
		},
		{
			name:              "Non-existent ID",
			id:                "nonexistent",
			expectedStatus:    http.StatusNotFound,
			expectedLocation:  "",
		},
		{
			name:              "Empty ID",
			id:                "",
			expectedStatus:    http.StatusBadRequest,
			expectedLocation:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем GET-запрос
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			
			// Важно! Добавляем chi контекст с параметром
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			// Создаем ResponseRecorder
			w := httptest.NewRecorder()

			// Вызываем обработчик
			handleGet(w, req)

			// Проверяем результат
			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Проверяем Location заголовок
			location := resp.Header.Get("Location")
			if location != tt.expectedLocation {
				t.Errorf("Expected Location %q, got %q", tt.expectedLocation, location)
			}

			// Для успешных редиректов тело должно быть пустым
			if tt.expectedStatus == http.StatusTemporaryRedirect {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}
				if len(body) != 0 {
					t.Errorf("Expected empty body for redirect, got %q", string(body))
				}
			}
		})
	}
}