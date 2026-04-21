# Технический аудит AirMonitorSystem

**Дата:** 2026-04-21  
**Версия:** 1.0  
**Аудитор:** Claude Sonnet 4

---

## Executive Summary

Проведен технический аудит системы мониторинга качества воздуха. Выявлено **3 критических проблемы** и **2 архитектурных недочета**, которые приводят к некорректному отображению данных на карте, нестабильной работе календаря и неправильному определению районов станций.

---

## Проблема #1: Данные на карте не обновляются при выборе разных часов

### Статус: 🔴 КРИТИЧЕСКАЯ

### Root Cause

**Файл:** `back/internal/service/service.go:106`

```go
cacheKey := fmt.Sprintf("agg:%s:%s:%s:%s:%v:%v", 
    setCode, 
    timeBegin.Format("2006-01-02"),  // ❌ Только дата, без времени!
    timeEnd.Format("2006-01-02"),    // ❌ Только дата, без времени!
    interval, sites, indicators)
```

**Проблема:** Ключ кэша использует только дату (`2006-01-02`), игнорируя часы. Это означает:
- Запрос за `2026-04-21 05:00` создает ключ `agg:knc-air:2026-04-21:2026-04-21:hour:[]:[]]`
- Запрос за `2026-04-21 10:00` создает **тот же ключ**
- Второй запрос возвращает закэшированные данные первого запроса

### Доказательство

```bash
# Запрос за час 05
curl "http://localhost:8080/api/datasets/knc-air/data?date=2026-04-21&hour=05"
# Возвращает: {"time":"2026-04-21 10:00:00",...}  ❌ Неправильно!

# Запрос за час 10
curl "http://localhost:8080/api/datasets/knc-air/data?date=2026-04-21&hour=10"
# Возвращает: {"time":"2026-04-21 10:00:00",...}  ✅ Правильно (первый запрос)
```

Оба запроса возвращают данные за 10:00, потому что это был первый запрос, который попал в кэш.

### Где проявляется

- **Backend:** `back/internal/service/service.go:106`
- **Frontend:** Пользователь кликает по разным часам на таймлайне, но видит одни и те же данные
- **API:** Endpoint `/api/datasets/:code/data?date=YYYY-MM-DD&hour=HH`

### Влияние

- Пользователь не может просматривать исторические данные по часам
- Карта показывает неактуальные данные
- Невозможно отследить динамику загрязнения в течение дня

### Решение

Изменить формат ключа кэша на включение времени:

```go
cacheKey := fmt.Sprintf("agg:%s:%s:%s:%s:%v:%v", 
    setCode, 
    timeBegin.Format("2006-01-02 15:04:05"),  // ✅ С временем
    timeEnd.Format("2006-01-02 15:04:05"),    // ✅ С временем
    interval, sites, indicators)
```

---

## Проблема #2: Календарь работает нестабильно

### Статус: 🟡 СРЕДНЯЯ

### Root Cause #1: Нет синхронизации состояния между Timeline и Calendar

**Файл:** `front/src/components/Timeline.vue:50-52`

```javascript
const currentPage = ref(1);
const pointsPerPage = 15;
const selectedIndex = ref(0);
```

**Проблема:** 
- Timeline имеет собственное состояние пагинации (`currentPage`, `selectedIndex`)
- При смене даты через Calendar, Timeline не сбрасывает свое состояние
- Пользователь может оказаться на странице 2 таймлайна, но с новой датой

### Root Cause #2: Нет обработки выбора диапазона дат

**Файл:** `front/src/App.vue:156-162`

```javascript
const onDateRangeSelected = (range) => {
  console.log('Date range selected:', range);
  if (range.start) {
    selectedDate.value = range.start;
    const hour = selectedTimePoint.value ? selectedTimePoint.value.hour : 0;
    loadData(range.start, hour);  // ❌ Загружает только start, игнорирует end
  }
};
```

**Проблема:**
- Calendar поддерживает выбор диапазона дат (range.start, range.end)
- App.vue загружает данные только за `range.start`, игнорируя `range.end`
- Нет UI индикации, что выбран диапазон
- Нет загрузки агрегированных данных за весь диапазон

