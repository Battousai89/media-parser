# Media Parser

🌐 **Media Parser** — это мощное приложение для парсинга медиа-контента из интернета. Состоит из бэкенда на Go, фронтенда на Vue 3 + Tauri и набора сервисов для обработки задач.

---

## 🚀 Возможности

- **Парсинг URL** — извлечение медиа из веб-страниц
- **Пакетная обработка** — массовый парсинг нескольких URL
- **Кэширование** — Redis для ускорения повторных запросов
- **Очереди задач** — RabbitMQ для надёжной обработки
- **Хранение медиа** — MinIO S3-совместимое хранилище
- **Tauri Desktop** — кроссплатформенное десктопное приложение
- **Swagger API** — полная документация API

---

## 🏗 Архитектура

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   Client    │────▶│   Backend    │────▶│   RabbitMQ  │
│ (Vue + Tauri)     │    (Go)      │     │   (Queue)   │
└─────────────┘     └──────────────┘     └─────────────┘
                          │
         ┌────────────────┼────────────────┐
         ▼                ▼                ▼
   ┌──────────┐    ┌──────────┐    ┌──────────┐
   │ PostgreSQL│   │  Redis   │    │   MinIO  │
   │  (DB)    │    │ (Cache)  │    │ (Storage)│
   └──────────┘    └──────────┘    └──────────┘
```

---

## 📦 Технологии

| Компонент | Технологии |
|-----------|------------|
| **Backend** | Go, Gin, Swagger |
| **Frontend** | Vue 3, TypeScript, Vite, Naive UI, Pinia, Vue Router |
| **Desktop** | Tauri v2 |
| **БД** | PostgreSQL 18 |
| **Кэш** | Redis 8 |
| **Очереди** | RabbitMQ 4 |
| **Хранилище** | MinIO |
| **Контейнеризация** | Docker, Docker Compose |

---

## 🛠 Быстрый старт

### Требования

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- Make (опционально)

### 1. Клонирование и настройка

```bash
git clone <repository-url>
cd media-parser
make setup
```

### 2. Запуск инфраструктуры

```bash
make docker-up
```

### 3. Применение миграций

```bash
make backend-migrate
```

### 4. Генерация API ключа

```bash
make backend-generate-key
```

### 5. Запуск разработки

```bash
# Бэкенд + фронтенд
make dev

# Полная среда (Docker + бэкенд + Tauri)
make dev-full
```

---

## 📚 API Документация

После запуска бэкенда Swagger доступен по адресу:

👉 **http://localhost:8080/swagger/index.html**

### Основные эндпоинты

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/v1/parse/url` | Парсинг одного URL |
| `POST` | `/api/v1/parse/batch` | Пакетный парсинг |
| `GET` | `/api/v1/media` | Список медиа |
| `POST` | `/api/v1/media/upload` | Загрузка медиа |
| `GET` | `/api/v1/requests` | Список запросов |
| `GET` | `/api/v1/sources` | Управление источниками |

---

## 🎯 Команды Make

### Backend

```bash
make backend-install       # Установить зависимости
make backend-run           # Запустить локально
make backend-build         # Собрать бэкенд
make backend-test          # Тесты
make backend-lint          # Линтер
make backend-swagger       # Swagger документация
make backend-migrate       # Применить миграции
make backend-generate-key  # Сгенерировать API ключ
```

### Frontend

```bash
make frontend-install      # Установить зависимости
make frontend-dev          # Режим разработки
make frontend-build        # Сборка
make frontend-tauri-dev    # Tauri dev
make frontend-tauri-build  # Сборка Tauri приложения
```

### Docker

```bash
make docker-up             # Запустить все сервисы
make docker-down           # Остановить все сервисы
make docker-logs           # Логи всех сервисов
make docker-clean          # Очистка
```

---

## 🌐 Официальный сервер

Официальный бэкенд развернут и доступен по адресу:

🔗 **https://battousai.fun**

### Получение токена доступа

Для получения API токена обратитесь к разработчику:

📬 **Telegram:** [@HimuraDev](https://t.me/HimuraDev)

---

## 📁 Структура проекта

```
media-parser/
├── backend/
│   ├── cmd/
│   │   ├── main.go          # Точка входа
│   │   └── keygen/          # Генератор ключей
│   ├── internal/
│   │   ├── handler/         # HTTP обработчики
│   │   ├── service/         # Бизнес-логика
│   │   ├── repository/      # Работа с данными
│   │   ├── queue/           # RabbitMQ
│   │   └── config/          # Конфигурация
│   ├── migrations/          # SQL миграции
│   └── docs/swagger/        # Swagger документация
├── client/
│   ├── src/                 # Исходный код Vue
│   ├── src-tauri/           # Tauri конфигурация
│   ├── public/              # Статические файлы
│   └── dist/                # Сборка
├── docker-compose.yaml      # Docker сервисы
├── Makefile                 # Make команды
└── .env.template            # Шаблон окружения
```

---

## 🔐 Безопасность

- **API Keys** — аутентификация через заголовок `X-Auth-Token`
- **Rate Limiting** — ограничение запросов (10 req/s, burst 20)
- **CORS** — настройка кросс-доменных запросов
- **Graceful Shutdown** — корректное завершение работы

---

## 🧪 Тестирование

```bash
# Все тесты
make test

# Только бэкенд
make backend-test

# Только фронтенд
make frontend-test-run
```

---

## 📝 Лицензия

[MIT](./LICENSE)

---

## 🤝 Контакты

- **Разработчик:** [@HimuraDev](https://t.me/HimuraDev)
- **Официальный сервер:** https://battousai.fun

---

## 🙏 Благодарности

- [Vue.js](https://vuejs.org/)
- [Tauri](https://tauri.app/)
- [Go](https://go.dev/)
- [Naive UI](https://www.naiveui.com/)
