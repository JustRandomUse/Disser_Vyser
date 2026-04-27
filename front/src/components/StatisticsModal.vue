<template>
  <div v-if="isOpen" class="modal-backdrop" @click="closeModal">
    <div class="modal-content" @click.stop>
      <button class="close-btn" @click="closeModal">&times;</button>

      <h2>Среднее значение по выбранным районам</h2>

      <div class="param-selector">
        <button
          v-for="param in availableParams"
          :key="param"
          :class="{ active: selectedParams.includes(param) }"
          @click="toggleParam(param)"
        >
          {{ formatKey(param) }}
        </button>
      </div>

      <div class="chart-container">
        <div v-if="selectedParams.length > 0" ref="statsChart" class="chart"></div>
        <p v-else class="no-selection">Выберите параметры для отображения</p>
        <p v-if="dateRangeText" class="date-range-info">{{ dateRangeText }}</p>
      </div>

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
              <td>{{ formatDisplayValue(stat.avg) }}</td>
              <td>{{ formatDisplayValue(stat.min) }}</td>
              <td>{{ formatDisplayValue(stat.max) }}</td>
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
import { isValidMetricValue, formatDisplayValue } from '../utils/sensorDataRules';

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
  },
  showIndividual: {
    type: Boolean,
    default: false
  }
});

const emit = defineEmits(['close']);

const statsChart = ref(null);
const chartInstance = ref(null);
const selectedParams = ref(['pm25']);
const availableParams = ['pm25', 'pm10', 'temperature', 'humidity', 'pressure'];
let resizeTimer = null;

const toggleParam = (param) => {
  const index = selectedParams.value.indexOf(param);
  if (index > -1) {
    selectedParams.value.splice(index, 1);
  } else {
    selectedParams.value.push(param);
  }
};

