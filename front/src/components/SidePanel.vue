<template>
  <div class="side-panel" :class="{ 'collapsed': isCollapsed }">
    <button class="toggle-btn" @click="togglePanel">
      {{ isCollapsed ? '◀' : '▶' }}
    </button>

    <div v-if="!isCollapsed" class="panel-content">
      <h3>Датчики</h3>

      <div class="preset-section">
        <label>Районы города:</label>
        <select v-model="selectedPreset" @change="applyPreset" class="preset-select">
          <option value="">Выбрать район</option>
          <option
            v-for="district in availableDistricts"
            :key="district.key"
            :value="district.key"
          >
            {{ district.name }}
          </option>
          <option value="unassigned">Без района</option>
        </select>
      </div>

      <div class="sensors-list">
        <div class="select-all">
          <label>
            <input
              type="checkbox"
              :checked="allSelected"
              @change="toggleAll"
            />
            <span>Выбрать все</span>
          </label>
        </div>

        <div
          v-for="sensor in sensors"
          :key="sensor.id"
          class="sensor-item"
        >
          <label>
            <input
              type="checkbox"
              :value="sensor.id"
              v-model="selectedSensors"
            />
            <span>{{ sensor.name }}</span>
          </label>
        </div>
      </div>

      <div class="date-section">
        <label>Период для статистики:</label>
        <button @click="isCalendarOpen = true" class="date-select-btn">
          {{ dateRangeText }}
        </button>
      </div>

      <div class="display-mode-section">
        <label class="switch-label">
          <input
            type="checkbox"
            v-model="showIndividualData"
            class="switch-checkbox"
          />
          <span class="switch-slider"></span>
          <span class="switch-text">{{ showIndividualData ? 'Показать по отдельности' : 'Показать средние' }}</span>
        </label>
      </div>

      <button
        class="show-stats-btn"
        :disabled="selectedSensors.length === 0"
        @click="showStatistics"
      >
        Показать статистику ({{ selectedSensors.length }})
      </button>
    </div>

    <Calendar
      :isOpen="isCalendarOpen"
      :selectedDate="selectedDate"
      :selectedDateRange="selectedDateRange"
      @date-selected="onDateSelected"
      @date-range-selected="onDateRangeSelected"
      @close="isCalendarOpen = false"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue';
import { getAvailableDistricts, groupSensorsByDistrict, validateDistrictMapping } from '../utils/districtMapping';
import Calendar from './Calendar.vue';
import { formatDateISO, formatDateRangeISO } from '../utils/dateFormat';

const props = defineProps({
  sensors: {
    type: Array,
    default: () => []
  },
  selectedDate: {
    type: Date,
    default: null
  },
  selectedDateRange: {
    type: Object,
    default: null
  }
});

const emit = defineEmits(['show-statistics', 'calendar-date-changed']);

const isCollapsed = ref(false);
const selectedPreset = ref('');
const selectedSensors = ref([]);
const availableDistricts = ref([]);
const sensorsByDistrict = ref(new Map());
const isCalendarOpen = ref(false);
const selectedDate = ref(new Date());
const selectedDateRange = ref(null);
const showIndividualData = ref(false);

// Watch for external date changes from Timeline
watch(() => props.selectedDate, (newDate) => {
  if (newDate) {
    selectedDate.value = newDate;
  }
}, { immediate: true });

watch(() => props.selectedDateRange, (newRange) => {
  if (newRange) {
    selectedDateRange.value = newRange;
  }
}, { immediate: true });

const dateRangeText = computed(() => {
  // Priority: props.selectedDateRange > local selectedDateRange > props.selectedDate > local selectedDate
  if (props.selectedDateRange && props.selectedDateRange.start && props.selectedDateRange.end) {
    return formatDateRangeISO(props.selectedDateRange.start, props.selectedDateRange.end);
  }
  if (selectedDateRange.value && selectedDateRange.value.start && selectedDateRange.value.end) {
    return formatDateRangeISO(selectedDateRange.value.start, selectedDateRange.value.end);
  }
  if (props.selectedDate) {
    return formatDateISO(props.selectedDate);
  }
  if (selectedDate.value) {
    return formatDateISO(selectedDate.value);
  }
  // Fallback to today
  return formatDateISO(new Date());
});

const allSelected = computed(() => {
  return props.sensors.length > 0 && selectedSensors.value.length === props.sensors.length;
});

const togglePanel = () => {
  isCollapsed.value = !isCollapsed.value;
};

const toggleAll = () => {
  if (allSelected.value) {
    selectedSensors.value = [];
  } else {
    selectedSensors.value = props.sensors.map(s => s.id);
  }
};

const updateDistrictMapping = async () => {
  // Only update if we don't have districts yet or sensors list is empty
  if (props.sensors.length === 0) return;

  // Don't recalculate if we already have districts mapped
  if (sensorsByDistrict.value.size > 0) return;

  try {
    sensorsByDistrict.value = await groupSensorsByDistrict(props.sensors);

    // Log validation report
    const report = await validateDistrictMapping(props.sensors);
    console.log('District mapping validation:', report);

    if (report.unassigned > 0) {
      console.warn(`${report.unassigned} sensors without district assignment`);
    }

    if (report.multipleDistricts.length > 0) {
      console.warn('Sensors in multiple districts:', report.multipleDistricts);
    }
  } catch (error) {
    console.error('Failed to update district mapping:', error);
    // Fallback: empty map
    sensorsByDistrict.value = new Map();
  }
};

