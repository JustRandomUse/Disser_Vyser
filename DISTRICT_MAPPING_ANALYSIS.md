# Анализ текущей реализации районов

## Текущее состояние

### SidePanel.vue (строки 81-91)
```javascript
const presets = {
  left_bank: ['Черемушки'],
  nikolaevka: ['Николаевка'],
  center: ['Центр', 'Площадь Революции'],
  soviet: ['Советский', 'Академгородок', 'Взлетка'],
  zheleznodorozhny: ['Вокзал'],
  kirovsky: ['Кировский'],
  leninsky: ['Ленинский'],
  oktyabrsky: ['Октябрьский'],
  sverdlovsky: ['Свердловский']
};
```

**Проблемы:**
- ❌ Хардкод названий станций
- ❌ Нет географической привязки
- ❌ Exact match по `sensor.name === name` (строка 117)
- ❌ Если название станции изменится в API — сломается привязка
- ❌ Невозможно добавить новую станцию без изменения кода

### api.js (строки 20-26)
```javascript
coordinates[site.id] = {
  lat: site.geom_y,
  lon: site.geom_x,
  name: site.name,
  code: site.code
};
```

**Хорошо:**
- ✅ Координаты уже приходят из API
- ✅ Используется WGS84 (geom_x/geom_y)
- ✅ Есть кэширование координат

### MapView.vue
- ✅ Использует OpenLayers
- ✅ Отображает станции по координатам
- ✅ Нет привязки к районам (только визуализация)

## План реализации

### Этап 1: Получить границы районов Красноярска
Источник: OpenStreetMap via Overpass API

Районы Красноярска (административные):
1. Железнодорожный район (relation/1215893)
2. Кировский район (relation/1215894)
3. Ленинский район (relation/1215895)
4. Октябрьский район (relation/1215896)
5. Советский район (relation/1215897)
6. Свердловский район (relation/1215898)
7. Центральный район (relation/1215899)

### Этап 2: Создать GeoJSON файл
- Формат: FeatureCollection
- CRS: EPSG:4326 (WGS84)
- Свойства каждого района:
  - name: официальное название
  - key: ключ для кода (zheleznodorozhny, kirovsky, etc.)
  - type: "district"

### Этап 3: Реализовать утилиту districtMapping
- Использовать @turf/boolean-point-in-polygon
- Функция: `getDistrictForSensor(lat, lon) => { key, name } | null`
- Функция: `groupSensorsByDistrict(sensors) => Map<districtKey, sensors[]>`

### Этап 4: Обновить SidePanel.vue
- Убрать хардкод presets
- Загружать список районов из GeoJSON
- Фильтровать станции по геометрии, а не по названию
- Добавить категорию "Без района" для станций вне полигонов

### Этап 5: Валидация
- Проверить, что все станции попали в районы
- Проверить, что нет станций в нескольких районах одновременно
- Вывести summary в консоль при загрузке

## Ожидаемый результат

**Было:**
```javascript
presets.soviet.includes('Академгородок') // true
```

**Станет:**
```javascript
isPointInPolygon([92.8672, 56.0153], sovietDistrict.geometry) // true
```
