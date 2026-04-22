# Анализ текущей реализации календаря и таймлайна

## Проблемы текущей реализации

### 1. Calendar.vue - кнопка "Сегодня"

**Текущее поведение (строки 155-161):**
```javascript
const selectToday = () => {
  const today = new Date();
  currentMonth.value = today.getMonth();
  currentYear.value = today.getFullYear();
  pendingStart.value = today;
  pendingEnd.value = today;
};
```

**Проблема:** 
- ❌ Только устанавливает pending state
- ❌ Не эмитит событие
- ❌ Не закрывает календарь
- ❌ Требует нажатия "Применить"

**Нужно:**
- ✅ Сразу эмитить `date-selected`
- ✅ Закрыть календарь
- ✅ Быстрый сценарий без подтверждения

### 2. Calendar.vue - нет режимов выбора

**Текущее состояние:**
- Только day mode (выбор конкретных дней)
- Нет переключателя режимов
- Нет month mode
- Нет year mode

**Нужно добавить:**
- Переключатель режимов: day / month / year
- Month mode: выбор месяцев, преобразование в диапазон дат
- Year mode: выбор годов, преобразование в диапазон дат

### 3. App.vue - onTimeSelected привязан к hour

**Текущий код (строки 223-227):**
```javascript
const onTimeSelected = (timePoint) => {
  console.log('Time selected:', timePoint);
  selectedTimePoint.value = timePoint;
  loadData(selectedDate.value, timePoint.hour);  // ❌ Всегда использует .hour
};
```

**Проблема:**
- ❌ Работает только для hour scale
- ❌ Для day/month/year scale падает (нет поля .hour)
- ❌ Не учитывает timelineScale

**Нужно:**
- Проверять тип точки (hour/day/month/year)
- Загружать данные соответствующим образом

### 4. Timeline.vue - формат точек

**Текущие точки (App.vue строки 98-102, 108-112):**
```javascript
// Hour mode:
{ time: '00:00', color: '#5DADE2', hour: 0 }

// Day mode:
{ time: '1/4', color: '#5DADE2', date: Date }
```

**Проблема:**
- ❌ Нет единого формата
- ❌ Нет поля type
- ❌ Нет startDate/endDate для интервалов
- ❌ Формат времени непоследовательный

**Нужно:**
```javascript
{
  type: 'hour' | 'day' | 'month' | 'year',
  time: 'отображаемый текст',
  color: '#5DADE2',
  startDate: Date,
  endDate: Date,
  hour?: number  // только для hour type
}
```

### 5. generateTimePointsForRange - формат дат

**Текущий формат (строки 109, 121, 133):**
```javascript
time: `${current.getDate()}/${current.getMonth() + 1}`  // "1/4"
time: `${months[current.getMonth()]} ${current.getFullYear()}`  // "Апр 2026"
time: `${current.getFullYear()}`  // "2026"
```

**Проблема:**
- ❌ Day format неоднозначный (1/4 = 1 апреля или 4 января?)
- ❌ Нет года в day format
- ❌ Непоследовательный формат

**Нужно:**
```javascript
// Day: "01.04.2026" или "01.04"
// Month: "Апр 2026"
// Year: "2026"
```

## План реализации

### Этап 1: Исправить кнопку "Сегодня" в Calendar.vue

1. Изменить `selectToday()`:
   - Эмитить `date-selected` сразу
   - Закрыть календарь
   - Не требовать "Применить"

### Этап 2: Добавить режимы выбора в Calendar.vue

1. Добавить state `selectionMode: 'day' | 'month' | 'year'`
2. Добавить переключатель режимов в UI
3. Реализовать month mode:
   - Отображать сетку месяцев
   - Выбор месяца/диапазона месяцев
   - Преобразование в даты (начало/конец месяца)
4. Реализовать year mode:
   - Отображать сетку годов
   - Выбор года/диапазона лет
   - Преобразование в даты (начало/конец года)

### Этап 3: Унифицировать формат timeline points

1. Обновить `generateTimePoints()`:
   - Добавить type: 'hour'
   - Добавить startDate/endDate
2. Обновить `generateTimePointsForRange()`:
   - Добавить type для каждого scale
   - Добавить startDate/endDate
   - Улучшить формат time

### Этап 4: Исправить onTimeSelected в App.vue

1. Проверять `timePoint.type`
2. Для hour: `loadData(date, hour)`
3. Для day/month/year: `loadAggregatedData(startDate, endDate)`

### Этап 5: Улучшить форматирование дат

1. Создать helper функции:
   - `formatDay(date)` → "01.04.2026"
   - `formatMonth(date)` → "Апр 2026"
   - `formatYear(date)` → "2026"
   - `formatDateRange(start, end)` → "01.04.2026 — 05.04.2026"

## Ожидаемый результат

### UX сценарии:

**1. Кнопка "Сегодня":**
- Клик → календарь закрывается → загружаются данные за сегодня

**2. Выбор одного дня:**
- Day mode → выбрать день → "Применить" → таймлайн показывает часы

**3. Выбор диапазона дней:**
- Day mode → выбрать 2 дня → "Применить" → таймлайн показывает дни

**4. Выбор месяца:**
- Month mode → выбрать апрель 2026 → "Применить" → диапазон 01.04.2026–30.04.2026 → таймлайн показывает дни

**5. Выбор диапазона месяцев:**
- Month mode → выбрать март–июнь → "Применить" → диапазон 01.03.2026–30.06.2026 → таймлайн показывает месяцы

**6. Выбор года:**
- Year mode → выбрать 2025 → "Применить" → диапазон 01.01.2025–31.12.2025 → таймлайн показывает месяцы

**7. Выбор диапазона лет:**
- Year mode → выбрать 2022–2024 → "Применить" → диапазон 01.01.2022–31.12.2024 → таймлайн показывает годы
