<template>
  <div v-if="isOpen" class="modal-backdrop" @click="closeModal">
    <div class="modal-content" @click.stop>
      <button class="close-btn" @click="closeModal">&times;</button>

      <h2>{{ sensorData.name }}</h2>

      <div class="param-selector">
        <button
          v-for="(key, index) in Object.keys(measurements)"
          :key="index"
          :class="{ active: selectedParams.includes(key) }"
          @click="toggleParam(key)"
        >
          {{ formatKey(key) }}
        </button>
      </div>

      <div class="chart-container">
        <div v-if="chartData.length > 0 && selectedParams.length > 0" ref="mainChart" class="chart"></div>
        <p v-if="selectedParams.length === 0" class="no-selection">
          Выберите параметры для отображения
        </p>
        <p v-if="selectedParams.length > 0 && chartData.length === 0" class="no-data">
          Нет данных за выбранный период
        </p>
        <p v-if="dateRangeText" class="date-range-info">{{ dateRangeText }}</p>
      </div>

      <div class="stats-section">
        <h4>Средние значения</h4>
        <table class="stats-table">
          <thead>
            <tr>
              <th>Показатель</th>
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
  sensorData: {
    type: Object,
    default: () => ({})
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
    default: 'instant'
  }
});

const emit = defineEmits(['close']);

const mainChart = ref(null);
const mainChartInstance = ref(null);
const selectedParams = ref(['pm25']); // Array of selected parameters

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

// Get chart data from timeSeriesData prop (real API data)
const chartData = computed(() => {
  if (props.timeSeriesData && props.timeSeriesData.length > 0) {
    const sensorData = props.timeSeriesData[0];
    if (sensorData.data && sensorData.data.length > 0) {
      return sensorData.data.map(point => ({
        date: point.time,
        pm25: point.pm25 || 0,
        pm10: point.pm10 || 0,
        temperature: point.temperature || 0,
        humidity: point.humidity || 0,
        pressure: point.pressure || 0
      }));
    }
  }
  return [];
});

