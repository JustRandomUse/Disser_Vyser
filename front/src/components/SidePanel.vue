<template>
  <div class="side-panel" :class="{ 'collapsed': isCollapsed }">
    <button class="toggle-btn" @click="togglePanel">
      {{ isCollapsed ? '◀' : '▶' }}
    </button>

    <div v-if="!isCollapsed" class="panel-content">
      <h3>Датчики</h3>

      <div class="preset-section">
        <label>Пресеты районов:</label>
        <select v-model="selectedPreset" @change="applyPreset" class="preset-select">
          <option value="">Выбрать район</option>
          <option value="left_bank">Левый берег</option>
          <option value="nikolaevka">Николаевка</option>
          <option value="center">Центр</option>
          <option value="soviet">Советский район</option>
          <option value="zheleznodorozhny">Железнодорожный район</option>
          <option value="kirovsky">Кировский район</option>
          <option value="leninsky">Ленинский район</option>
          <option value="oktyabrsky">Октябрьский район</option>
          <option value="sverdlovsky">Свердловский район</option>
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

<script>
export default {
  name: 'SidePanel',
  props: {
    sensors: {
      type: Array,
      default: () => []
    }
  },
  data() {
    return {
      isCollapsed: false,
      selectedPreset: '',
      selectedSensors: [],
      presets: {
        left_bank: ['Черемушки', 'Взлетка', 'Академгородок'],
        nikolaevka: ['Николаевка'],
        center: ['Центр', 'Площадь Революции'],
        soviet: ['Советский', 'Академгородок', 'Взлетка'],
        zheleznodorozhny: ['Вокзал', 'Черемушки'],
        kirovsky: ['Кировский', 'Николаевка'],
        leninsky: ['Ленинский', 'Центр'],
        oktyabrsky: ['Октябрьский'],
        sverdlovsky: ['Свердловский']
      }
    };
  },
  computed: {
    allSelected() {
      return this.sensors.length > 0 && this.selectedSensors.length === this.sensors.length;
    }
  },
  methods: {
    togglePanel() {
      this.isCollapsed = !this.isCollapsed;
    },
    toggleAll() {
      if (this.allSelected) {
        this.selectedSensors = [];
      } else {
        this.selectedSensors = this.sensors.map(s => s.id);
      }
    },
    applyPreset() {
      if (!this.selectedPreset) {
        this.selectedSensors = [];
        return;
      }

      const presetNames = this.presets[this.selectedPreset] || [];
      this.selectedSensors = this.sensors
        .filter(sensor => presetNames.some(name => sensor.name.includes(name)))
        .map(s => s.id);
    },
    showStatistics() {
      const selected = this.sensors.filter(s => this.selectedSensors.includes(s.id));
      this.$emit('show-statistics', selected);
    }
  }
};
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
