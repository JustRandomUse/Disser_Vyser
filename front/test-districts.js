// Test script for district mapping validation
import { getDistrictForSensor, groupSensorsByDistrict, validateDistrictMapping } from './src/utils/districtMapping.js';

// Test sensors with real coordinates from API
const testSensors = [
  { id: 3837, name: "Академгородок", longitude: 92.762222, latitude: 55.986571 },
  { id: 3853, name: "Ветлужанка", longitude: 92.76537, latitude: 56.0293 },
  { id: 3844, name: "Дрокино", longitude: 92.7427, latitude: 56.0987 },
  { id: 4363, name: "Караульная", longitude: 92.674777, latitude: 55.973219 },
  { id: 4325, name: "Качинская, 56А", longitude: 92.871068, latitude: 56.018823 },
  { id: 3859, name: "Кировский", longitude: 92.958275, latitude: 55.994686 },
  { id: 3849, name: "КрАЗ", longitude: 93.0148, latitude: 56.0878 },
  { id: 3857, name: "Ленина, 41", longitude: 92.877796, latitude: 56.013546 },
  { id: 3847, name: "Минино", longitude: 92.70456, latitude: 56.07743 },
  { id: 3842, name: "Николаевка", longitude: 92.82126, latitude: 56.00772 },
  { id: 4227, name: "о. Молокова", longitude: 92.90986, latitude: 56.00716 },
  { id: 4228, name: "о. Татышев", longitude: 92.943622, latitude: 56.022003 },
  { id: 3850, name: "Овинный-Таймыр", longitude: 92.705374, latitude: 56.045544 },
  { id: 3846, name: "Партизана, 3г", longitude: 92.910795, latitude: 56.024966 },
  { id: 3856, name: "Песчанка", longitude: 93.084937, latitude: 56.081751 },
  { id: 3855, name: "Покровка", longitude: 92.86059, latitude: 56.02562 },
  { id: 4364, name: "Посадный", longitude: 92.860259, latitude: 56.002347 },
  { id: 4324, name: "Руслов", longitude: 92.82215, latitude: 55.99284 },
  { id: 3851, name: "Свердловский", longitude: 92.854022, latitude: 55.976276 },
  { id: 3852, name: "Светлый", longitude: 92.859047, latitude: 56.089951 },
  { id: 3915, name: "Солонцы", longitude: 92.843109, latitude: 56.066574 },
  { id: 3858, name: "Спутник", longitude: 92.97329, latitude: 56.01204 },
  { id: 3843, name: "Телевизорная, 1/31", longitude: 92.799131, latitude: 56.02513 },
  { id: 3854, name: "Удачный", longitude: 92.677273, latitude: 55.978387 },
  { id: 3841, name: "Шахтеров, 25", longitude: 92.886663, latitude: 56.032797 }
];

console.log('=== District Mapping Test ===\n');

// Test individual sensors
console.log('Individual sensor tests:');
testSensors.slice(0, 5).forEach(sensor => {
  const district = getDistrictForSensor(sensor.longitude, sensor.latitude);
  console.log(`${sensor.name}: ${district ? district.name : 'Без района'}`);
});

console.log('\n=== Validation Report ===');
const report = validateDistrictMapping(testSensors);
console.log(JSON.stringify(report, null, 2));

console.log('\n=== Sensors by District ===');
const grouped = groupSensorsByDistrict(testSensors);
grouped.forEach((sensors, districtKey) => {
  if (sensors.length > 0) {
    console.log(`\n${districtKey}:`);
    sensors.forEach(s => console.log(`  - ${s.name}`));
  }
});
