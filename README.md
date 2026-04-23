# Air Quality Monitor - Monorepo

Система мониторинга качества воздуха с Go backend и Vue 3 frontend.

## Структура проекта

```
project/
├── back/                 # Go backend
│   ├── cmd/
│   │   └── server/      # Точка входа приложения
│   ├── internal/
│   │   ├── handler/     # HTTP handlers
│   │   ├── service/     # Бизнес-логика)
│   │   ├── client/      # Клиент внешнего API
│   │   ├── aggregator/  # Агрегация данных
│   │   ├── model/       # Модели данных
│   │   └── cache/       # In-memory кэш
│   └── go.mod
├── front/               # Vue 3 frontend
│   ├── src/
│   ├── public/
│   └── package.json
└── README.md
```

## Backend

### Архитектура

Backend реализован на Go с использованием Gin framework и следует принципам чистой архитектуры:

- **Handler** - обработка HTTP запросов
- **Service** - бизнес-логика, кэширование
- **Client** - взаимодействие с внешним API
- **Aggregator** - агрегация данных по временным интервалам
- **Cache** - in-memory кэш с TTL

### API Endpoints

```
GET  /api/health                              # Health check
GET  /api/datasets                            # Список наборов данных
GET  /api/datasets/:code                      # Детали набора данных
GET  /api/datasets/:code/last                 # Актуальные данные
GET  /api/datasets/:code/aggregated           # Агрегированные данные
GET  /api/datasets/:code/timeseries           # Временные ряды для графиков
GET  /api/datasets/:code/statistics           # Общая статистика
```

### Запуск backend

```bash
cd back
go mod download
export SENSOR_API_KEY="your-api-key"
go run cmd/server/main.go
```

## Frontend

```bash
cd front
npm install
npm run dev
```

## Production

### Local Production Build

```bash
cd front && npm run build
cd ../back && export SENSOR_API_KEY="your-api-key" && go run cmd/server/main.go
```

Backend отдаст frontend на http://localhost:8080

### Vercel Deployment

Проект готов к деплою на Vercel. См. подробную инструкцию: [DEPLOY_VERCEL.md](./DEPLOY_VERCEL.md)

**Краткая схема:**
1. Создать backend project (Root Directory: `back`)
2. Создать frontend project (Root Directory: `front`)
3. Настроить environment variables
4. Frontend проксирует `/api/*` на backend через Vercel rewrites
