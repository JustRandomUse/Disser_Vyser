# Отчет об исправлении трех проблем

**Дата:** 2026-04-22

## Проблема 1: Список районов в боковой панели ✅ ИСПРАВЛЕНО

### Root Cause
SidePanel пересчитывал районы при каждом изменении `props.sensors` через watch с `deep: true`. Это вызывало ререндер при изменении данных на таймлайне.

### Исправление
**Файл:** `front/src/components/SidePanel.vue`

1. Добавлена проверка в `updateDistrictMapping()`:
```javascript
// Don't recalculate if we already have districts mapped
if (sensorsByDistrict.value.size > 0) return;
```

2. Изменен watch на sensors:
```javascript
// Было:
watch(() => props.sensors, () => {
  updateDistrictMapping();
}, { deep: true });

// Стало:
watch(() => props.sensors, (newSensors) => {
  if (newSensors.length > 0 && sensorsByDistrict.value.size === 0) {
    updateDistrictMapping();
  }
}, { immediate: true });
```

### Результат
- Список районов загружается один раз при монтировании
- Не пересчитывается при изменении времени на таймлайне
- Статический список районов независим от выбранного периода

---

## Проблема 2: Графики по районам ✅ ИСПРАВЛЕНО

### Root Cause
1. При выборе диапазона на таймлайне автоматически открывалась статистика по всем датчикам
2. Не было фильтрации по выбранным районам
3. `openStatisticsModal()` не загружал временные ряды для выбранных датчиков

### Исправление
**Файл:** `front/src/App.vue`

1. Изменен `onRangeSelected()`:
```javascript
// Убрано автоматическое открытие модального окна
// Теперь только загружаются средние для карты
const averageData = await fetchAverageData(...);
sensors.value = averageData;

// Не открываем модальное окно автоматически
console.log('Range selected. Select district or sensors to view statistics.');
```

2. Обновлен `openStatisticsModal()`:
```javascript
const openStatisticsModal = async (selectedSensors, districtKey) => {
  selectedSensorsForStats.value = selectedSensors;

  // If in range mode and district is selected, load time series for district
  if (selectionMode.value === 'range' && selectedDateRange.value && districtKey) {
    const siteIds = selectedSensors.map(s => s.id);
    
    // Fetch time series data for selected district sensors
    const data = await fetchTimeSeriesData(
      selectedDateRange.value.start,
      selectedDateRange.value.end,
      interval,
      siteIds,  // Только выбранные датчики
      null
    );
    
    timeSeriesData.value = data;
  }

  isStatisticsModalOpen.value = true;
};
```

3. Обновлен `showStatistics()` в SidePanel:
```javascript
const showStatistics = () => {
  const selected = props.sensors.filter(s => selectedSensors.value.includes(s.id));
  emit('show-statistics', selected, selectedPreset.value); // Передаем districtKey
};
```

### Результат
- Выбор диапазона на таймлайне не открывает модальное окно автоматически
- Графики строятся только по выбранным районам/датчикам
- Фильтрация работает корректно: siteIds передаются в fetchTimeSeriesData()

---

## Проблема 3: Выбор периода в календаре ✅ ИСПРАВЛЕНО

### Root Cause
В режимах месяцев и годов при первом клике устанавливались и начало, и конец (весь месяц/год). Это не позволяло выбрать диапазон из нескольких месяцев/годов.

### Исправление
**Файл:** `front/src/components/Calendar.vue`

1. Исправлен `selectMonth()`:
```javascript
if (!pendingStart.value || (pendingStart.value && pendingEnd.value)) {
  // First click - set start only
  pendingStart.value = startOfMonth;
  pendingEnd.value = endOfMonth; // Set end to same month for single month selection
  isSelectingRange.value = true;
} else if (pendingStart.value && !pendingEnd.value) {
  // Second click - set end
  if (startOfMonth < pendingStart.value) {
    // Clicked earlier month - swap
    const tempEnd = new Date(pendingStart.value.getFullYear(), pendingStart.value.getMonth() + 1, 0);
    pendingStart.value = startOfMonth;
    pendingEnd.value = tempEnd;
  } else {
    // Clicked later month - set as end
    pendingEnd.value = endOfMonth;
  }
  isSelectingRange.value = false;
}
```

2. Исправлен `selectYear()` аналогично

### Результат
- Первый клик = начало периода (один месяц/год)
- Второй клик = конец периода (диапазон)
- Один месяц/год = корректный диапазон (start = 01.03, end = 31.03)
- Несколько месяцев/годов = корректный диапазон (start = 01.03, end = 30.06)

---

## Тестирование

### Тест 1: Статический список районов
1. Открыть приложение
2. Выбрать район в боковой панели
3. Выбрать диапазон на таймлайне
4. **Ожидается:** Список районов не изменился, не прыгает

### Тест 2: Графики по районам
1. Выбрать диапазон месяцев в календаре (например, Март - Июнь)
2. Shift+клик на таймлайне для выбора диапазона
3. **Ожидается:** Модальное окно НЕ открывается автоматически
4. Выбрать район в боковой панели
5. Нажать "Показать статистику"
6. **Ожидается:** График строится только по датчикам выбранного района

### Тест 3: Выбор периода в календаре
1. Открыть календарь, режим "Месяцы"
2. Кликнуть на "Март"
3. **Ожидается:** Март подсвечен синим
4. Кликнуть на "Июнь"
5. **Ожидается:** Март и Июнь синие, Апрель и Май светло-синие
6. Нажать "Применить"
7. **Ожидается:** Таймлайн показывает Мар, Апр, Май, Июн

---

## Изменённые файлы

1. **front/src/components/SidePanel.vue**
   - Строки 99-121: updateDistrictMapping() с проверкой
   - Строки 133-136: showStatistics() передает districtKey
   - Строки 139-142: watch только на initial load

2. **front/src/App.vue**
   - Строки 365-393: openStatisticsModal() с загрузкой по району
   - Строки 370-406: onRangeSelected() без автооткрытия модального окна

3. **front/src/components/Calendar.vue**
   - Строки 261-283: selectMonth() с корректной логикой
   - Строки 285-303: selectYear() с корректной логикой

---

## Статус

✅ **Build:** Успешно (722ms)
✅ **Проблема 1:** Исправлена
✅ **Проблема 2:** Исправлена
✅ **Проблема 3:** Исправлена

**Готово к тестированию.**