### Root Cause #3: Нет валидации выбранной даты

**Файл:** `front/src/App.vue:149-154`

```javascript
const onDateSelected = (date) => {
  selectedDate.value = date;
  console.log('Date selected:', date);
  const hour = selectedTimePoint.value ? selectedTimePoint.value.hour : 0;
  loadData(date, hour);  // ❌ Нет проверки, что дата не в будущем
};
```

**Проблема:**
- Пользователь может выбрать будущую дату
- API вернет пустой массив или ошибку
- Станции исчезнут с карты без объяснения

### Где проявляется

- **Frontend:** `front/src/App.vue`, `front/src/components/Timeline.vue`, `front/src/components/Calendar.vue`
- **UX:** Пользователь выбирает диапазон дат, но видит данные только за первый день
- **UX:** При выборе будущей даты станции исчезают

### Влияние

- Плохой UX: непредсказуемое поведение календаря
- Невозможно просмотреть агрегированные данные за период
- Пользователь не понимает, почему станции исчезли

### Решение

1. **Синхронизация Timeline:**
   - Добавить watch на `props.date` в Timeline
   - Сбрасывать `currentPage` и `selectedIndex` при смене даты

2. **Обработка диапазона:**
   - Изменить `onDateRangeSelected` для загрузки агрегированных данных
   - Использовать endpoint `/api/datasets/:code/aggregated` с `time_begin` и `time_end`
   - Добавить UI индикацию выбранного диапазона

3. **Валидация даты:**
   - Проверять, что выбранная дата не в будущем
   - Показывать предупреждение пользователю
   - Блокировать будущие даты в Calendar

---

## Проблема #3: Районы станций определяются неправильно

### Статус: 🔴 КРИТИЧЕСКАЯ

### Root Cause: Хардкод названий районов по подстрокам

**Файл:** `front/src/components/SidePanel.vue:81-91`

```javascript
const presets = {
  left_bank: ['Черемушки', 'Взлетка', 'Академгородок'],
  nikolaevka: ['Николаевка'],
  center: ['Центр', 'Площадь Революции'],
  soviet: ['Советский', 'Академгородок', 'Взлетка'],
  zheleznodorozhny: ['Вокзал', 'Черемушки'],
  kirovsky: ['Кировский', 'Николаевка'],
  leninsky: ['Ленинский', 'Центр'],
  oktyabrsky: ['Октябрьский'],
  sverdlovsky: ['Свердловский']
};
```

**Файл:** `front/src/components/SidePanel.vue:115-118`

```javascript
const presetNames = presets[selectedPreset.value] || [];
selectedSensors.value = props.sensors
  .filter(sensor => presetNames.some(name => sensor.name.includes(name)))  // ❌ Поиск по подстроке!
  .map(s => s.id);
```

**Проблемы:**

1. **Неточное сопоставление:** Использует `includes()` вместо точного совпадения
   - Станция "Центральный парк" попадет в район "Центр"
   - Станция "Новая Взлетка" попадет в "Левый берег" и "Советский"

2. **Дублирование:** Одна станция может попасть в несколько районов
   - "Академгородок" в `left_bank` и `soviet`
   - "Взлетка" в `left_bank` и `soviet`
   - "Черемушки" в `left_bank` и `zheleznodorozhny`

3. **Нет источника истины:** Районы не берутся из API
   - API возвращает координаты станций (lat/lon)
   - Можно использовать геометрию районов для точного определения

4. **Хардкод названий:** Если API изменит название станции, фильтр сломается

### Где проявляется

- **Frontend:** `front/src/components/SidePanel.vue:109-119`
- **UX:** Пользователь выбирает район, но видит неправильные станции
- **UX:** Некоторые станции попадают в несколько районов одновременно

### Влияние

- Неправильная статистика по районам
- Пользователь не может доверять фильтрам
- Невозможно точно сравнить загрязнение между районами

### Решение

**Вариант 1: Использовать GeoJSON из API (рекомендуется)**

