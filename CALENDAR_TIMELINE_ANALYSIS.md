# Анализ текущей реализации календаря и таймлайна

## Проблемы текущей реализации

### 1. Calendar.vue
**Проблемы:**
- ❌ Нет кнопки "Применить" — диапазон эмитится сразу при выборе второй даты
- ❌ Кнопка "Сегодня" закрывает календарь сразу без подтверждения
- ❌ Нет явного UX для подтверждения выбора диапазона
- ❌ Пользователь не может отменить выбор без закрытия календаря

**Текущий код (Calendar.vue:132-155):**
```javascript
const selectDate = (day) => {
  if (day.otherMonth) return;
  const clickedDate = new Date(day.year, day.month, day.date);
  
  if (!startDate.value || (startDate.value && endDate.value)) {
    startDate.value = clickedDate;
    endDate.value = null;
    isSelectingRange.value = true;
  } else if (startDate.value && !endDate.value) {
    if (clickedDate < startDate.value) {
      endDate.value = startDate.value;
      startDate.value = clickedDate;
    } else {
      endDate.value = clickedDate;
    }
    isSelectingRange.value = false;
    
    emit('date-range-selected', {  // ❌ Эмитит сразу!
      start: startDate.value,
      end: endDate.value
    });
  }
};
```

### 2. App.vue
**Проблемы:**
- ❌ `onDateRangeSelected` использует только `range.start` для загрузки данных
- ❌ `loadAggregatedData` — заглушка, грузит данные только за первый день
- ❌ `generateTimePoints()` всегда генерирует 24 часа, не адаптируется под диапазон
- ❌ `setInterval(loadData, 300000)` вызывает `loadData()` без параметров, сбрасывая выбранный диапазон
- ❌ Нет состояния для режима (live vs historical)
- ❌ Нет отдельного хранения `selectedRangeStart` и `selectedRangeEnd`

**Текущий код (App.vue:130-142):**
```javascript
const loadAggregatedData = async (startDate, endDate) => {
  try {
    console.log('Loading aggregated data for range:', startDate, 'to', endDate);
    // For now, load data for the start date
    // TODO: Implement proper aggregated data loading from backend  // ❌ TODO!
    const data = await fetchAirQualityData(startDate, 12);  // ❌ Только startDate!
    console.log('Loaded aggregated sensor data:', data);
    sensors.value = data;
  } catch (error) {
    console.error('Failed to load aggregated data:', error);
    sensors.value = [];
  }
};
```

**Текущий код (App.vue:220-224):**
```javascript
onMounted(() => {
  generateTimePoints();
  loadData();
  setInterval(loadData, 300000);  // ❌ Сбрасывает выбранный диапазон!
});
```

### 3. Timeline.vue
**Проблемы:**
- ❌ Всегда показывает часы, не адаптируется под диапазон
- ❌ Нет логики переключения scale (hour/day/month/year)
- ❌ Получает `timePoints` из App.vue, но они всегда 24 часа
- ❌ `formattedDate` показывает только одну дату, не диапазон

### 4. api.js
**Проблемы:**
- ❌ `fetchAirQualityData` не поддерживает диапазон дат
- ❌ Нет функции для загрузки агрегированных данных за период

**Текущий код (api.js:38-50):**
```javascript
export const fetchAirQualityData = async (date = null, hour = null) => {
  // ...
  let endpoint = `${API_BASE_URL}/datasets/knc-air/last`;
  
  if (date && hour !== null) {
    const dateStr = date.toISOString().split('T')[0];
    const hourStr = hour.toString().padStart(2, '0');
    endpoint = `${API_BASE_URL}/datasets/knc-air/data?date=${dateStr}&hour=${hourStr}`;
  }
  // ❌ Нет поддержки диапазона!
}
```

## Доступные API endpoints (из API.md)

### Для одиночной даты/часа:
- `/sets/<code>/data/last` — актуальные данные (live)
- `/sets/<code>/data/raw?time_begin=...&time_end=...` — сырые данные за период

### Для диапазона дат:
- `/sets/<code>/data/archive?time_begin=...&time_end=...&time_interval={hour|day|month}` — агрегированные данные
- `/sets/<code>/data/archive-ext?...` — агрегированные данные со статистикой

**Важно:** API поддерживает диапазоны через `time_begin` и `time_end`!

## План реализации

### Этап 1: Исправить Calendar.vue
1. Добавить состояние `pendingStart` и `pendingEnd` для временного выбора
2. Заменить кнопки "Сегодня" и "Закрыть" на "Отмена" и "Применить"
3. Эмитить `date-range-selected` только при нажатии "Применить"
4. Добавить визуальную индикацию выбранного диапазона

### Этап 2: Исправить App.vue
1. Добавить состояние:
   - `mode` — 'live' | 'single-day' | 'range'
   - `rangeStart` и `rangeEnd`
   - `timelineScale` — 'hour' | 'day' | 'month' | 'year'
2. Создать `generateTimePointsForRange(start, end)` — динамическая генерация
3. Исправить `loadAggregatedData` для реальной загрузки диапазона
4. Исправить `setInterval` — проверять режим перед автообновлением
5. Добавить логику определения scale по размеру диапазона

### Этап 3: Исправить Timeline.vue
1. Добавить prop `dateRange` для отображения диапазона
2. Обновить `formattedDate` для показа диапазона
3. Адаптировать под разные scale (hour/day/month/year)

### Этап 4: Расширить api.js
1. Добавить `fetchAggregatedData(startDate, endDate, interval)`
2. Использовать `/api/datasets/:code/aggregated` endpoint

## Пороги для scale

```javascript
const daysDiff = Math.ceil((endDate - startDate) / (1000 * 60 * 60 * 24));

if (daysDiff === 0) {
  scale = 'hour';  // 1 день → часы
} else if (daysDiff <= 31) {
  scale = 'day';   // 2-31 день → дни
} else if (daysDiff <= 365) {
  scale = 'month'; // 32-365 дней → месяцы
} else {
  scale = 'year';  // > 365 дней → годы
}
```

## Backend follow-up

**Текущий backend endpoint:**
- `/api/datasets/:code/data?date=YYYY-MM-DD&hour=HH` — только одиночный час

**Нужно добавить:**
- `/api/datasets/:code/aggregated?time_begin=...&time_end=...&interval={hour|day|month}` — диапазон

Или использовать существующий `/api/datasets/:code/aggregated` (уже есть в handler.go).
