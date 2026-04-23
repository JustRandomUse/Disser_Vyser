<template>
  <div v-if="isOpen" class="modal-backdrop" @click="closeModal">
    <div class="modal-content" @click.stop>
      <button class="close-btn" @click="closeModal">&times;</button>

      <h2>Статистика по выбранным датчикам</h2>
      <p class="sensors-count">Датчиков: {{ sensors.length }}</p>

      <div class="parameter-selector">
        <label>Параметр:</label>
        <select v-model="selectedParameter" @change="renderChart">
          <option value="pm25">PM2.5</option>
          <option value="pm10">PM10</option>
          <option value="temperature">Температура</option>
          <option value="humidity">Влажность</option>
          <option value="pressure">Давление</option>
        </select>
      </div>

      <div class="chart-section">
        <div ref="statsChart" style="width: 100%; height: 400px;"></div>
      </div>

      <div class="stats-section">
        <h3>Средние значения</h3>
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
  sensors: {
    type: Array,
    default: () => []
  },
  timeSeriesData: {
    type: Array,
    default: () => []
  },
  dateRange: {
    type: Object,
    default: null
  },
  rangeType: {
    type: String,
    default: 'instant' // 'instant' | 'hour' | 'day' | 'month' | 'year'
  }
});

const emit = defineEmits(['close']);

const statsChart = ref(null);
const chartInstance = ref(null);
const selectedParameter = ref('pm25');

const statistics = computed(() => {
  if (props.sensors.length === 0) return {};

  const params = ['pm25', 'pm10', 'temperature', 'humidity', 'pressure'];
  const stats = {};

  params.forEach(param => {
    const values = props.sensors.map(s => s[param] || 0).filter(v => v > 0);
    if (values.length > 0) {
      const avg = values.reduce((a, b) => a + b, 0) / values.length;
      const min = Math.min(...values);
      const max = Math.max(...values);

      stats[param] = {
        avg: Math.round(avg * 10) / 10,
        min: Math.round(min * 10) / 10,
        max: Math.round(max * 10) / 10
      };
    }
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

const renderChart = () => {
  if (chartInstance.value) {
    chartInstance.value.dispose();
  }

  chartInstance.value = echarts.init(statsChart.value);

  // Check if we have time series data
  if (props.timeSeriesData && props.timeSeriesData.length > 0 && props.rangeType !== 'instant') {
    renderTimeSeriesChart();
  } else {
    renderInstantChart();
  }
};

const renderTimeSeriesChart = () => {
  const params = ['pm25', 'pm10', 'temperature', 'humidity', 'pressure'];
  const selectedParam = selectedParameter.value;

  // Prepare data for each site
  const series = props.timeSeriesData.map(site => {
    const values = site.data.map(d => d[selectedParam]);
    const times = site.data.map(d => d.time);

    const avg = values.reduce((a, b) => a + b, 0) / values.length;
    const min = Math.min(...values);
    const max = Math.max(...values);

    return {
      name: site.name,
      type: 'line',
      data: values,
      smooth: true,
      markLine: {
        data: [
          { type: 'average', name: 'Среднее', label: { formatter: 'Среднее: {c}' } },
          { type: 'min', name: 'Минимум', label: { formatter: 'Мин: {c}' } },
          { type: 'max', name: 'Максимум', label: { formatter: 'Макс: {c}' } }
        ],
        lineStyle: {
          type: 'dashed'
        }
      }
    };
  });

  const times = props.timeSeriesData[0]?.data.map(d => {
    const date = new Date(d.time);
    if (props.rangeType === 'hour') {
      return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
    } else if (props.rangeType === 'day') {
      return date.toLocaleDateString('ru-RU', { day: '2-digit', month: 'short' });
    } else if (props.rangeType === 'month') {
      return date.toLocaleDateString('ru-RU', { month: 'short', year: 'numeric' });
    } else if (props.rangeType === 'year') {
      return date.getFullYear().toString();
    }
    return d.time;
  }) || [];

  const option = {
    title: {
      text: `${formatKey(selectedParam)} за период`,
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
      data: props.timeSeriesData.map(s => s.name),
      top: 40,
      type: 'scroll',
      selectedMode: false
    },
    grid: {
      left: '10%',
      right: '10%',
      bottom: '15%',
      top: '25%'
    },
    xAxis: {
      type: 'category',
      data: times,
      boundaryGap: false,
      axisLabel: {
        rotate: 45
      }
    },
    yAxis: {
      type: 'value',
      name: getUnit(selectedParam),
      axisLabel: {
        formatter: '{value}'
      }
    },
    series: series,
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

  chartInstance.value.setOption(option);
};

const renderInstantChart = () => {
  const selectedParam = selectedParameter.value;

  if (!statistics.value[selectedParam]) {
    return;
  }

  const colors = {
    pm25: '#ff6384',
    pm10: '#36a2eb',
    temperature: '#ffce56',
    humidity: '#4bc0c0',
    pressure: '#9966ff'
  };

  const stat = statistics.value[selectedParam];

  // Create data points showing current values for each sensor
  const data = props.sensors.map((sensor, index) => {
    return [index, sensor[selectedParam] || 0];
  });

  const sensorNames = props.sensors.map(s => s.name || `Датчик ${s.id}`);

  const option = {
    title: {
      text: formatKey(selectedParam) + ' (текущие значения)',
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
      data: [formatKey(selectedParam)],
      top: 40,
      selectedMode: false
    },
    grid: {
      left: '10%',
      right: '10%',
      bottom: '15%',
      top: '25%'
    },
    xAxis: {
      type: 'category',
      data: sensorNames,
      boundaryGap: false,
      axisLabel: {
        rotate: 45,
        interval: 0
      }
    },
    yAxis: {
      type: 'value',
      name: getUnit(selectedParam),
      axisLabel: {
        formatter: '{value}'
      }
    },
    series: [
      {
        name: formatKey(selectedParam),
        type: 'line',
        smooth: true,
        data: data,
        itemStyle: {
          color: colors[selectedParam]
        },
        markLine: {
          data: [
            {
              yAxis: stat.avg,
              name: 'Среднее',
              label: {
                formatter: 'Среднее: {c}',
                position: 'end'
              }
            },
            {
              yAxis: stat.min,
              name: 'Минимум',
              label: {
                formatter: 'Мин: {c}',
                position: 'start'
              }
            },
            {
              yAxis: stat.max,
              name: 'Максимум',
              label: {
                formatter: 'Макс: {c}',
                position: 'start'
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

  chartInstance.value.setOption(option);
};

watch(() => props.isOpen, (newVal) => {
  if (newVal) {
    nextTick(() => {
      renderChart();
    });
  }
});

onBeforeUnmount(() => {
  if (chartInstance.value) {
    chartInstance.value.dispose();
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
  z-index: 3000;
}

.modal-content {
  background: white;
  border-radius: 12px;
  padding: 30px;
  max-width: 1000px;
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
  margin: 0 0 10px 0;
  color: #333;
}

.sensors-count {
  color: #666;
  font-size: 14px;
  margin-bottom: 20px;
}

.stats-section {
  margin-bottom: 30px;
}

h3 {
  margin: 20px 0 10px 0;
  color: #555;
  font-size: 18px;
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

.chart-section {
  margin-top: 20px;
  background: #f9f9f9;
  border-radius: 8px;
  padding: 10px;
}

.parameter-selector {
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.parameter-selector label {
  font-weight: 600;
  color: #333;
}

.parameter-selector select {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 14px;
  background: white;
  cursor: pointer;
}

.parameter-selector select:hover {
  border-color: #3b82f6;
}

.parameter-selector select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}
</style>