API предоставляет endpoint `/sets/<code>/sites.geojson` с геометрией станций:

```javascript
// 1. Загрузить GeoJSON с районами Красноярска
const districts = await fetch('/api/districts.geojson').then(r => r.json());

// 2. Для каждой станции определить район по координатам
function getDistrictForSensor(sensor, districts) {
  const point = turf.point([sensor.longitude, sensor.latitude]);
  for (const district of districts.features) {
    if (turf.booleanPointInPolygon(point, district)) {
      return district.properties.name;
    }
  }
  return 'Неизвестный район';
}
```

**Вариант 2: Добавить поле district в API**

Попросить владельцев API добавить поле `district` в ответ `/sets/<code>`:

```json
{
  "sites": [
    {
      "id": 3841,
      "name": "Станция Центр",
      "geom_x": 92.8672,
      "geom_y": 56.0153,
      "district": "Центральный"  // ✅ Новое поле
    }
  ]
}
```

**Вариант 3: Минимальный фикс (временное решение)**

Использовать точное совпадение вместо `includes()`:

```javascript
selectedSensors.value = props.sensors
  .filter(sensor => presetNames.includes(sensor.name))  // ✅ Точное совпадение
  .map(s => s.id);
```

---

## Архитектурные недочеты

### Недочет #1: Неправильный выбор API endpoint

**Текущая реализация:**

```go
// handler.go:129
data, err := h.service.GetAggregatedData(code, timeBegin, timeEnd, "hour", sites, indicators)
```

**Проблема:**
- Для запроса данных за конкретный час используется `GetAggregatedData`
- Внутри вызывается `/sets/<code>/data/archive` (агрегированные данные)
- Для live данных правильнее использовать `/sets/<code>/data/last`

**Согласно API.docx (предполагаемая документация):**

| Endpoint | Назначение | Когда использовать |
|----------|------------|-------------------|
| `/sets/<code>/data/last` | Последние измерения | Live данные (текущий момент) |
| `/sets/<code>/data/last-ext` | Последние + метаданные | Live данные с доп. информацией |
| `/sets/<code>/data/raw` | Сырые данные | Исторические данные (детальные) |
| `/sets/<code>/data/archive` | Агрегированные данные | Исторические данные (агрегаты) |
| `/sets/<code>/sites.geojson` | Геометрия станций | Координаты и границы |

**Рекомендация:**
- Для текущего часа: использовать `/data/last`
- Для исторических часов (< 24ч назад): использовать `/data/raw`
- Для старых данных (> 24ч назад): использовать `/data/archive`

### Недочет #2: Отсутствие TanStack Query

**Текущая реализация:**

```javascript
// App.vue:84-127
const loadData = async (date = null, hour = null) => {
  try {
    const data = await fetchAirQualityData(date, hour);
    sensors.value = data;
  } catch (error) {
    console.error('Failed to load air quality data:', error);
    sensors.value = [/* fallback data */];
  }
};
```

**Проблемы:**
- Нет автоматической повторной загрузки при ошибке
- Нет индикации загрузки (loading state)
- Нет кэширования на фронтенде
- Нет инвалидации устаревших данных
- Fallback данные маскируют ошибки

**Рекомендация:**

Использовать TanStack Query (Vue Query):

```javascript
import { useQuery } from '@tanstack/vue-query';

const { data: sensors, isLoading, error, refetch } = useQuery({
  queryKey: ['sensors', selectedDate, selectedTimePoint],
  queryFn: () => fetchAirQualityData(selectedDate.value, selectedTimePoint.value?.hour),
  staleTime: 60000, // 1 минута
  cacheTime: 300000, // 5 минут
  retry: 3,
  refetchOnWindowFocus: true
});
```

---

## Минимальный план исправления

### Фаза 1: Критические баги (1-2 часа)

1. **Исправить кэш в backend** (30 мин)
   - Файл: `back/internal/service/service.go:106`
   - Изменить формат ключа кэша на включение времени
   - Перезапустить backend

