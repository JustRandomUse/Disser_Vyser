<template>
  <div class="legend">
    <h3>Индекс качества воздуха</h3>
    <div class="legend-items">
      <div v-for="item in legendItems" :key="item.level" class="legend-item">
        <div class="legend-color" :style="{ backgroundColor: item.color }"></div>
        <div class="legend-text">
          <div class="legend-level">{{ item.level }}</div>
          <div class="legend-range">{{ item.range }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Legend',
  props: {
    selectedIndicator: {
      type: Number,
      default: 1
    }
  },
  computed: {
    legendItems() {
      const legends = {
        0: [ // AQI
          { level: '0-50', range: '0-50', color: '#00e400' },
          { level: '51-100', range: '51-100', color: '#ffff00' },
          { level: '101-150', range: '101-150', color: '#ff7e00' },
          { level: '151-200', range: '151-200', color: '#ff0000' },
          { level: '201-300', range: '201-300', color: '#8f3f97' },
          { level: '301+', range: '301+', color: '#7e0023' }
        ],
        1: [ // PM2.5
          { level: '0-12', range: '0-12', color: '#00e400' },
          { level: '12-35', range: '12-35', color: '#ffff00' },
          { level: '35-55', range: '35-55', color: '#ff7e00' },
          { level: '55-150', range: '55-150', color: '#ff0000' },
          { level: '150-250', range: '150-250', color: '#8f3f97' },
          { level: '250+', range: '250+', color: '#7e0023' }
        ],
        2: [ // PM10
          { level: '0-54', range: '0-54', color: '#00e400' },
          { level: '55-154', range: '55-154', color: '#ffff00' },
          { level: '155-254', range: '155-254', color: '#ff7e00' },
          { level: '255-354', range: '255-354', color: '#ff0000' },
          { level: '355-424', range: '355-424', color: '#8f3f97' },
          { level: '425+', range: '425+', color: '#7e0023' }
        ],
        3: [ // Temperature
          { level: '-30', range: '-30', color: '#2E86C1' },
          { level: '-20', range: '-20', color: '#3498DB' },
          { level: '-10', range: '-10', color: '#5DADE2' },
          { level: '0', range: '0', color: '#85C1E9' },
          { level: '10', range: '10', color: '#D5DBDB' },
          { level: '20', range: '20', color: '#F7DC6F' },
          { level: '30', range: '30', color: '#F0B27A' },
          { level: '40', range: '40', color: '#E74C3C' }
        ],
        4: [ // Humidity
          { level: '0-20', range: '0-20', color: '#8B4513' },
          { level: '20-40', range: '20-40', color: '#F4A460' },
          { level: '40-60', range: '40-60', color: '#87CEEB' },
          { level: '60-80', range: '60-80', color: '#4682B4' },
          { level: '80-100', range: '80-100', color: '#191970' }
        ],
        5: [ // Pressure
          { level: '980-990', range: '980-990', color: '#8B0000' },
          { level: '990-1000', range: '990-1000', color: '#FF6347' },
          { level: '1000-1010', range: '1000-1010', color: '#FFD700' },
          { level: '1010-1020', range: '1010-1020', color: '#90EE90' },
          { level: '1020-1030', range: '1020-1030', color: '#4169E1' },
          { level: '1030+', range: '1030+', color: '#4B0082' }
        ]
      };

      return legends[this.selectedIndicator] || legends[1];
    }
  }
};
</script>

<style scoped>
.legend {
  position: absolute;
  bottom: 68px;
  right: 8px;
  background: rgba(0, 0, 0, 0.3);
  padding: 10px;
  border-radius: 12px;
  z-index: 1;
  display: flex;
  flex-direction: column;
  color: white;
}

.legend h3 {
  display: none;
}

.legend-items {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 15px;
}

.legend-color {
  width: 45px;
  height: 35px;
  border-radius: 0;
  flex-shrink: 0;
}

.legend-text {
  display: flex;
  flex-direction: column;
}

.legend-level {
  display: none;
}

.legend-range {
  font-size: 17px;
  color: white;
  font-weight: 500;
}
</style>
