# Анализ текущего состояния для добавления средних значений

## Что уже реализовано (из предыдущих коммитов):

### ✅ Calendar.vue
- Режимы выбора: day/month/year
- Кнопка "Сегодня" работает мгновенно
- Выбор диапазонов дней/месяцев/годов
- Преобразование месяцев/годов в диапазоны дат

### ✅ Timeline.vue
- Отображение точек по дням/месяцам/годам
- Формат отображения даты/диапазона

### ✅ App.vue
- generateTimePointsForRange() с type: hour/day/month/year
- onTimeSelected() обрабатывает разные типы точек
- Унифицированный формат точек с startDate/endDate

### ✅ dateFormat.js
- Форматирование дат для всех режимов

### ✅ StatisticsModal.vue
- Отображение статистики по датчикам
- ECharts график (candlestick)
- Таблица средних/мин/макс значений

## Что нужно добавить:

### 1. API для средних значений по диапазону
**Файл:** `front/src/services/api.js`

Нужно добавить:
```javascript
// Получить средние значения по диапазону
export const fetchAverageData = async (startDate, endDate, interval, sites, indicators)

// Получить средние значения по району
export const fetchDistrictAverageData = async (district, startDate, endDate, interval)
```

### 2. Timeline - выбор диапазона точек
**Файл:** `front/src/components/Timeline.vue`

Текущее поведение: клик по одной точке
Нужно: клик по первой точке + клик по последней точке = диапазон

Добавить:
- State для выбора диапазона (rangeSelectionStart, rangeSelectionEnd)
- Визуальная индикация выбранного диапазона
- Emit события 'range-selected' с {start, end}

### 3. StatisticsModal - улучшенный график
**Файл:** `front/src/components/StatisticsModal.vue`

Текущий график: candlestick для мгновенных значений
Нужно: линейный график с markLine для временных рядов

Добавить:
- Props: timeSeriesData, dateRange, rangeType
- Линейный график (type: 'line')
- markLine для min/max
- Ось X: дни/месяцы/годы в зависимости от rangeType

### 4. App.vue - интеграция средних значений
**Файл:** `front/src/App.vue`

Добавить:
- State: selectedRangeType ('day' | 'month' | 'year')
- Функция loadAverageData(startDate, endDate, interval)
- Обработка выбора диапазона в Timeline
- Передача данных в StatisticsModal

### 5. MapView - клик по станции с диапазоном
**Файл:** `front/src/components/MapView.vue`

Текущее: emit sensor-click с мгновенными данными
Нужно: при клике загружать средние за выбранный период

### 6. SidePanel - средние по району
**Файл:** `front/src/components/SidePanel.vue`

Текущее: выбор датчиков района
Нужно: кнопка "Показать средние по району за период"

## Приоритет реализации:

**Высокий приоритет:**
1. API fetchAverageData - без этого ничего не работает
2. StatisticsModal - улучшенный график с временными рядами
3. Timeline - выбор диапазона точек

**Средний приоритет:**
4. App.vue - интеграция средних значений
5. MapView - клик по станции

**Низкий приоритет:**
6. SidePanel - средние по району (можно сделать позже)

## Проблемы текущей реализации:

### StatisticsModal.vue
- ❌ Использует candlestick вместо line chart
- ❌ Нет поддержки временных рядов
- ❌ Нет markLine для min/max
- ❌ Нет оси времени (дни/месяцы/годы)

### Timeline.vue
- ❌ Нет выбора диапазона точек
- ❌ Клик только по одной точке
- ❌ Нет визуальной индикации диапазона

### api.js
- ❌ Нет функции для средних значений
- ❌ fetchAggregatedData возвращает сырые данные, не средние

## Решение:

Буду реализовывать по приоритету, начиная с критичных компонентов.