const applyPreset = () => {
  if (!selectedPreset.value) {
    selectedSensors.value = [];
    return;
  }

  const districtSensors = sensorsByDistrict.value.get(selectedPreset.value) || [];
  selectedSensors.value = districtSensors.map(s => s.id);
};

const showStatistics = () => {
  const selected = props.sensors.filter(s => selectedSensors.value.includes(s.id));
  emit('show-statistics', selected, selectedPreset.value, selectedDateRange.value, showIndividualData.value);
};

const onDateSelected = (date) => {
  console.log('📅 SidePanel onDateSelected:', date);
  selectedDate.value = date;
  selectedDateRange.value = null;
  isCalendarOpen.value = false;

  // Emit calendar change for SensorModal
  const dateRange = {
    start: new Date(date.getFullYear(), date.getMonth(), date.getDate(), 0, 0, 0),
    end: new Date(date.getFullYear(), date.getMonth(), date.getDate(), 23, 59, 59)
  };
  console.log('📅 SidePanel emitting dateRange:', dateRange);
  emit('calendar-date-changed', dateRange);
};

const onDateRangeSelected = (range) => {
  console.log('📅 SidePanel onDateRangeSelected:', range);
  selectedDateRange.value = range;
  isCalendarOpen.value = false;

  // Emit calendar change for SensorModal
  console.log('📅 SidePanel emitting range:', range);
  emit('calendar-date-changed', range);
};

// Watch for sensor changes only on initial load
watch(() => props.sensors, (newSensors) => {
  if (newSensors.length > 0 && sensorsByDistrict.value.size === 0) {
    updateDistrictMapping();
  }
}, { immediate: true });

onMounted(async () => {
  try {
    availableDistricts.value = await getAvailableDistricts();
    await updateDistrictMapping();
  } catch (error) {
    console.error('Failed to initialize districts:', error);
    // Fallback: empty districts
    availableDistricts.value = [];
  }
});
</script>

<style scoped>
.side-panel {
  position: fixed;
  top: 0;
  right: 0;
  width: 320px;
  height: 100vh;
  background: rgba(255, 255, 255, 0.95);
  box-shadow: -2px 0 10px rgba(0, 0, 0, 0.1);
  transition: transform 0.3s ease;
  z-index: 1500;
  display: flex;
  flex-direction: column;
}

.side-panel.collapsed {
  transform: translateX(320px);
}

.toggle-btn {
  position: absolute;
  left: -40px;
  top: 50%;
  transform: translateY(-50%);
  width: 40px;
  height: 80px;
  background: rgba(255, 255, 255, 0.95);
  border: none;
  border-radius: 8px 0 0 8px;
  cursor: pointer;
  font-size: 20px;
  box-shadow: -2px 0 10px rgba(0, 0, 0, 0.1);
  transition: background 0.2s;
}

.toggle-btn:hover {
  background: rgba(255, 255, 255, 1);
}

.panel-content {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}

h3 {
  margin: 0 0 20px 0;
  color: #333;
  font-size: 20px;
}

.preset-section {
  margin-bottom: 20px;
}

.preset-section label {
  display: block;
  margin-bottom: 8px;
  color: #555;
  font-size: 14px;
  font-weight: 500;
}

.preset-select {
  width: 100%;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 14px;
  background: white;
  cursor: pointer;
}

.preset-select:focus {
  outline: none;
  border-color: #3b82f6;
}

.sensors-list {
  max-height: calc(100vh - 300px);
  overflow-y: auto;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 10px;
  background: #f9fafb;
}

.select-all {
  padding: 10px;
  border-bottom: 2px solid #ddd;
  margin-bottom: 10px;
}

.select-all label {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  font-weight: 600;
  color: #333;
}

.sensor-item {
  padding: 8px 10px;
  border-bottom: 1px solid #e5e7eb;
}

.sensor-item:last-child {
  border-bottom: none;
}

.sensor-item label {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  color: #555;
  font-size: 14px;
}

.sensor-item:hover {
  background: #f3f4f6;
}

input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.show-stats-btn {
  width: 100%;
  margin-top: 20px;
  padding: 12px;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.show-stats-btn:hover:not(:disabled) {
  background: #2563eb;
}

.show-stats-btn:disabled {
  background: #9ca3af;
  cursor: not-allowed;
}

.date-section {
  margin-bottom: 20px;
}

.date-section label {
  display: block;
  margin-bottom: 8px;
  color: #555;
  font-size: 14px;
  font-weight: 500;
}

.date-select-btn {
  width: 100%;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 14px;
  background: white;
  cursor: pointer;
  text-align: left;
  transition: border-color 0.2s;
}

.date-select-btn:hover {
  border-color: #3b82f6;
}

.display-mode-section {
  margin-bottom: 20px;
}

.switch-label {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  user-select: none;
}

.switch-checkbox {
  display: none;
}

.switch-slider {
  position: relative;
  width: 48px;
  height: 24px;
  background: #ccc;
  border-radius: 24px;
  transition: background 0.3s;
  flex-shrink: 0;
}

.switch-slider::before {
  content: '';
  position: absolute;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: white;
  top: 2px;
  left: 2px;
  transition: transform 0.3s;
}

.switch-checkbox:checked + .switch-slider {
  background: #3b82f6;
}

.switch-checkbox:checked + .switch-slider::before {
  transform: translateX(24px);
}

.switch-text {
  color: #555;
  font-size: 14px;
  font-weight: 500;
}
</style>