const statistics = computed(() => {
  const data = chartData.value;

  // If no data available, return empty stats
  if (data.length === 0) {
    return {};
  }

  const stats = {};

  Object.keys(measurements.value).forEach(param => {
    const values = data.map(d => d[param]).filter(v => v > 0);

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

const dateRangeText = computed(() => {
  if (props.rangeType === 'instant') {
    return null;
  }

  if (!props.dateRange || !props.dateRange.start || !props.dateRange.end) {
    return null;
  }

  const start = new Date(props.dateRange.start);
  const end = new Date(props.dateRange.end);

  if (props.rangeType === 'hour') {
    return `Период: ${start.toLocaleDateString('ru-RU')} ${start.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })} - ${end.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}`;
  } else if (props.rangeType === 'day') {
    return `Период: ${start.toLocaleDateString('ru-RU', { day: '2-digit', month: 'short', year: 'numeric' })} - ${end.toLocaleDateString('ru-RU', { day: '2-digit', month: 'short', year: 'numeric' })}`;
  } else if (props.rangeType === 'month') {
    return `Период: ${start.toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' })} - ${end.toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' })}`;
  } else if (props.rangeType === 'year') {
    return `Период: ${start.getFullYear()} - ${end.getFullYear()}`;
  }

  return null;
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

const toggleParam = (key) => {
  const index = selectedParams.value.indexOf(key);
  if (index > -1) {
    selectedParams.value.splice(index, 1);
  } else {
    selectedParams.value.push(key);
  }
};

const renderChart = () => {
  if (selectedParams.value.length === 0) {
    if (mainChartInstance.value) {
      mainChartInstance.value.dispose();
      mainChartInstance.value = null;
    }
    return;
  }

  if (selectedParams.value.length === 1) {
    // Single parameter - show with markPoint and markLine
    renderSingleParamChart();
  } else {
    // Multiple parameters - show comparison
    renderComparisonChart();
  }
};

const renderSingleParamChart = () => {
  if (mainChartInstance.value) {
    mainChartInstance.value.dispose();
  }

  const data = chartData.value;

  // If no data, show empty state
  if (data.length === 0) {
    mainChartInstance.value = null;
    return;
  }

  mainChartInstance.value = echarts.init(mainChart.value);

  const param = selectedParams.value[0]; // First selected parameter

  const colors = {
    pm25: '#ff6384',
    pm10: '#36a2eb',
    temperature: '#ffce56',
    humidity: '#4bc0c0',
    pressure: '#9966ff'
  };

  const values = data.map(d => d[param]).filter(v => v > 0);

  if (values.length === 0) {
    mainChartInstance.value = null;
    return;
  }

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
      selectedMode: false,
      orient: 'vertical',
      right: 10,
      top: 80,
      textStyle: {
        fontSize: 11
      }
    },
    grid: {
      left: 60,
      right: 200,
      bottom: 60,
      top: 80
    },
    xAxis: {
      type: 'time',
      boundaryGap: false,
      axisLabel: {
        fontSize: 11
      }
    },
    yAxis: {
      type: 'value',
      name: getUnit(param),
      nameLocation: 'middle',
      nameGap: 50,
      nameTextStyle: {
        color: colors[param],
        fontWeight: 'bold',
        fontSize: 12
      },
      axisLine: {
        show: true,
        lineStyle: {
          color: colors[param],
          width: 2
        }
      },
      axisLabel: {
        formatter: '{value}',
        color: colors[param],
        fontSize: 11
      },
      splitLine: {
        show: true,
        lineStyle: {
          color: '#e0e0e0',
          type: 'dashed'
        }
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
            }
          ],
          lineStyle: {
            type: 'dashed'
          }
        },
        markPoint: {
          data: [
            {
              type: 'max',
              name: 'Максимум',
              label: {
                formatter: 'Макс: {c}',
                position: 'top'
              },
              itemStyle: {
                color: '#e74c3c'
              }
            },
            {
              type: 'min',
              name: 'Минимум',
              label: {
                formatter: 'Мин: {c}',
                position: 'bottom'
              },
              itemStyle: {
                color: '#3498db'
              }
            }
          ],
          symbolSize: 60,
          label: {
            fontSize: 12,
            fontWeight: 'bold'
          }
        }
      }
    ],
    toolbox: {
      feature: {
        restore: {},
        saveAsImage: {}
      },
      right: 20,
      top: 10
    },
    dataZoom: [
      {
        type: 'inside',
        start: 0,
        end: 100,
        zoomOnMouseWheel: true,
        moveOnMouseMove: true
      }
    ]
  };

  mainChartInstance.value.setOption(option);
};

const renderComparisonChart = () => {
  if (mainChartInstance.value) {
    mainChartInstance.value.dispose();
  }

  const data = chartData.value;

  // If no data, show empty state
  if (data.length === 0) {
    mainChartInstance.value = null;
    return;
  }

  mainChartInstance.value = echarts.init(mainChart.value);

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

  // Нормализация данных для каждого параметра (0-100%)
  const normalizedData = {};
  const ranges = {};

  selectedParams.value.forEach(param => {
    const values = data.map(d => d[param]);
    const min = Math.min(...values);
    const max = Math.max(...values);
    ranges[param] = { min, max, range: max - min };

    normalizedData[param] = data.map(d => {
      const normalized = ranges[param].range > 0
        ? ((d[param] - min) / ranges[param].range) * 100
        : 50;
      return [d.date, normalized];
    });
  });

  // Создаем серии с нормализованными данными
  const series = selectedParams.value.map((param) => ({
    name: formatKey(param),
    type: 'line',
    smooth: true,
    yAxisIndex: 0,
    data: normalizedData[param],
    itemStyle: {
      color: colors[param]
    },
    lineStyle: {
      width: 2
    }
  }));

  // Формируем легенду с текущими средними значениями
  const legendData = selectedParams.value.map(param => {
    const stat = statistics.value[param];
    return `${formatKey(param)}: ${stat.avg} ${units[param]}`;
  });

  const gridLeft = 60;
  const gridRight = 200;

  const option = {
    title: {
      text: 'Сравнение параметров (нормализовано)',
      left: 'center',
      top: 10
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross'
      },
      formatter: function(params) {
        let result = params[0].axisValueLabel + '<br/>';
        params.forEach(item => {
          const param = selectedParams.value[item.seriesIndex];
          const normalizedValue = item.value[1];
          const originalValue = ranges[param].min + (normalizedValue / 100) * ranges[param].range;
          result += `${item.marker} ${item.seriesName}: ${originalValue.toFixed(1)} ${units[param]}<br/>`;
        });
        return result;
      }
    },
    legend: {
      data: legendData,
      top: 40,
      selectedMode: false,
      orient: 'vertical',
      right: 10,
      top: 80,
      textStyle: {
        fontSize: 11
      }
    },
    grid: {
      left: gridLeft,
      right: gridRight,
      bottom: 60,
      top: 80,
      containLabel: false
    },
    xAxis: {
      type: 'time',
      boundaryGap: false,
      axisLabel: {
        fontSize: 11
      }
    },
    yAxis: {
      type: 'value',
      name: 'Нормализовано (%)',
      nameLocation: 'middle',
      nameGap: 50,
      min: 0,
      max: 100,
      axisLabel: {
        formatter: '{value}%',
        fontSize: 11
      },
      splitLine: {
        show: true,
        lineStyle: {
          color: '#e0e0e0',
          type: 'dashed'
        }
      }
    },
    toolbox: {
      feature: {
        restore: {},
        saveAsImage: {}
      },
      right: 20,
      top: 10
    },
    dataZoom: [
      {
        type: 'inside',
        start: 0,
        end: 100,
        zoomOnMouseWheel: true,
        moveOnMouseMove: true
      }
    ],
    series: series
  };

  mainChartInstance.value.setOption(option);
};

const renderCharts = () => {
  renderChart();

  // Resize chart after render to ensure correct dimensions
  nextTick(() => {
    if (mainChartInstance.value) {
      mainChartInstance.value.resize();
    }
  });
};

watch(() => props.isOpen, (newVal) => {
  if (newVal) {
    nextTick(() => {
      // Small delay to ensure modal container has final size
      setTimeout(() => {
        renderCharts();
      }, 50);
    });
  }
});

// Watch for timeSeriesData changes
watch(() => props.timeSeriesData, () => {
  if (props.isOpen) {
    nextTick(() => {
      renderChart();
    });
  }
}, { deep: true });

watch(selectedParams, () => {
  if (props.isOpen) {
    nextTick(() => {
      renderChart();
    });
  }
}, { deep: true });

onBeforeUnmount(() => {
  if (mainChartInstance.value) {
    mainChartInstance.value.dispose();
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
  border-radius: 0;
  padding: 20px;
  width: 100vw;
  height: 100vh;
  overflow-y: auto;
  position: relative;
  box-shadow: none;
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

.chart-container {
  margin-bottom: 20px;
}

.chart {
  width: 100%;
  height: calc(100vh - 300px);
  min-height: 500px;
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

.no-data {
  text-align: center;
  color: #e74c3c;
  padding: 40px;
  font-style: italic;
  font-weight: 500;
}

.date-range-info {
  text-align: center;
  color: #666;
  font-size: 14px;
  margin-top: 10px;
  padding: 8px;
  background: #f5f5f5;
  border-radius: 4px;
  font-weight: 500;
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
