# API Diagnostic Tool

Инструмент для диагностики и сравнения прямого Sensor Hub API и backend proxy.

## Использование

```bash
cd back/cmd/diagnose
go run main.go <your-api-key>
```

## Что проверяет

1. **Archive Data (hourly)** - Архивные данные с интервалом час для 2026-04-24
2. **Archive Data (daily)** - Архивные данные с интервалом день для 2026-04-24
3. **Last Data** - Текущие данные с датчиков

## Вывод

Для каждого теста показывает:
- URL запросов (с замаскированным API ключом)
- HTTP статус ответов
- Размер ответов в байтах
- Количество записей данных
- Сравнение результатов
- Примеры данных при расхождениях

## Отключение кэша

Для тестирования без кэша запустите backend с переменной окружения:

```bash
DISABLE_CACHE=1 go run cmd/server/main.go
```

## Пример вывода

```
=== Air Quality Monitor API Diagnostic Tool ===
Testing date: 2026-04-24

--- Test 1: Archive Data (hourly) ---
Direct API URL: http://sensor.krasn.ru/hub/api/3.0/sets/knc-air/data/archive?... (uid=***1234)
Backend URL:    http://localhost:8080/api/datasets/knc-air/aggregated?time_begin=2026-04-24...

✅ Direct API Status: 200
   Response Size: 15234 bytes
   Data Records: 48

✅ Backend Status: 200
   Response Size: 15234 bytes
   Data Records: 48

✅ Record counts match: 48
✅ Response sizes match: 15234 bytes
```
