# Quick Start Guide

## Быстрый запуск для разработки

### 1. Backend

```bash
cd back

# Установить зависимости
go mod download

# Создать .env файл
cp .env.example .env
# Отредактировать .env и добавить ваш SENSOR_API_KEY

# Или установить через переменную окружения
export SENSOR_API_KEY="your-api-key-here"

# Запустить сервер
go run cmd/server/main.go
```

Backend запустится на http://localhost:8080

### 2. Frontend (в отдельном терминале)

```bash
cd front

# Установить зависимости (только первый раз)
npm install

# Запустить dev сервер
npm run dev
```

Frontend запустится на http://localhost:5173

## Production сборка

```bash
# 1. Собрать frontend
cd front
npm run build

# 2. Запустить backend (он будет отдавать статику)
cd ../back
export SENSOR_API_KEY="your-api-key-here"
go run cmd/server/main.go
```

Приложение доступно на http://localhost:8080

## API Endpoints

- `GET /api/health` - проверка работоспособности
- `GET /api/datasets` - список наборов данных
- `GET /api/datasets/:code` - детали набора
- `GET /api/datasets/:code/last` - актуальные данные
- `GET /api/datasets/:code/timeseries` - временные ряды
- `GET /api/datasets/:code/statistics` - статистика

## Примеры запросов

```bash
# Health check
curl http://localhost:8080/api/health

# Получить список датасетов
curl http://localhost:8080/api/datasets

# Получить актуальные данные
curl "http://localhost:8080/api/datasets/air/last"

# Получить временной ряд за последние 7 дней
curl "http://localhost:8080/api/datasets/air/timeseries?time_begin=2026-04-13&time_end=2026-04-20&interval=hour"
```

## Структура

```
air-quality-monitor/
├── back/           # Go backend
│   ├── cmd/
│   ├── internal/
│   └── go.mod
├── front/          # Vue 3 frontend
│   ├── src/
│   └── package.json
└── README.md
```
