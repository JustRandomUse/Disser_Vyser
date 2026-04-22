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

      <button
        class="show-stats-btn"
        :disabled="selectedSensors.length === 0"
        @click="showStatistics"
      >
        Показать статистику ({{ selectedSensors.length }})
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue';
import { getAvailableDistricts, groupSensorsByDistrict, validateDistrictMapping } from '../utils/districtMapping';

const props = defineProps({
  sensors: {
    type: Array,
    default: () => []
  }
});

const emit = defineEmits(['show-statistics']);

const isCollapsed = ref(false);
const selectedPreset = ref('');
const selectedSensors = ref([]);
const availableDistricts = ref([]);
const sensorsByDistrict = ref(new Map());

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
  emit('show-statistics', selected, selectedPreset.value);
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
</style>
