<template>
  <div v-if="isOpen" class="modal-backdrop" @click="closeModal">
    <div class="modal-content" @click.stop>
      <button class="close-btn" @click="closeModal">&times;</button>

      <h2>Датчик: {{ sensorData.name || 'Неизвестно' }}</h2>

      <div class="data-section">
        <h3>Текущие измерения</h3>
        <div class="current-values-chart">
          <div ref="currentValuesChart" style="width: 100%; height: 300px;"></div>
        </div>

        <table class="data-table">
          <thead>
            <tr>
              <th>Параметр</th>
              <th>Значение</th>
              <th>Единица</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(value, key) in measurements"
              :key="key"
              @click="selectSingleParam(key)"
              class="clickable-row"
              :class="{ 'selected-row': selectedSingleParam === key }"
            >
              <td>{{ formatKey(key) }}</td>
              <td>{{ value }}</td>
              <td>{{ getUnit(key) }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="charts-section">
        <div class="chart-container-large">
          <div ref="timeSeriesChart" style="width: 100%; height: 100%;"></div>
        </div>

        <div class="comparison-section">
          <h3>Сравнение параметров</h3>
          <div class="comparison-checkboxes">
            <label v-for="(value, key) in measurements" :key="key" class="checkbox-label">
              <input
                type="checkbox"
                :checked="comparisonParams.includes(key)"
                @change="toggleComparisonParam(key)"
                class="param-checkbox"
              />
              <span>{{ formatKey(key) }}</span>
            </label>
          </div>
          <div class="chart-container-comparison">
            <div ref="comparisonChart" style="width: 100%; height: 100%;"></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

          boundaryGap: false
        },
        yAxis: yAxisConfig,
        dataZoom: [
          {
            type: 'inside',
            start: 0,
            end: 100
          },
          {
            start: 0,
            end: 100
          }
        ],
        series: series
      };

      this.comparisonChart.setOption(option);
    }
  },
  beforeUnmount() {
    if (this.timeSeriesChart) {
      this.timeSeriesChart.dispose();
    }
    if (this.comparisonChart) {
      this.comparisonChart.dispose();
    }
    if (this.currentValuesChart) {
      this.currentValuesChart.dispose();
    }
  }
};
</script>

<style scoped>
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 2000;
}

.modal-content {
  background: white;
  border-radius: 12px;
  padding: 30px;
  max-width: 1200px;
  width: 90vw;
  max-height: 90vh;
  overflow-y: auto;
  position: relative;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

.close-btn {
  position: absolute;
  top: 15px;
  right: 15px;
  background: none;
  border: none;
  font-size: 32px;
  cursor: pointer;
  color: #666;
  line-height: 1;
  padding: 0;
  width: 32px;
  height: 32px;
}

.close-btn:hover {
  color: #000;
}

h2 {
  margin: 0 0 20px 0;
  color: #333;
}

h3 {
  margin: 20px 0 10px 0;
  color: #555;
  font-size: 18px;
}

.data-section {
  margin-bottom: 30px;
}

.current-values-chart {
  margin-bottom: 20px;
  background: #f9f9f9;
  border-radius: 8px;
  padding: 10px;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 10px;
}

.data-table th,
.data-table td {
  padding: 10px;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

.data-table th:first-child,
.data-table td:first-child {
  text-align: left;
  padding-left: 15px;
}

.data-table th {
  background: #f5f5f5;
  font-weight: 600;
  color: #333;
}

.data-table tr:hover {
  background: #f9f9f9;
}

.clickable-row {
  cursor: pointer;
  transition: background 0.2s;
}

.clickable-row:hover {
  background: #e3f2fd !important;
}

.selected-row {
  background: #bbdefb !important;
}

.charts-section {
  margin-top: 20px;
  display: flex;
  flex-direction: column;
  gap: 30px;
}

.chart-container-large {
  width: 100%;
  height: 500px;
  background: #f9f9f9;
  border-radius: 8px;
  padding: 10px;
}

.comparison-section {
  width: 100%;
}

.comparison-section h3 {
  margin: 0 0 15px 0;
  color: #555;
  font-size: 18px;
}

.comparison-checkboxes {
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
  margin-bottom: 15px;
  padding: 15px;
  background: #f5f5f5;
  border-radius: 8px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 14px;
  color: #333;
}

.param-checkbox {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.chart-container-comparison {
  width: 100%;
  height: 400px;
  background: #f9f9f9;
  border-radius: 8px;
  padding: 10px;
}
</style>
