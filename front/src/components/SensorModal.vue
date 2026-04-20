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

<script>
import * as echarts from 'echarts';

export default {
  name: 'SensorModal',
  props: {
    isOpen: {
      type: Boolean,
      default: false
    },
    sensorData: {
      type: Object,
      default: () => ({})
    }
  },
  data() {
    return {
      timeSeriesChart: null,
      comparisonChart: null,
      currentValuesChart: null,
      selectedSingleParam: 'pm25',
      comparisonParams: []
    };
  },
  computed: {
    measurements() {
      const data = this.sensorData;
      return {
        pm25: data.pm25 || 0,
        pm10: data.pm10 || 0,
        temperature: data.temperature || 0,
        humidity: data.humidity || 0,
        pressure: data.pressure || 0
      };
    }
  },
  watch: {
    isOpen(newVal) {
      if (newVal) {
        this.$nextTick(() => {
          this.renderCharts();
        });
      }
    },
    selectedSingleParam() {
      if (this.isOpen) {
        this.$nextTick(() => {
          this.renderTimeSeriesChart();
        });
      }
    },
    comparisonParams: {
      handler() {
        if (this.isOpen) {
          this.$nextTick(() => {
            this.renderComparisonChart();
          });
        }
      },
      deep: true
    }
  },
  methods: {
    closeModal() {
      this.$emit('close');
    },
    formatKey(key) {
      const labels = {
        pm25: 'PM2.5',
        pm10: 'PM10',
        temperature: 'Температура',
        humidity: 'Влажность',
        pressure: 'Давление'
      };
      return labels[key] || key;
    },
    getUnit(key) {
      const units = {
        pm25: 'мкг/м³',
        pm10: 'мкг/м³',
        temperature: '°C',
        humidity: '%',
        pressure: 'гПа'
      };
      return units[key] || '';
    },
    selectSingleParam(key) {
      this.selectedSingleParam = key;
    },
    toggleComparisonParam(key) {
      const index = this.comparisonParams.indexOf(key);
      if (index > -1) {
        this.comparisonParams.splice(index, 1);
      } else {
        this.comparisonParams.push(key);
      }
    },
    renderCharts() {
      this.renderCurrentValuesChart();
      this.renderTimeSeriesChart();
      this.renderComparisonChart();
    },
    renderCurrentValuesChart() {
      if (this.currentValuesChart) {
        this.currentValuesChart.dispose();
      }

      this.currentValuesChart = echarts.init(this.$refs.currentValuesChart);

      const params = Object.keys(this.measurements);
      // Simulate range data: [min, max, min, current]
      // For demo, we'll use current value +/- 20%
      const candlestickData = params.map(key => {
        const current = this.measurements[key];
        const min = Math.max(0, current * 0.8);
        const max = current * 1.2;
        return [min, max, min, current];
      });

      const option = {
        title: {
          text: 'Текущие показатели датчика',
          left: 'center',
          top: 10
        },
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'cross'
          },
          formatter: function(params) {
            const param = params[0];
            const data = param.data;
            return `${param.name}<br/>
              Текущее: ${data[3]}<br/>
              Диапазон: ${data[2].toFixed(1)} - ${data[1].toFixed(1)}`;
          }
        },
        brush: {
          toolbox: ['rect', 'polygon', 'lineX', 'lineY', 'keep', 'clear'],
          xAxisIndex: 0
        },
        toolbox: {
          feature: {
            brush: {
              type: ['rect', 'polygon', 'lineX', 'lineY', 'keep', 'clear']
            },
            dataZoom: {
              yAxisIndex: false
            },
            restore: {},
            saveAsImage: {}
          }
        },
        grid: {
          left: '10%',
          right: '10%',
          bottom: '15%',
          top: '25%'
        },
        xAxis: {
          type: 'category',
          data: params.map(p => this.formatKey(p)),
          scale: true,
          boundaryGap: true,
          axisLine: { onZero: false },
          splitLine: { show: false }
        },
        yAxis: {
          scale: true,
          splitArea: {
            show: true
          }
        },
        dataZoom: [
          {
            type: 'inside',
            start: 0,
            end: 100
          },
          {
            show: true,
            type: 'slider',
            top: '90%',
            start: 0,
            end: 100
          }
        ],
        series: [
          {
            name: 'Значения',
            type: 'candlestick',
            data: candlestickData,
            itemStyle: {
              color: '#3b82f6',
              color0: '#10b981',
              borderColor: '#2563eb',
              borderColor0: '#059669'
            }
          }
        ]
      };

      this.currentValuesChart.setOption(option);
    },
    generateTimeSeriesData() {
      const data = [];
      const now = new Date();

      // Generate 7 days of hourly data (168 hours)
      for (let i = 0; i < 168; i++) {
        const date = new Date(now.getTime() - (168 - i) * 3600000);
        const hour = date.getHours();

        // Simulate daily patterns with rush hour peaks
        const basePM25 = 30 + Math.sin(i / 12) * 20 + Math.random() * 15;
        const pm25 = Math.max(0, basePM25 + (hour >= 7 && hour <= 9 ? 20 : 0) + (hour >= 17 && hour <= 19 ? 25 : 0));

        const basePM10 = pm25 * 1.5 + Math.random() * 10;
        const pm10 = Math.max(0, basePM10);

        const baseTemp = 15 + Math.sin(i / 24 * Math.PI * 2) * 8 + Math.random() * 3;
        const temperature = Math.round(baseTemp * 10) / 10;

        const baseHumidity = 60 + Math.sin(i / 24 * Math.PI * 2) * 15 + Math.random() * 10;
        const humidity = Math.max(30, Math.min(90, Math.round(baseHumidity)));

        const basePressure = 1013 + Math.sin(i / 48 * Math.PI * 2) * 5 + Math.random() * 3;
        const pressure = Math.round(basePressure * 10) / 10;

        data.push({
          date: date.toISOString(),
          pm25: Math.round(pm25 * 10) / 10,
          pm10: Math.round(pm10 * 10) / 10,
          temperature: temperature,
          humidity: humidity,
          pressure: pressure
        });
      }

      return data;
    },
    renderTimeSeriesChart() {
      if (this.timeSeriesChart) {
        this.timeSeriesChart.dispose();
      }

      this.timeSeriesChart = echarts.init(this.$refs.timeSeriesChart);
      const data = this.generateTimeSeriesData();

      const param = this.selectedSingleParam;
      const colors = {
        pm25: '#ff6384',
        pm10: '#36a2eb',
        temperature: '#ffce56',
        humidity: '#4bc0c0',
        pressure: '#9966ff'
      };

      const units = {
        pm25: 'мкг/м³',
        pm10: 'мкг/м³',
        temperature: '°C',
        humidity: '%',
        pressure: 'гПа'
      };

      const option = {
        title: {
          text: `${this.formatKey(param)} за 7 дней`,
          left: 'center',
          top: 10
        },
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'cross'
          }
        },
        legend: {
          data: [this.formatKey(param)],
          top: 40
        },
        grid: {
          left: '3%',
          right: '3%',
          bottom: '10%',
          top: '20%',
          containLabel: true
        },
        xAxis: {
          type: 'time',
          boundaryGap: false
        },
        yAxis: {
          type: 'value',
          name: `${this.formatKey(param)} (${units[param]})`,
          axisLine: {
            lineStyle: {
              color: colors[param]
            }
          }
        },
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
        series: [
          {
            name: this.formatKey(param),
            type: 'line',
            smooth: true,
            data: data.map(d => [d.date, d[param]]),
            itemStyle: {
              color: colors[param]
            },
            areaStyle: {
              color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                { offset: 0, color: colors[param] + '4D' },
                { offset: 1, color: colors[param] + '0D' }
              ])
            }
          }
        ]
      };

      this.timeSeriesChart.setOption(option);
    },
    renderComparisonChart() {
      if (this.comparisonChart) {
        this.comparisonChart.dispose();
      }

      if (this.comparisonParams.length === 0) {
        return;
      }

      this.comparisonChart = echarts.init(this.$refs.comparisonChart);
      const data = this.generateTimeSeriesData();

      const colors = {
        pm25: '#ff6384',
        pm10: '#36a2eb',
        temperature: '#ffce56',
        humidity: '#4bc0c0',
        pressure: '#9966ff'
      };

      const units = {
        pm25: 'мкг/м³',
        pm10: 'мкг/м³',
        temperature: '°C',
        humidity: '%',
        pressure: 'гПа'
      };

      const yAxisConfig = this.comparisonParams.map((param, index) => ({
        type: 'value',
        name: `${this.formatKey(param)} (${units[param]})`,
        position: index % 2 === 0 ? 'left' : 'right',
        offset: Math.floor(index / 2) * 60,
        axisLine: {
          lineStyle: {
            color: colors[param]
          }
        },
        axisLabel: {
          color: colors[param]
        }
      }));

      const series = this.comparisonParams.map((param, index) => ({
        name: this.formatKey(param),
        type: 'line',
        smooth: true,
        yAxisIndex: index,
        data: data.map(d => [d.date, d[param]]),
        itemStyle: {
          color: colors[param]
        },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: colors[param] + '4D' },
            { offset: 1, color: colors[param] + '0D' }
          ])
        }
      }));

      const option = {
        title: {
          text: 'Сравнение параметров за 7 дней',
          left: 'center',
          top: 10
        },
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'cross'
          }
        },
        legend: {
          data: this.comparisonParams.map(p => this.formatKey(p)),
          top: 40
        },
        grid: {
          left: '3%',
          right: '3%',
          bottom: '10%',
          top: '20%',
          containLabel: true
        },
        xAxis: {
          type: 'time',
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
