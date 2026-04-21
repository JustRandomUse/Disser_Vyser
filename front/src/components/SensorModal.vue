<template>
  <div v-if="isOpen" class="modal-backdrop" @click="closeModal">
    <div class="modal-content" @click.stop>
      <button class="close-btn" @click="closeModal">&times;</button>

      <h2>{{ sensorData.name }}</h2>

      <div class="current-values">
        <h3>Текущие показатели</h3>
        <div ref="currentValuesChart" style="width: 100%; height: 300px;"></div>
      </div>

      <div class="time-series-section">
        <h3>Динамика за 7 дней</h3>
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

  for (let i = 0; i < 168; i++) {
    const date = new Date(now.getTime() - (168 - i) * 3600000);
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

  const params = Object.keys(measurements.value);
  const candlestickData = params.map(key => {
    const current = measurements.value[key];
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
        return param.name + '<br/>' +
          'Текущее: ' + data[3] + '<br/>' +
          'Диапазон: ' + data[2].toFixed(1) + ' - ' + data[1].toFixed(1);
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
      data: params.map(p => formatKey(p)),
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
      text: formatKey(param) + ' за 7 дней',
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
      name: formatKey(param) + ' (' + units[param] + ')',
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
      data: comparisonParams.value.map(p => formatKey(p)),
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
</style>