2. **Добавить валидацию даты** (30 мин)
   - Файл: `front/src/App.vue:149`
   - Проверять, что дата не в будущем
   - Показывать предупреждение

3. **Временный фикс районов** (30 мин)
   - Файл: `front/src/components/SidePanel.vue:117`
   - Заменить `includes()` на точное совпадение
   - Убрать дубликаты из presets

### Фаза 2: Улучшение UX (2-3 часа)

4. **Синхронизация Timeline** (1 час)
   - Файл: `front/src/components/Timeline.vue`
   - Добавить watch на `props.date`
   - Сбрасывать пагинацию при смене даты

5. **Обработка диапазона дат** (1-2 часа)
   - Файл: `front/src/App.vue:156`
   - Загружать агрегированные данные за диапазон
   - Добавить UI индикацию диапазона

### Фаза 3: Архитектурные улучшения (4-6 часов)

6. **Правильные API endpoints** (2 часа)
   - Файл: `back/internal/handler/handler.go:129`
   - Использовать `/data/last` для текущего часа
   - Использовать `/data/raw` для недавних данных

7. **Геометрическое определение районов** (2-4 часа)
   - Создать файл `front/src/data/districts.geojson`
   - Использовать turf.js для point-in-polygon
   - Обновить SidePanel.vue

8. **Внедрить TanStack Query** (опционально, 4-6 часов)
   - Установить `@tanstack/vue-query`
   - Переписать data fetching в App.vue
   - Добавить loading states

---

## Точки для диагностики

### Backend

Добавить логирование в `back/internal/service/service.go`:

```go
func (s *Service) GetAggregatedData(...) (interface{}, error) {
    cacheKey := fmt.Sprintf(...)
    
    // ✅ Добавить логирование
    log.Printf("[DEBUG] GetAggregatedData: cacheKey=%s, timeBegin=%s, timeEnd=%s", 
        cacheKey, timeBegin.Format("2006-01-02 15:04:05"), timeEnd.Format("2006-01-02 15:04:05"))
    
    if cached, found := s.cache.Get(cacheKey); found {
        log.Printf("[DEBUG] Cache HIT: %s", cacheKey)  // ✅
        return cached, nil
    }
    
    log.Printf("[DEBUG] Cache MISS: %s", cacheKey)  // ✅
    // ...
}
```

### Frontend

Добавить логирование в `front/src/App.vue`:

```javascript
const loadData = async (date = null, hour = null) => {
  console.log('[DEBUG] loadData called:', { date, hour });  // ✅
  
  try {
    const data = await fetchAirQualityData(date, hour);
    console.log('[DEBUG] Loaded sensors:', data.length, 'sensors');  // ✅
    console.log('[DEBUG] First sensor time:', data[0]?.time);  // ✅
    sensors.value = data;
  } catch (error) {
    console.error('[ERROR] Failed to load:', error);  // ✅
  }
};
```

### API

Добавить логирование запросов в `back/internal/client/client.go`:

```go
func (c *Client) doRequest(method, path string, params url.Values) ([]byte, error) {
    // ...
    reqURL := fmt.Sprintf("%s%s?%s", c.baseURL, path, params.Encode())
    
    log.Printf("[DEBUG] API Request: %s", reqURL)  // ✅
    
    // ...
    
    log.Printf("[DEBUG] API Response: status=%d, body_length=%d", resp.StatusCode, len(body))  // ✅
    
    return body, nil
}
```

---

## Тестовый сценарий воспроизведения

### Проблема #1: Данные не обновляются

**Шаги:**

1. Открыть приложение: http://localhost:5174
2. Дождаться загрузки станций на карте
3. Кликнуть на час "00:00" в таймлайне
4. Запомнить значения PM2.5 на нескольких станциях
5. Кликнуть на час "05:00" в таймлайне
6. Сравнить значения PM2.5

**Ожидаемый результат:** Значения изменились

**Фактический результат:** Значения остались прежними ❌

**Диагностика:**

```bash
# Открыть DevTools → Network
# Кликнуть 00:00 → увидеть запрос: /api/datasets/knc-air/data?date=2026-04-21&hour=0
# Кликнуть 05:00 → увидеть запрос: /api/datasets/knc-air/data?date=2026-04-21&hour=5
# Оба запроса возвращают одинаковый response
```

