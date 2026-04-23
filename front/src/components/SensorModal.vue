<template>
  <div v-if="isOpen" class="modal-backdrop" @click="closeModal">
    <div class="modal-content" @click.stop>
      <button class="close-btn" @click="closeModal">&times;</button>

      <h2>{{ sensorData.name }}</h2>

      <div class="current-values">
        <h3>Средние значения</h3>
        <div class="param-selector">
          <button
            v-for="(key, index) in Object.keys(measurements)"
            :key="index"
            :class="{ active: selectedCurrentParam === key }"
            @click="selectCurrentParam(key)"
          >
            {{ formatKey(key) }}
          </button>
        </div>
        <div ref="currentValuesChart" style="width: 100%; height: 300px;"></div>

        <div class="stats-section">
          <h4>Средние значения</h4>
          <table class="stats-table">
            <thead>
              <tr>
                <th>Параметр</th>
                <th>Среднее</th>
                <th>Минимум</th>
                <th>Максимум</th>
                <th>Единица</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(stat, key) in statistics" :key="key">
                <td>{{ formatKey(key) }}</td>
                <td>{{ stat.avg }}</td>
                <td>{{ stat.min }}</td>
                <td>{{ stat.max }}</td>
                <td>{{ getUnit(key) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="time-series-section">
        <h3>Динамика значений</h3>
        <div class="param-selector">
          <button
            v-for="(key, index) in Object.keys(measurements)"
            :key="index"
            :class="{ active: selectedSingleParam === key }"
            @click="selectSingleParam(key)"
          >
            {{ formatKey(key) }}
          </button>
        </div>
        <div ref="timeSeriesChart" style="width: 100%; height: 400px;"></div>
      </div>

      <div class="comparison-section">
        <h3>Сравнение параметров</h3>
        <div class="param-selector">
          <button
            v-for="(key, index) in Object.keys(measurements)"
            :key="index"
            :class="{ active: comparisonParams.includes(key) }"
            @click="toggleComparisonParam(key)"
          >
            {{ formatKey(key) }}
          </button>
        </div>
        <div v-if="comparisonParams.length > 0" ref="comparisonChart" style="width: 100%; height: 400px;"></div>
        <p v-else class="no-selection">Выберите параметры для сравнения</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onBeforeUnmount } from 'vue';
import * as echarts from 'echarts';

const props = defineProps({
  isOpen: {
    type: Boolean,
    default: false
  },
  sensorData: {
    type: Object,
    default: () => ({})
  }
});

const emit = defineEmits(['close']);

const timeSeriesChart = ref(null);
const comparisonChart = ref(null);
const currentValuesChart = ref(null);
const timeSeriesChartInstance = ref(null);
const comparisonChartInstance = ref(null);
const currentValuesChartInstance = ref(null);
const selectedSingleParam = ref('pm25');
const selectedCurrentParam = ref('pm25');
const comparisonParams = ref([]);

const measurements = computed(() => {
  const data = props.sensorData;
  return {
    pm25: data.pm25 || 0,
    pm10: data.pm10 || 0,
    temperature: data.temperature || 0,
    humidity: data.humidity || 0,
    pressure: data.pressure || 0
  };
});

const statistics = computed(() => {
  const data = generateTimeSeriesData();
  const stats = {};

  Object.keys(measurements.value).forEach(param => {
    const values = data.map(d => d[param]);
    const avg = values.reduce((a, b) => a + b, 0) / values.length;
    const min = Math.min(...values);
    const max = Math.max(...values);

    stats[param] = {
      avg: Math.round(avg * 10) / 10,
      min: Math.round(min * 10) / 10,
      max: Math.round(max * 10) / 10
    };
  });

  return stats;
});

const closeModal = () => {
  emit('close');
};

const formatKey = (key) => {
  const labels = {
    pm25: 'PM2.5',
    pm10: 'PM10',
    temperature: 'Температура',
    humidity: 'Влажность',
    pressure: 'Давление'
  };
  return labels[key] || key;
};

const getUnit = (key) => {
  const units = {
    pm25: 'мкг/м³',
    pm10: 'мкг/м³',
    temperature: '°C',
    humidity: '%',
    pressure: 'гПа'
  };
  return units[key] || '';
};

const selectSingleParam = (key) => {
  selectedSingleParam.value = key;
};

const selectCurrentParam = (key) => {
  selectedCurrentParam.value = key;
};

const toggleComparisonParam = (key) => {
  const index = comparisonParams.value.indexOf(key);
  if (index > -1) {
    comparisonParams.value.splice(index, 1);
  } else {
    comparisonParams.value.push(key);
  }
};

const generateTimeSeriesData = () => {
  const data = [];
  const now = new Date();

  for (let i = 0; i < 24; i++) {
    const date = new Date(now.getTime() - (24 - i) * 3600000);
    const hour = date.getHours();

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
};

const renderCurrentValuesChart = () => {
  if (currentValuesChartInstance.value) {
    currentValuesChartInstance.value.dispose();
  }

  currentValuesChartInstance.value = echarts.init(currentValuesChart.value);

  const data = generateTimeSeriesData();
  const param = selectedCurrentParam.value;

  const colors = {
    pm25: '#ff6384',
    pm10: '#36a2eb',
    temperature: '#ffce56',
    humidity: '#4bc0c0',
    pressure: '#9966ff'
  };

  const values = data.map(d => d[param]);
  const avg = values.reduce((a, b) => a + b, 0) / values.length;

  const option = {
    title: {
      text: formatKey(param) + ' ',
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
      data: [formatKey(param)],
      top: 40,
      selectedMode: false
    },
    grid: {
      left: '15%',
      right: '15%',
      bottom: '15%',
      top: '25%'
    },
    xAxis: {
      type: 'time',
      boundaryGap: false
    },
    yAxis: {
      type: 'value',
      name: getUnit(param),
      axisLabel: {
        formatter: '{value}'
      }
    },
    series: [
      {
        name: formatKey(param),
        type: 'line',
        smooth: true,
        data: data.map(d => [d.date, d[param]]),
        itemStyle: {
          color: colors[param]
        },
        markLine: {
          data: [
            {
              type: 'average',
              name: 'Среднее',
              label: {
                formatter: 'Среднее: {c}',
                position: 'end'
              }
            },
            {
              type: 'min',
              name: 'Минимум',
              label: {
                formatter: 'Мин: {c}',
                position: 'end'
              }
            },
            {
              type: 'max',
              name: 'Максимум',
              label: {
                formatter: 'Макс: {c}',
                position: 'end'
              }
            }
          ],
          lineStyle: {
            type: 'dashed'
          }
        }
      }
    ],
    toolbox: {
      feature: {
        dataZoom: {
          yAxisIndex: 'none'
        },
        restore: {},
        saveAsImage: {}
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
    ]
  };

  currentValuesChartInstance.value.setOption(option);
};

const renderTimeSeriesChart = () => {
  if (timeSeriesChartInstance.value) {
    timeSeriesChartInstance.value.dispose();
  }

  timeSeriesChartInstance.value = echarts.init(timeSeriesChart.value);
  const data = generateTimeSeriesData();

  const param = selectedSingleParam.value;
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
      text: formatKey(param) + ' ',
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
      data: [formatKey(param)],
      top: 40,
      selectedMode: false
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
      name: formatKey(param) + ' (' + units[param] + ')',
      axisLine: {
        lineStyle: {
          color: colors[param]
        }
      }
    },
    toolbox: {
      feature: {
        dataZoom: {
          yAxisIndex: 'none'
        },
        restore: {},
        saveAsImage: {}
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
        name: formatKey(param),
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

  timeSeriesChartInstance.value.setOption(option);
};

const renderComparisonChart = () => {
  if (comparisonChartInstance.value) {
    comparisonChartInstance.value.dispose();
  }

  if (comparisonParams.value.length === 0) {
    return;
  }

  comparisonChartInstance.value = echarts.init(comparisonChart.value);
  const data = generateTimeSeriesData();

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

  const yAxisConfig = comparisonParams.value.map((param, index) => ({
    type: 'value',
    name: formatKey(param) + ' (' + units[param] + ')',
    position: index % 2 === 0 ? 'left' : 'right',
    offset: Math.floor(index / 2) * 80,
    axisLine: {
      lineStyle: {
        color: colors[param]
      }
    },
    axisLabel: {
      color: colors[param]
    }
  }));

  const series = comparisonParams.value.map((param, index) => ({
    name: formatKey(param),
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

  // Calculate dynamic grid margins based on number of axes
  const leftAxesCount = Math.ceil(comparisonParams.value.length / 2);
  const rightAxesCount = Math.floor(comparisonParams.value.length / 2);
  const gridLeft = leftAxesCount > 1 ? `${10 + (leftAxesCount - 1) * 8}%` : '10%';
  const gridRight = rightAxesCount > 0 ? `${10 + rightAxesCount * 8}%` : '10%';

  const option = {
    title: {
      text: 'Сравнение параметров',
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
      data: comparisonParams.value.map(p => formatKey(p)),
      top: 40,
      selectedMode: false
    },
    grid: {
      left: gridLeft,
      right: gridRight,
      bottom: '10%',
      top: '20%',
      containLabel: false
    },
    xAxis: {
      type: 'time',
      boundaryGap: false
    },
    yAxis: yAxisConfig,
    toolbox: {
      feature: {
        dataZoom: {
          yAxisIndex: 'none'
        },
        restore: {},
        saveAsImage: {}
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
    series: series
  };

  comparisonChartInstance.value.setOption(option);
};

const renderCharts = () => {
  renderCurrentValuesChart();
  renderTimeSeriesChart();
  renderComparisonChart();
};

watch(() => props.isOpen, (newVal) => {
  if (newVal) {
    nextTick(() => {
      renderCharts();
    });
  }
});

watch(selectedSingleParam, () => {
  if (props.isOpen) {
    nextTick(() => {
      renderTimeSeriesChart();
    });
  }
});

watch(selectedCurrentParam, () => {
  if (props.isOpen) {
    nextTick(() => {
      renderCurrentValuesChart();
    });
  }
});

watch(comparisonParams, () => {
  if (props.isOpen) {
    nextTick(() => {
      renderComparisonChart();
    });
  }
}, { deep: true });

onBeforeUnmount(() => {
  if (timeSeriesChartInstance.value) {
    timeSeriesChartInstance.value.dispose();
  }
  if (comparisonChartInstance.value) {
    comparisonChartInstance.value.dispose();
  }
  if (currentValuesChartInstance.value) {
    currentValuesChartInstance.value.dispose();
  }
});
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
  margin: 30px 0 15px 0;
  color: #555;
  font-size: 18px;
}

.current-values,
.time-series-section,
.comparison-section {
  margin-bottom: 30px;
}

.param-selector {
  display: flex;
  gap: 10px;
  margin-bottom: 15px;
  flex-wrap: wrap;
}

.param-selector button {
  padding: 8px 16px;
  border: 2px solid #ddd;
  border-radius: 6px;
  background: white;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.param-selector button:hover {
  border-color: #3b82f6;
  background: #eff6ff;
}

.param-selector button.active {
  background: #3b82f6;
  color: white;
  border-color: #3b82f6;
}

.no-selection {
  text-align: center;
  color: #999;
  padding: 40px;
  font-style: italic;
}

.stats-section {
  margin-top: 20px;
}

.stats-section h4 {
  margin: 10px 0;
  color: #555;
  font-size: 16px;
}

.stats-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 10px;
}

.stats-table th,
.stats-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

.stats-table th {
  background: #f5f5f5;
  font-weight: 600;
  color: #333;
}

.stats-table tr:hover {
  background: #f9f9f9;
}
</style>