const statistics = computed(() => {
  // If we have time series data for a range, calculate statistics from it
  if (props.timeSeriesData && props.timeSeriesData.length > 0 && props.rangeType !== 'instant') {
    const params = ['pm25', 'pm10', 'temperature', 'humidity', 'pressure'];
    const stats = {};

    params.forEach(param => {
      const allValues = [];

      // Collect all values for this parameter across all sites and time points
      props.timeSeriesData.forEach(site => {
        if (site.data && Array.isArray(site.data)) {
          site.data.forEach(point => {
            if (isValidMetricValue(point[param])) {
              allValues.push(point[param]);
            }
          });
        }
      });

      if (allValues.length > 0) {
        const avg = allValues.reduce((a, b) => a + b, 0) / allValues.length;
        const min = Math.min(...allValues);
        const max = Math.max(...allValues);

        stats[param] = {
          avg: Math.round(avg * 10) / 10,
          min: Math.round(min * 10) / 10,
          max: Math.round(max * 10) / 10
        };
      }
    });

    return stats;
  }

  // Fallback to instant mode: calculate from current sensor values
  if (props.sensors.length === 0) return {};

  const params = ['pm25', 'pm10', 'temperature', 'humidity', 'pressure'];
  const stats = {};

  params.forEach(param => {
    const values = props.sensors.map(s => s[param]).filter(v => isValidMetricValue(v));
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

  if (props.dateRange && props.dateRange.start && props.dateRange.end) {
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
  }

  // Fallback to timeSeriesData if dateRange is not available
  if (props.timeSeriesData && props.timeSeriesData.length > 0 && props.timeSeriesData[0].data && props.timeSeriesData[0].data.length > 0) {
    const firstPoint = props.timeSeriesData[0].data[0];
    const lastPoint = props.timeSeriesData[0].data[props.timeSeriesData[0].data.length - 1];

    if (firstPoint && lastPoint && firstPoint.time && lastPoint.time) {
      const start = new Date(firstPoint.time);
      const end = new Date(lastPoint.time);

      if (props.rangeType === 'hour') {
        return `Период: ${start.toLocaleDateString('ru-RU')} ${start.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })} - ${end.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}`;
      } else if (props.rangeType === 'day') {
        return `Период: ${start.toLocaleDateString('ru-RU', { day: '2-digit', month: 'short', year: 'numeric' })} - ${end.toLocaleDateString('ru-RU', { day: '2-digit', month: 'short', year: 'numeric' })}`;
      } else if (props.rangeType === 'month') {
        return `Период: ${start.toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' })} - ${end.toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' })}`;
      } else if (props.rangeType === 'year') {
        return `Период: ${start.getFullYear()} - ${end.getFullYear()}`;
      }
    }
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

const renderChart = () => {
  // Clear any pending resize timers
  if (resizeTimer) {
    clearTimeout(resizeTimer);
    resizeTimer = null;
  }

  if (selectedParams.value.length === 0) {
    if (chartInstance.value) {
      chartInstance.value.dispose();
      chartInstance.value = null;
    }
    return;
  }

  // Check if ref is ready
  if (!statsChart.value) {
    console.warn('statsChart ref not ready');
    return;
  }

  // Dispose old instance if exists
  if (chartInstance.value) {
    try {
      chartInstance.value.dispose();
    } catch (e) {
      console.warn('Failed to dispose chart:', e);
    }
    chartInstance.value = null;
  }

  // Initialize new instance
  try {
    chartInstance.value = echarts.init(statsChart.value);
  } catch (e) {
    console.error('Failed to init chart:', e);
    return;
  }

  // Check if we have time series data
  if (props.timeSeriesData && props.timeSeriesData.length > 0 && props.rangeType !== 'instant') {
    renderTimeSeriesChart();
  } else {
    renderInstantChart();
  }
};

const renderTimeSeriesChart = () => {
  console.log('📊 ДИАГНОСТИКА renderTimeSeriesChart:');
  console.log('  props.timeSeriesData.length:', props.timeSeriesData.length);
  console.log('  props.rangeType:', props.rangeType);
  console.log('  props.showIndividual:', props.showIndividual);
  console.log('  props.showIndividual type:', typeof props.showIndividual);
  console.log('  selectedParams:', selectedParams.value);
  console.log('  timeSeriesData[0]?.data?.length:', props.timeSeriesData[0]?.data?.length);
  console.log('  timeSeriesData[0]?.data (first 3):', props.timeSeriesData[0]?.data?.slice(0, 3));
  console.log('  All sites:', props.timeSeriesData.map(s => ({ name: s.name, dataPoints: s.data?.length })));

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

  // Check if we have valid data
  if (!props.timeSeriesData || props.timeSeriesData.length === 0) {
    console.warn('⚠️ No timeSeriesData available');
    return;
  }

  if (!props.timeSeriesData[0] || !props.timeSeriesData[0].data || props.timeSeriesData[0].data.length === 0) {
    console.warn('⚠️ First site has no data');
    return;
  }

  // Prepare data for each selected parameter
  const series = [];

  if (props.showIndividual) {
    // Show individual lines for each sensor
    console.log('  Mode: Individual lines for each sensor');
    selectedParams.value.forEach(param => {
      console.log(`  Processing param: ${param}`);
      props.timeSeriesData.forEach((site, index) => {
        console.log(`    Site ${index}: ${site.name}, data points: ${site.data?.length}`);
        const validData = site.data
          .map(d => d[param]) // Return just the value, not [time, value]
          .filter(val => isValidMetricValue(val));

        console.log(`    Valid data points for ${param}: ${validData.length}`);

        if (validData.length > 0) {
          series.push({
            name: `${site.name} - ${formatKey(param)}`,
            type: 'line',
            data: validData,
            smooth: true,
            symbolSize: 8,
            emphasis: {
              focus: 'series'
            },
            itemStyle: {
              color: colors[param]
            },
            lineStyle: {
              width: 2
            },
            yAxisIndex: selectedParams.value.indexOf(param)
          });
          console.log(`    ✅ Added series: ${site.name} - ${formatKey(param)}`);
        }
      });
    });
  } else {
    // Show averaged lines for each parameter
    console.log('  Mode: Averaged line');
    selectedParams.value.forEach(param => {
      // Get all time points from first site
      const timePoints = props.timeSeriesData[0]?.data.map(d => d.time) || [];

      // Calculate average for each time point
      const averagedData = timePoints.map(time => {
        const valuesAtTime = props.timeSeriesData
          .map(site => {
            const point = site.data.find(d => d.time === time);
            return point ? point[param] : null;
          })
          .filter(v => isValidMetricValue(v));

        if (valuesAtTime.length > 0) {
          const avg = valuesAtTime.reduce((a, b) => a + b, 0) / valuesAtTime.length;
          return avg; // Return just the value, not [time, value]
        }
        return null;
      }).filter(d => d !== null);

      if (averagedData.length > 0) {
        series.push({
          name: `Среднее - ${formatKey(param)}`,
          type: 'line',
          data: averagedData,
          smooth: true,
          symbolSize: 8,
          emphasis: {
            focus: 'series'
          },
          itemStyle: {
            color: colors[param]
          },
          lineStyle: {
            width: 3
          },
          yAxisIndex: selectedParams.value.indexOf(param)
        });
      }
    });
  }

  console.log('  series.length:', series.length);

  // Check if we have any valid series
  if (series.length === 0) {
    console.warn('⚠️ No valid series data to display');
    return;
  }

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

  console.log('  times.length:', times.length);
  console.log('  First 3 times:', times.slice(0, 3));

  // Check if we have time data
  if (times.length === 0) {
    console.warn('⚠️ No time data available');
    return;
  }

  // Create yAxis config for each selected parameter
  const yAxisConfig = selectedParams.value.map((param, index) => ({
    type: 'value',
    name: formatKey(param) + ' (' + units[param] + ')',
    position: index % 2 === 0 ? 'left' : 'right',
    offset: Math.floor(index / 2) * 60,
    nameLocation: 'middle',
    nameGap: 30,
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
      color: colors[param],
      fontSize: 11
    },
    splitLine: {
      show: index === 0,
      lineStyle: {
        color: '#e0e0e0',
        type: 'dashed'
      }
    }
  }));

  // Calculate dynamic grid margins
  const leftAxesCount = Math.ceil(selectedParams.value.length / 2);
  const rightAxesCount = Math.floor(selectedParams.value.length / 2);
  const gridLeft = leftAxesCount > 1 ? 60 + (leftAxesCount - 1) * 60 : 60;
  const gridRight = rightAxesCount > 0 ? 60 + (rightAxesCount - 1) * 60 : 60;

  const option = {
    title: {
      text: props.showIndividual ? 'Сравнение параметров (по отдельности)' : 'Сравнение параметров (средние)',
      left: 'center',
      top: 10
    },
    tooltip: {
      trigger: 'item',
      formatter: function(params) {
        const seriesName = params.seriesName;
        const value = params.value;
        const time = times[params.dataIndex];
        return `<strong>${seriesName}</strong><br/>${time}: ${value}`;
      }
    },
    legend: {
      data: series.map(s => s.name),
      top: 40,
      type: 'scroll',
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
      right: gridRight + 150,
      bottom: 80,
      top: 80,
      containLabel: false
    },
    xAxis: {
      type: 'category',
      data: times,
      boundaryGap: false,
      axisLabel: {
        rotate: 45,
        fontSize: 11
      }
    },
    yAxis: yAxisConfig,
    series: series,
    toolbox: {
      feature: {
        dataZoom: {
          yAxisIndex: 'none'
        },
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

  chartInstance.value.setOption(option);
};

const renderInstantChart = () => {
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

  // Create series for each selected parameter
  const series = [];
  const yAxisConfig = [];

  selectedParams.value.forEach((param, index) => {
    if (!statistics.value[param]) {
      return;
    }

    const stat = statistics.value[param];

    // Create single data point for aggregated average value
    const data = [stat.avg]; // Just the value, not [x, y] format for category axis

    series.push({
      name: formatKey(param),
      type: 'line',
      smooth: true,
      data: data,
      yAxisIndex: index,
      symbolSize: 8,
      emphasis: {
        focus: 'series'
      },
      itemStyle: {
        color: colors[param]
      },
      lineStyle: {
        width: 2
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
          }
        ],
        lineStyle: {
          type: 'dashed'
        }
      },
      markPoint: {
        data: [
          {
            yAxis: stat.max,
            xAxis: 0,
            name: 'Максимум',
            label: {
              formatter: 'Макс: {c}'
            }
          },
          {
            yAxis: stat.min,
            xAxis: 0,
            name: 'Минимум',
            label: {
              formatter: 'Мин: {c}'
            }
          }
        ]
      }
    });

    yAxisConfig.push({
      type: 'value',
      name: formatKey(param) + ' (' + units[param] + ')',
      position: index % 2 === 0 ? 'left' : 'right',
      offset: Math.floor(index / 2) * 60,
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
        color: colors[param],
        fontSize: 11
      },
      splitLine: {
        show: index === 0,
        lineStyle: {
          color: '#e0e0e0',
          type: 'dashed'
        }
      }
    });
  });

  // Use time label for X axis instead of sensor names
  const timeLabel = dateRangeText.value || 'Текущий момент';

  // Calculate dynamic grid margins - минимальные отступы
  const leftAxesCount = Math.ceil(selectedParams.value.length / 2);
  const rightAxesCount = Math.floor(selectedParams.value.length / 2);
  const gridLeft = leftAxesCount > 1 ? 60 + (leftAxesCount - 1) * 60 : 60;
  const gridRight = rightAxesCount > 0 ? 60 + (rightAxesCount - 1) * 60 : 60;

  const option = {
    title: {
      text: 'Сравнение параметров (средние значения)',
      left: 'center',
      top: 10
    },
    tooltip: {
      trigger: 'item',
      formatter: function(params) {
        const seriesName = params.seriesName;
        const value = params.value;
        const time = times[params.dataIndex];
        return `<strong>${seriesName}</strong><br/>${time}: ${value}`;
      }
    },
    legend: {
      data: selectedParams.value.map(p => formatKey(p)),
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
      right: gridRight + 150,
      bottom: 80,
      top: 80,
      containLabel: false
    },
    xAxis: {
      type: 'category',
      data: [timeLabel],
      boundaryGap: false,
      axisLabel: {
        rotate: 0,
        interval: 0,
        fontSize: 11
      }
    },
    yAxis: yAxisConfig,
    series: series,
    toolbox: {
      feature: {
        dataZoom: {
          yAxisIndex: 'none'
        },
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

  chartInstance.value.setOption(option);
};

watch(() => props.isOpen, (newVal) => {
  if (newVal) {
    nextTick(() => {
      // Small delay to ensure modal container has final size
      resizeTimer = setTimeout(() => {
        renderChart();
        // Resize chart after render to ensure correct dimensions
        nextTick(() => {
          if (chartInstance.value && !chartInstance.value.isDisposed()) {
            try {
              chartInstance.value.resize();
            } catch (e) {
              console.warn('Failed to resize chart:', e);
            }
          }
        });
      }, 50);
    });
  } else {
    // Clear timer when modal closes
    if (resizeTimer) {
      clearTimeout(resizeTimer);
      resizeTimer = null;
    }
  }
});

watch(selectedParams, () => {
  if (props.isOpen) {
    nextTick(() => {
      renderChart();
    });
  }
}, { deep: true });

watch(() => props.showIndividual, () => {
  if (props.isOpen) {
    nextTick(() => {
      renderChart();
    });
  }
});

onBeforeUnmount(() => {
  // Clear any pending timers
  if (resizeTimer) {
    clearTimeout(resizeTimer);
    resizeTimer = null;
  }

  // Dispose chart instance
  if (chartInstance.value) {
    try {
      chartInstance.value.dispose();
    } catch (e) {
      console.warn('Failed to dispose chart on unmount:', e);
    }
    chartInstance.value = null;
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
</style>
