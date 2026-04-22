<template>
  <div id="app">
    <MapView
      :sensors="sensors"
      :selectedIndicator="selectedIndicator"
      :selectionMode="selectionMode"
      :dateRange="selectedDateRange"
      @sensor-click="openModal"
    />
    <Indicators :selectedIndex="selectedIndicator" @indicator-selected="onIndicatorSelected" />
    <Legend :selectedIndicator="selectedIndicator" />
    <Timeline
      :date="selectedDate"
      :timePoints="timePoints"
      :dateRange="selectedDateRange"
      @time-selected="onTimeSelected"
      @range-selected="onRangeSelected"
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
      :timeSeriesData="timeSeriesData"
      :dateRange="statisticsDateRange"
      :rangeType="statisticsRangeType"
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
import { fetchAirQualityData, fetchAggregatedData, fetchTimeSeriesData, fetchAverageData } from './services/api';
import { formatDay, formatDayShort, formatMonth, formatYear } from './utils/dateFormat';

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
    const mode = ref('live');
    const selectionMode = ref('instant'); // 'instant' | 'range'
    const rangeStart = ref(null);
    const rangeEnd = ref(null);
    const timelineScale = ref('hour');
    const timeSeriesData = ref([]);
    const statisticsDateRange = ref(null);
    const statisticsRangeType = ref('instant');
    let autoRefreshInterval = null;

    const generateTimePoints = () => {
      const points = [];
      const baseDate = selectedDate.value || new Date();

      for (let i = 0; i < 24; i++) {
        const startDate = new Date(baseDate.getFullYear(), baseDate.getMonth(), baseDate.getDate(), i, 0, 0);
        const endDate = new Date(baseDate.getFullYear(), baseDate.getMonth(), baseDate.getDate(), i, 59, 59);

        points.push({
          type: 'hour',
          time: `${i.toString().padStart(2, '0')}:00`,
          color: '#5DADE2',
          hour: i,
          startDate,
          endDate
        });
      }
      timePoints.value = points;
    };

    const generateTimePointsForRange = (startDate, endDate) => {
      const points = [];
      const daysDiff = Math.ceil((endDate - startDate) / (1000 * 60 * 60 * 24));

      if (daysDiff === 0) {
        // Single day - show hours
        timelineScale.value = 'hour';
        for (let i = 0; i < 24; i++) {
          const pointStart = new Date(startDate.getFullYear(), startDate.getMonth(), startDate.getDate(), i, 0, 0);
          const pointEnd = new Date(startDate.getFullYear(), startDate.getMonth(), startDate.getDate(), i, 59, 59);

          points.push({
            type: 'hour',
            time: `${i.toString().padStart(2, '0')}:00`,
            color: '#5DADE2',
            hour: i,
            startDate: pointStart,
            endDate: pointEnd
          });
        }
      } else if (daysDiff <= 31) {
        // Days range
        timelineScale.value = 'day';
        const current = new Date(startDate);
        while (current <= endDate) {
          const pointStart = new Date(current.getFullYear(), current.getMonth(), current.getDate(), 0, 0, 0);
          const pointEnd = new Date(current.getFullYear(), current.getMonth(), current.getDate(), 23, 59, 59);

          points.push({
            type: 'day',
            time: formatDayShort(current),
            color: '#5DADE2',
            startDate: new Date(pointStart),
            endDate: new Date(pointEnd)
          });
          current.setDate(current.getDate() + 1);
        }
      } else if (daysDiff <= 365) {
        // Months range
        timelineScale.value = 'month';
        const current = new Date(startDate.getFullYear(), startDate.getMonth(), 1);
        const end = new Date(endDate.getFullYear(), endDate.getMonth(), 1);

        while (current <= end) {
          const pointStart = new Date(current.getFullYear(), current.getMonth(), 1, 0, 0, 0);
          const pointEnd = new Date(current.getFullYear(), current.getMonth() + 1, 0, 23, 59, 59);

          points.push({
            type: 'month',
            time: formatMonth(current),
            color: '#5DADE2',
            startDate: new Date(pointStart),
            endDate: new Date(pointEnd)
          });
          current.setMonth(current.getMonth() + 1);
        }
      } else {
        // Years range
        timelineScale.value = 'year';
        const currentYear = startDate.getFullYear();
        const endYear = endDate.getFullYear();

        for (let year = currentYear; year <= endYear; year++) {
          const pointStart = new Date(year, 0, 1, 0, 0, 0);
          const pointEnd = new Date(year, 11, 31, 23, 59, 59);

          points.push({
            type: 'year',
            time: formatYear(pointStart),
            color: '#5DADE2',
            startDate: pointStart,
            endDate: pointEnd
          });
        }
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

        const daysDiff = Math.ceil((endDate - startDate) / (1000 * 60 * 60 * 24));
        let interval = 'hour';

        if (daysDiff === 0) {
          interval = 'hour';
        } else if (daysDiff <= 31) {
          interval = 'day';
        } else if (daysDiff <= 365) {
          interval = 'month';
        } else {
          interval = 'month';
        }

        const data = await fetchAggregatedData(startDate, endDate, interval);
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
      selectionMode.value = 'instant';

      if (!timePoint) return;

      if (timePoint.type === 'hour') {
        // Load data for specific hour
        loadData(selectedDate.value, timePoint.hour);
      } else if (timePoint.type === 'day' || timePoint.type === 'month' || timePoint.type === 'year') {
        // Load aggregated data for the period
        loadAggregatedData(timePoint.startDate, timePoint.endDate);
      }
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

      mode.value = 'single-day';
      selectedDate.value = date;
      selectedDateRange.value = null;
      rangeStart.value = null;
      rangeEnd.value = null;
      generateTimePoints();
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
          const endDay = new Date(range.end.getFullYear(), range.end.getMonth(), range.end.getDate());

          if (range.start.getTime() === range.end.getTime()) {
            mode.value = 'single-day';
            selectedDateRange.value = null;
            selectedDate.value = range.start;
            rangeStart.value = null;
            rangeEnd.value = null;
            generateTimePoints();
            const hour = selectedTimePoint.value ? selectedTimePoint.value.hour : 0;
            loadData(range.start, hour);
          } else {
            mode.value = 'range';
            selectedDateRange.value = range;
            selectedDate.value = range.start;
            rangeStart.value = range.start;
            rangeEnd.value = range.end;
            generateTimePointsForRange(range.start, range.end);
            loadAggregatedData(range.start, range.end);
          }
        } else {
          mode.value = 'single-day';
          selectedDateRange.value = null;
          selectedDate.value = range.start;
          rangeStart.value = null;
          rangeEnd.value = null;
          generateTimePoints();
          const hour = selectedTimePoint.value ? selectedTimePoint.value.hour : 0;
          loadData(range.start, hour);
        }
      }
    };

    const openStatisticsModal = async (selectedSensors, districtKey) => {
      selectedSensorsForStats.value = selectedSensors;

      // If in range mode and district is selected, load time series for district
      if (selectionMode.value === 'range' && selectedDateRange.value && districtKey) {
        const siteIds = selectedSensors.map(s => s.id);

        let interval = timelineScale.value || 'hour';

        try {
          // Fetch time series data for selected district sensors
          const data = await fetchTimeSeriesData(
            selectedDateRange.value.start,
            selectedDateRange.value.end,
            interval,
            siteIds,
            null
          );

          console.log('Loaded time series data for district:', data);

          timeSeriesData.value = data;
          statisticsDateRange.value = {
            start: selectedDateRange.value.start,
            end: selectedDateRange.value.end
          };
          statisticsRangeType.value = interval;
        } catch (error) {
          console.error('Failed to load district time series:', error);
        }
      }

      isStatisticsModalOpen.value = true;
    };

    const onRangeSelected = async (rangeData) => {
      console.log('Range selected in Timeline:', rangeData);

      const { start, end, points } = rangeData;

      // Switch to range mode
      selectionMode.value = 'range';

      // Determine interval based on point type
      let interval = 'hour';
      if (start.type === 'day') interval = 'day';
      else if (start.type === 'month') interval = 'month';
      else if (start.type === 'year') interval = 'year';

      // Store interval for later use
      timelineScale.value = interval;

      // Extract site IDs from current sensors
      const siteIds = sensors.value.map(s => s.id);

      try {
        // Load average data for map display (all sensors)
        const averageData = await fetchAverageData(
          start.startDate,
          end.endDate,
          interval,
          siteIds.length > 0 ? siteIds : null,
          null
        );

        console.log('Loaded average data for map:', averageData);
        sensors.value = averageData;

        // Don't automatically open modal - wait for user to select district/sensors
        console.log('Range selected. Select district or sensors to view statistics.');
      } catch (error) {
        console.error('Failed to load time series data:', error);
      }
    };

    const closeStatisticsModal = () => {
      isStatisticsModalOpen.value = false;
    };

    onMounted(() => {
      mode.value = 'live';
      generateTimePoints();
      loadData();

      autoRefreshInterval = setInterval(() => {
        if (mode.value === 'live') {
          loadData();
        }
      }, 300000);
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
      timeSeriesData,
      statisticsDateRange,
      statisticsRangeType,
      selectionMode,
      openModal,
      closeModal,
      onTimeSelected,
      onRangeSelected,
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
