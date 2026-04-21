<template>
  <div id="app">
    <MapView :sensors="sensors" :selectedIndicator="selectedIndicator" @sensor-click="openModal" />
    <Indicators :selectedIndex="selectedIndicator" @indicator-selected="onIndicatorSelected" />
    <Legend :selectedIndicator="selectedIndicator" />
    <Timeline
      :date="selectedDate"
      :timePoints="timePoints"
      @time-selected="onTimeSelected"
      @open-calendar="isCalendarOpen = true"
    />
    <Calendar
      :isOpen="isCalendarOpen"
      :selectedDate="selectedDate"
      @date-selected="onDateSelected"
      @date-range-selected="onDateRangeSelected"
      @close="isCalendarOpen = false"
    />
    <SensorModal
      :isOpen="isModalOpen"
      :sensorData="selectedSensor"
      @close="closeModal"
    />
    <SidePanel
      :sensors="sensors"
      @show-statistics="openStatisticsModal"
    />
    <StatisticsModal
      :isOpen="isStatisticsModalOpen"
      :sensors="selectedSensorsForStats"
      @close="closeStatisticsModal"
    />
  </div>
</template>

<script>
import { ref, onMounted } from 'vue';
import MapView from './components/MapView.vue';
import Legend from './components/Legend.vue';
import Timeline from './components/Timeline.vue';
import Indicators from './components/Indicators.vue';
import Calendar from './components/Calendar.vue';
import SensorModal from './components/SensorModal.vue';
import SidePanel from './components/SidePanel.vue';
import StatisticsModal from './components/StatisticsModal.vue';
import { fetchAirQualityData } from './services/api';

export default {
  name: 'App',
  components: {
    MapView,
    Legend,
    Timeline,
    Indicators,
    Calendar,
    SensorModal,
    SidePanel,
    StatisticsModal
  },
  setup() {
    const sensors = ref([]);
    const isModalOpen = ref(false);
    const isCalendarOpen = ref(false);
    const isStatisticsModalOpen = ref(false);
    const selectedSensor = ref({});
    const selectedSensorsForStats = ref([]);
    const selectedDate = ref(new Date());
    const selectedIndicator = ref(0);
    const selectedTimePoint = ref(null);
    const selectedDateRange = ref(null);
    const timePoints = ref([]);

    const generateTimePoints = () => {
      const points = [];
      for (let i = 0; i < 24; i++) {
        points.push({
          time: `${i.toString().padStart(2, '0')}:00`,
          color: '#5DADE2',
          hour: i
        });
      }
      timePoints.value = points;
    };

    const loadData = async (date = null, hour = null) => {
      try {
        const data = await fetchAirQualityData(date, hour);
        console.log('Loaded sensor data:', data);
        sensors.value = data;
      } catch (error) {
        console.error('Failed to load air quality data:', error);
        sensors.value = [
          {
            id: 1,
            name: 'Krasnoyarsk Center',
            latitude: 56.0153,
            longitude: 92.8672,
            pm25: 45,
            pm10: 60,
            temperature: 15,
            humidity: 65,
            pressure: 1013
          },
          {
            id: 2,
            name: 'Krasnoyarsk North',
            latitude: 56.0453,
            longitude: 92.8872,
            pm25: 25,
            pm10: 35,
            temperature: 14,
            humidity: 70,
            pressure: 1012
          },
          {
            id: 3,
            name: 'Krasnoyarsk South',
            latitude: 55.9853,
            longitude: 92.8472,
            pm25: 65,
            pm10: 85,
            temperature: 16,
            humidity: 60,
            pressure: 1014
          }
        ];
      }
    };

    const loadAggregatedData = async (startDate, endDate) => {
      try {
        console.log('Loading aggregated data for range:', startDate, 'to', endDate);
        // For now, load data for the start date
        // TODO: Implement proper aggregated data loading from backend
        const data = await fetchAirQualityData(startDate, 12);
        console.log('Loaded aggregated sensor data:', data);
        sensors.value = data;
      } catch (error) {
        console.error('Failed to load aggregated data:', error);
        sensors.value = [];
      }
    };

    const openModal = (sensorData) => {
      selectedSensor.value = sensorData;
      isModalOpen.value = true;
    };

    const closeModal = () => {
      isModalOpen.value = false;
    };

    const onTimeSelected = (timePoint) => {
      console.log('Time selected:', timePoint);
      selectedTimePoint.value = timePoint;
      loadData(selectedDate.value, timePoint.hour);
    };

    const onIndicatorSelected = (index) => {
      selectedIndicator.value = index;
      console.log('Indicator selected:', index);
    };

    const onDateSelected = (date) => {
      const now = new Date();
      const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
      const selectedDay = new Date(date.getFullYear(), date.getMonth(), date.getDate());

      if (selectedDay > today) {
        console.warn('Cannot select future date');
        alert('Невозможно выбрать будущую дату. Данные доступны только за прошедшие периоды.');
        return;
      }

      selectedDate.value = date;
      console.log('Date selected:', date);
      const hour = selectedTimePoint.value ? selectedTimePoint.value.hour : 0;
      loadData(date, hour);
    };

    const onDateRangeSelected = (range) => {
      console.log('Date range selected:', range);

      const now = new Date();
      const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());

      if (range.start) {
        const startDay = new Date(range.start.getFullYear(), range.start.getMonth(), range.start.getDate());

        if (startDay > today) {
          console.warn('Cannot select future date range');
          alert('Невозможно выбрать будущую дату. Данные доступны только за прошедшие периоды.');
          return;
        }

        if (range.end) {
          // Range mode: load aggregated data
          selectedDateRange.value = range;
          selectedDate.value = range.start;
          loadAggregatedData(range.start, range.end);
        } else {
          // Single date mode
          selectedDateRange.value = null;
          selectedDate.value = range.start;
          const hour = selectedTimePoint.value ? selectedTimePoint.value.hour : 0;
          loadData(range.start, hour);
        }
      }
    };

    const openStatisticsModal = (selectedSensors) => {
      selectedSensorsForStats.value = selectedSensors;
      isStatisticsModalOpen.value = true;
    };

    const closeStatisticsModal = () => {
      isStatisticsModalOpen.value = false;
    };

    onMounted(() => {
      generateTimePoints();
      loadData();
      setInterval(loadData, 300000);
    });

    return {
      sensors,
      isModalOpen,
      isCalendarOpen,
      isStatisticsModalOpen,
      selectedSensor,
      selectedSensorsForStats,
      selectedDate,
      selectedDateRange,
      selectedIndicator,
      selectedTimePoint,
      timePoints,
      openModal,
      closeModal,
      onTimeSelected,
      onIndicatorSelected,
      onDateSelected,
      onDateRangeSelected,
      openStatisticsModal,
      closeStatisticsModal
    };
  }
};
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

#app {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  width: 100%;
  height: 100vh;
  overflow: hidden;
}
</style>
