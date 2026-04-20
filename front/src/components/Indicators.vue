<template>
  <div class="indicators-dropdown">
    <div class="dropdown-header" @click="toggleDropdown">
      <span>{{ selectedIndicator.name }}</span>
      <span class="arrow">{{ isOpen ? '▲' : '▼' }}</span>
    </div>
    <div v-if="isOpen" class="dropdown-list">
      <div
        v-for="(indicator, index) in indicators"
        :key="index"
        class="dropdown-item"
        :class="{ 'selected': selectedIndex === index }"
        @click="selectIndicator(index)"
      >
        {{ indicator.name }}
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Indicators',
  props: {
    selectedIndex: {
      type: Number,
      default: 8
    }
  },
  data() {
    return {
      isOpen: false,
      indicators: [
        { name: 'InstantAQI', key: 'aqi' },
        { name: 'PM2.5 (мкг/м³)', key: 'pm25' },
        { name: 'PM10 (мкг/м³)', key: 'pm10' },
        { name: 'Температура (°С)', key: 'temperature' },
        { name: 'Влажность (%)', key: 'humidity' },
        { name: 'Давление (гПа)', key: 'pressure' }
      ]
    };
  },
  computed: {
    selectedIndicator() {
      return this.indicators[this.selectedIndex] || this.indicators[0];
    }
  },
  methods: {
    toggleDropdown() {
      this.isOpen = !this.isOpen;
    },
    selectIndicator(index) {
      this.$emit('indicator-selected', index);
      this.isOpen = false;
    }
  }
};
</script>

<style scoped>
.indicators-dropdown {
  position: absolute;
  bottom: 68px;
  left: 8px;
  z-index: 2;
  min-width: 250px;
}

.dropdown-header {
  background: rgba(0, 0, 0, 0.7);
  color: white;
  padding: 12px 16px;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 14px;
  font-weight: 500;
  transition: background 0.2s;
}

.dropdown-header:hover {
  background: rgba(0, 0, 0, 0.85);
}

.arrow {
  font-size: 12px;
  margin-left: 10px;
}

.dropdown-list {
  background: rgba(0, 0, 0, 0.85);
  border-radius: 8px;
  margin-top: 5px;
  max-height: 300px;
  overflow-y: auto;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.dropdown-item {
  color: white;
  padding: 10px 16px;
  cursor: pointer;
  font-size: 14px;
  transition: background 0.2s;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.dropdown-item:last-child {
  border-bottom: none;
}

.dropdown-item:hover {
  background: rgba(255, 255, 255, 0.1);
}

.dropdown-item.selected {
  background: #3b82f6;
  font-weight: 500;
}
</style>
