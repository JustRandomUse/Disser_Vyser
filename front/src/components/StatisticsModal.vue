<template>
  <div v-if="isOpen" class="modal-backdrop" @click="closeModal">
    <div class="modal-content" @click.stop>
      <button class="close-btn" @click="closeModal">&times;</button>

      <h2>Статистика по выбранным датчикам</h2>
      <p class="sensors-count">Датчиков: {{ sensors.length }}</p>

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

      <div class="chart-section">
        <div ref="statsChart" style="width: 100%; height: 400px;"></div>
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
  }
});

const emit = defineEmits(['close']);

const statsChart = ref(null);
const chartInstance = ref(null);

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

  const params = Object.keys(statistics.value);
  const candlestickData = params.map(p => {
    const stat = statistics.value[p];
    return [stat.min, stat.max, stat.min, stat.avg];
  });

  const option = {
    title: {
      text: 'Диапазон значений параметров',
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
          Минимум: ${data[2]}<br/>
          Среднее: ${data[3]}<br/>
          Максимум: ${data[1]}`;
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
      top: '20%'
    },
    xAxis: {
      type: 'category',
      data: params.map(p => formatKey(p)),
      scale: true,
      boundaryGap: true,
      axisLine: { onZero: false },
      splitLine: { show: false },
      min: 'dataMin',
      max: 'dataMax'
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
        name: 'Диапазон',
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
</style>
