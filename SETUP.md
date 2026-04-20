# Установка и запуск проекта

## Ваши данные
- API ключ: `tvcpgyjme771y5zo`
- Набор данных: `knc-air`

## 1. Установка Go

### Windows:
1. Скачайте Go с https://go.dev/dl/
2. Установите (рекомендуется версия 1.21 или новее)
3. Проверьте установку: `go version`

### Или через Chocolatey:
```bash
choco install golang
```

## 2. Запуск Backend

```bash
cd back

# Установить зависимости
go mod download

# Запустить сервер (API ключ уже в .env файле)
go run cmd/server/main.go
```

Backend запустится на http://localhost:8080

## 3. Запуск Frontend

В новом терминале:

```bash
cd front

# Установить зависимости (только первый раз)
npm install

# Запустить dev сервер
npm run dev
```

Frontend запустится на http://localhost:5173

## 4. Тестирование API

```bash
# Проверка работоспособности
curl http://localhost:8080/api/health

# Получить информацию о датасете knc-air
curl http://localhost:8080/api/datasets/knc-air

# Получить актуальные данные
curl http://localhost:8080/api/datasets/knc-air/last

# Получить данные за последние 7 дней (по часам)
curl "http://localhost:8080/api/datasets/knc-air/timeseries?time_begin=2026-04-13&time_end=2026-04-20&interval=hour"

# Получить статистику
curl "http://localhost:8080/api/datasets/knc-air/statistics?time_begin=2026-04-13&time_end=2026-04-20"
```

## 5. Production сборка

```bash
# Собрать frontend
cd front
npm run build

# Собрать backend
cd ../back
go build -o server.exe cmd/server/main.go

# Запустить
./server.exe
```

Приложение будет доступно на http://localhost:8080

## Структура проекта

```
air-quality-monitor/
├── back/
│   ├── cmd/server/main.go          # Точка входа
│   ├── internal/
│   │   ├── handler/                # HTTP handlers
│   │   ├── service/                # Бизнес-логика
│   │   ├── client/                 # API клиент
│   │   ├── aggregator/             # Агрегация данных
│   │   ├── cache/                  # Кэширование
│   │   └── model/                  # Модели
│   ├── .env                        # Конфигурация (API ключ)
│   └── go.mod
├── front/
│   ├── src/                        # Vue компоненты
│   ├── vite.config.js              # Vite конфиг с proxy
│   └── package.json
└── README.md
```

## Возможные проблемы

### Go не найден
Убедитесь что Go установлен и добавлен в PATH:
```bash
go version
```

### Порт 8080 занят
Измените порт в `.env`:
```
PORT=3000
```

### CORS ошибки
Проверьте что в `vite.config.js` настроен proxy на правильный порт backend.