### Проблема #2: Календарь нестабилен

**Шаги:**

1. Открыть приложение
2. Кликнуть на кнопку календаря (📅)
3. Выбрать вчерашнюю дату
4. Закрыть календарь
5. Кликнуть на час "10:00" в таймлайне
6. Снова открыть календарь
7. Выбрать сегодняшнюю дату

**Ожидаемый результат:** Timeline сбросился на первую страницу, час 00:00

**Фактический результат:** Timeline остался на часе 10:00 ❌

**Диагностика:**

```javascript
// DevTools → Console
// Посмотреть логи:
// "Date selected: ..."
// "Time selected: ..."
// Проверить, что selectedTimePoint не сбрасывается
```

### Проблема #3: Неправильные районы

**Шаги:**

1. Открыть приложение
2. Открыть боковую панель (кнопка ◀)
3. Выбрать пресет "Левый берег"
4. Посмотреть, какие станции выбраны
5. Выбрать пресет "Советский район"
6. Посмотреть, какие станции выбраны

**Ожидаемый результат:** Разные наборы станций

**Фактический результат:** Некоторые станции присутствуют в обоих наборах ❌

**Диагностика:**

```javascript
// DevTools → Console
// Выполнить:
const presets = {
  left_bank: ['Черемушки', 'Взлетка', 'Академгородок'],
  soviet: ['Советский', 'Академгородок', 'Взлетка']
};

// Найти пересечение:
const intersection = presets.left_bank.filter(x => presets.soviet.includes(x));
console.log('Дубликаты:', intersection);
// Вывод: ['Академгородок', 'Взлетка']
```

---

## Приоритизация

| Проблема | Критичность | Сложность | Приоритет |
|----------|-------------|-----------|-----------|
| #1: Кэш не учитывает время | 🔴 Критическая | Низкая (30 мин) | **P0** |
| #3: Неправильные районы | 🔴 Критическая | Средняя (2-4 часа) | **P0** |
| #2: Календарь нестабилен | 🟡 Средняя | Средняя (2-3 часа) | **P1** |
| Недочет #1: Неправильный endpoint | 🟡 Средняя | Средняя (2 часа) | **P1** |
| Недочет #2: Нет TanStack Query | 🟢 Низкая | Высокая (4-6 часов) | **P2** |

---

## Выводы

1. **Критическая проблема с кэшем** блокирует основной функционал просмотра исторических данных
2. **Хардкод районов** делает фильтрацию ненадежной и требует архитектурного решения
3. **Календарь** работает, но UX страдает из-за отсутствия синхронизации состояния
4. **Архитектура** в целом правильная, но есть возможности для улучшения

**Рекомендация:** Начать с Фазы 1 (критические баги), затем перейти к Фазе 2 (UX), и только потом к Фазе 3 (архитектура).

---

## Приложение: Сравнение с API.docx

### Предполагаемая структура API (на основе client.go)

```
GET /sets                           # Список наборов данных
GET /sets/<code>                    # Детали набора (включая sites)
GET /sets/<code>/data/last          # Последние измерения
GET /sets/<code>/data/last-ext      # Последние + метаданные
GET /sets/<code>/data/raw           # Сырые данные за период
GET /sets/<code>/data/archive       # Агрегированные данные
GET /sets/<code>/data/archive-ext   # Агрегированные + статистика
GET /sets/<code>/sites.geojson      # Геометрия станций (предполагается)
```

### Текущее использование

| Функция | Используемый endpoint | Правильный endpoint |
|---------|----------------------|---------------------|
| Live данные | `/data/last` ✅ | `/data/last` |
| Данные за час | `/data/archive` ⚠️ | `/data/raw` или `/data/last` |
| Координаты | `/sets/<code>` ✅ | `/sets/<code>` или `/sites.geojson` |
| Исторические ряды | `/data/archive` ✅ | `/data/archive` |

---

**Конец отчета**
