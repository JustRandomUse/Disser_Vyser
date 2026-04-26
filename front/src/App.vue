<template>
  <div id="app">
    <MapView
      :sensors="mapSensors"
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
      :timeSeriesData="sensorTimeSeriesData"
      :dateRange="sensorModalDateRange"
      :rangeType="sensorModalRangeType"
      @close="closeModal"
    />
    <SidePanel
      :sensors="baseSensors"
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
import { formatDateISO, formatMonthISO, formatYear } from './utils/dateFormat';

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
    const baseSensors = ref([]); // Static list of unique sensors for sidebar
    const mapSensors = ref([]); // Dynamic data for map display
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
    const sensorTimeSeriesData = ref([]); // Time series data for single sensor modal
    const sensorModalDateRange = ref(null); // Date range specifically for sensor modal
    const sensorModalRangeType = ref('instant'); // Range type for sensor modal
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
            time: formatDateISO(current),
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
            time: formatMonthISO(current),
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

        // Update both base sensors (if not yet loaded) and map sensors
        if (baseSensors.value.length === 0) {
          baseSensors.value = data.map(s => ({
            id: s.id,
            name: s.name,
            latitude: s.latitude,
            longitude: s.longitude
          }));
        }

        // If no data returned but we have base sensors, show stations with null values
        if (data.length === 0 && baseSensors.value.length > 0) {
          console.warn('⚠️ No data for selected period, showing stations with null values');
          mapSensors.value = baseSensors.value.map(sensor => ({
            ...sensor,
            pm25: null,
            pm10: null,
            temperature: null,
            humidity: null,
            pressure: null,
            aqi: null,
            time: null,
            noData: true
          }));
        } else {
          mapSensors.value = data;
        }
      } catch (error) {
        console.error('Failed to load air quality data:', error);

        alert('Ошибка загрузки данных о качестве воздуха.\n\nПроверьте подключение к интернету или попробуйте позже.');

        const fallbackData = [
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

        if (baseSensors.value.length === 0) {
          baseSensors.value = fallbackData.map(s => ({
            id: s.id,
            name: s.name,
            latitude: s.latitude,
            longitude: s.longitude
          }));
        }

        mapSensors.value = fallbackData;
      }
    };

    const loadAggregatedData = async (startDate, endDate) => {
      try {
        console.log('🔍 loadAggregatedData called:');
        console.log('  startDate:', startDate);
        console.log('  endDate:', endDate);
        console.log('  startDate year:', startDate.getFullYear());

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

        console.log('  daysDiff:', daysDiff);
        console.log('  interval:', interval);

        // Use fetchAverageData instead of fetchAggregatedData to get one record per sensor
        const siteIds = baseSensors.value.length > 0 ? baseSensors.value.map(s => s.id) : null;
        console.log('  siteIds:', siteIds);

        const data = await fetchAverageData(startDate, endDate, interval, siteIds, null);
        console.log('  ✅ Loaded average sensor data count:', data.length);
        console.log('  ✅ Loaded average sensor data:', data);

        if (data.length === 0) {
          console.warn('⚠️ No data available for selected period');

          const startDateStr = startDate.toLocaleDateString('ru-RU', { day: '2-digit', month: 'short', year: 'numeric' });
          const endDateStr = endDate.toLocaleDateString('ru-RU', { day: '2-digit', month: 'short', year: 'numeric' });

          if (startDateStr === endDateStr) {
            alert(`Нет данных за ${startDateStr}.\n\nДанные доступны только за текущий период. Исторические данные могут быть недоступны.`);
          } else {
            alert(`Нет данных за период ${startDateStr} - ${endDateStr}.\n\nДанные доступны только за текущий период. Исторические данные могут быть недоступны.`);
          }
        }

        mapSensors.value = data;
      } catch (error) {
        console.error('❌ Failed to load aggregated data:', error);
        mapSensors.value = [];
      }
    };

    const openModal = async (sensorData) => {
      selectedSensor.value = sensorData;

      // Determine date range and range type for sensor modal
      let startDate, endDate, rangeType, interval;

      console.log('🔍 openModal called with:');
      console.log('  selectedTimePoint:', selectedTimePoint.value);
      console.log('  selectedDateRange:', selectedDateRange.value);
      console.log('  selectedDate:', selectedDate.value);

      if (selectedTimePoint.value) {
        // User selected a specific time point on timeline
        startDate = selectedTimePoint.value.startDate;
        endDate = selectedTimePoint.value.endDate;
        rangeType = selectedTimePoint.value.type;

        // If user selected a specific hour, expand to full day for modal
        if (rangeType === 'hour') {
          const selectedDay = new Date(startDate.getFullYear(), startDate.getMonth(), startDate.getDate());
          startDate = new Date(selectedDay.getFullYear(), selectedDay.getMonth(), selectedDay.getDate(), 0, 0, 0);
          endDate = new Date(selectedDay.getFullYear(), selectedDay.getMonth(), selectedDay.getDate(), 23, 59, 59);
          rangeType = 'day'; // Treat as day for modal purposes
        }

        console.log('  ✅ Using selectedTimePoint (expanded to full day if hour)');
      } else if (selectedDateRange.value && selectedDateRange.value.start && selectedDateRange.value.end) {
        // User selected a range (from calendar or range selection)
        startDate = selectedDateRange.value.start;
        endDate = selectedDateRange.value.end;
        rangeType = timelineScale.value || 'hour';
        console.log('  ✅ Using selectedDateRange');
      } else {
        // No timeline selection - fallback to current day with hourly breakdown
        const currentDate = selectedDate.value || new Date();
        startDate = new Date(currentDate.getFullYear(), currentDate.getMonth(), currentDate.getDate(), 0, 0, 0);
        endDate = new Date(currentDate.getFullYear(), currentDate.getMonth(), currentDate.getDate(), 23, 59, 59);
        rangeType = 'hour';
        console.log('  ✅ Using fallback (current day)');
      }

      console.log('  startDate:', startDate);
      console.log('  endDate:', endDate);
      console.log('  rangeType:', rangeType);

      // Store for passing to SensorModal
      sensorModalDateRange.value = { start: startDate, end: endDate };
      sensorModalRangeType.value = rangeType;

      // Check if selected date is today
      const now = new Date();
      const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
      const selectedDay = new Date(startDate.getFullYear(), startDate.getMonth(), startDate.getDate());
      const isToday = selectedDay.getTime() === today.getTime();

      console.log('  isToday:', isToday);

      // For SensorModal, determine interval based on whether it's today or historical
      if (rangeType === 'day' || rangeType === 'hour') {
        // For a single day selection, always try hourly data first
        // API will return empty if not available, then we'll fallback to daily
        interval = 'hour';
      } else if (rangeType === 'month') {
        // For a specific month, show daily data throughout the month
        interval = 'day';
      } else if (rangeType === 'year') {
        // For a specific year, show monthly data throughout the year
        interval = 'month';
      } else {
        interval = 'hour'; // fallback
      }

      console.log('  interval:', interval);

      console.log('📊 openModal summary:', {
        rangeType,
        interval,
        startDate: startDate.toISOString(),
        endDate: endDate.toISOString(),
        selectedTimePoint: selectedTimePoint.value?.type,
        selectionMode: selectionMode.value,
        isToday
      });

      try {
        console.log('🌐 Calling fetchTimeSeriesData with:', {
          startDate: startDate.toISOString(),
          endDate: endDate.toISOString(),
          interval,
          sensorId: sensorData.id
        });

        const data = await fetchTimeSeriesData(
          startDate,
          endDate,
          interval,
          [sensorData.id],
          null
        );

        console.log('✅ Loaded time series data for sensor modal:', data);
        console.log('  data.length:', data.length);
        if (data.length > 0) {
          console.log('  data[0].data.length:', data[0].data?.length);
          console.log('  First 3 points:', data[0].data?.slice(0, 3));
        }

        // If API returned empty data for hourly interval, try daily interval
        if ((data.length === 0 || (data.length > 0 && data[0].data.length === 0)) && interval === 'hour') {
          console.warn('⚠️ No hourly data available, trying daily interval...');

          const dailyData = await fetchTimeSeriesData(
            startDate,
            endDate,
            'day',
            [sensorData.id],
            null
          );

          console.log('✅ Loaded daily time series data:', dailyData);
          if (dailyData.length > 0 && dailyData[0].data.length > 0) {
            sensorTimeSeriesData.value = dailyData;
          } else {
            // Still no data, use fallback
            const dateStr = startDate.toLocaleDateString('ru-RU', { day: '2-digit', month: 'short', year: 'numeric' });
            console.warn(`⚠️ No data available for ${dateStr}, using current sensor data as single point`);

            sensorTimeSeriesData.value = [{
              id: sensorData.id,
              name: sensorData.name,
              data: [{
                time: startDate.toISOString(),
                pm25: sensorData.pm25,
                pm10: sensorData.pm10,
                temperature: sensorData.temperature,
                humidity: sensorData.humidity,
                pressure: sensorData.pressure,
                aqi: sensorData.aqi
              }]
            }];
          }
        } else if (data.length === 0 || (data.length > 0 && data[0].data.length === 0)) {
          // No data for other intervals, use fallback
          const dateStr = startDate.toLocaleDateString('ru-RU', { day: '2-digit', month: 'short', year: 'numeric' });
          console.warn(`⚠️ No time series data available for ${dateStr}, using current sensor data as single point`);

          sensorTimeSeriesData.value = [{
            id: sensorData.id,
            name: sensorData.name,
            data: [{
              time: startDate.toISOString(),
              pm25: sensorData.pm25,
              pm10: sensorData.pm10,
              temperature: sensorData.temperature,
              humidity: sensorData.humidity,
              pressure: sensorData.pressure,
              aqi: sensorData.aqi
            }]
          }];
        } else {
          sensorTimeSeriesData.value = data;
        }
      } catch (error) {
        const dateStr = startDate.toLocaleDateString('ru-RU', { day: '2-digit', month: 'short', year: 'numeric' });
        console.error(`❌ Failed to load sensor time series data for ${dateStr}:`, error);

        // Fallback: use current sensor data as single point
        sensorTimeSeriesData.value = [{
          id: sensorData.id,
          name: sensorData.name,
          data: [{
            time: startDate.toISOString(),
            pm25: sensorData.pm25,
            pm10: sensorData.pm10,
            temperature: sensorData.temperature,
            humidity: sensorData.humidity,
            pressure: sensorData.pressure,
            aqi: sensorData.aqi
          }]
        }];

        // Show error message to user
        alert(`Ошибка загрузки данных за ${dateStr}.\n\nОтображаются текущие значения станции.`);
      }

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

      // Set selectedDateRange for the selected time point
      selectedDateRange.value = {
        start: timePoint.startDate,
        end: timePoint.endDate
      };

      if (timePoint.type === 'hour') {
        // Load data for specific hour
        loadData(selectedDate.value, timePoint.hour);
      } else if (timePoint.type === 'day') {
        // For a specific day, load aggregated data for that day
        // This will show stations on the map with averaged values for the day
        loadAggregatedData(timePoint.startDate, timePoint.endDate);
      } else if (timePoint.type === 'month' || timePoint.type === 'year') {
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

    const openStatisticsModal = async (selectedSensors, districtKey, dateRange) => {
      // Map selected sensor IDs to full sensor objects from baseSensors
      selectedSensorsForStats.value = selectedSensors.map(s => {
        const baseSensor = baseSensors.value.find(bs => bs.id === s.id);
        const mapSensor = mapSensors.value.find(ms => ms.id === s.id);

        return {
          id: s.id,
          name: baseSensor ? baseSensor.name : s.name,
          latitude: baseSensor ? baseSensor.latitude : s.latitude,
          longitude: baseSensor ? baseSensor.longitude : s.longitude,
          pm25: mapSensor ? mapSensor.pm25 : s.pm25 || 0,
          pm10: mapSensor ? mapSensor.pm10 : s.pm10 || 0,
          temperature: mapSensor ? mapSensor.temperature : s.temperature || 0,
          humidity: mapSensor ? mapSensor.humidity : s.humidity || 0,
          pressure: mapSensor ? mapSensor.pressure : s.pressure || 0
        };
      });

      // Use dateRange from SidePanel if provided, otherwise use current selection
      const effectiveDateRange = dateRange || selectedDateRange.value;
      const siteIds = selectedSensors.map(s => s.id);

      // If we have a date range, load time series data
      if (effectiveDateRange) {
        // Determine interval based on date range duration
        const startDate = new Date(effectiveDateRange.start);
        const endDate = new Date(effectiveDateRange.end);
        const daysDiff = Math.ceil((endDate - startDate) / (1000 * 60 * 60 * 24));

        let interval = 'hour';
        if (daysDiff > 7) {
          interval = 'day';
        }
        if (daysDiff > 90) {
          interval = 'month';
        }

        console.log('🔍 ДИАГНОСТИКА openStatisticsModal:');
        console.log('  effectiveDateRange:', effectiveDateRange);
        console.log('  daysDiff:', daysDiff);
        console.log('  interval (API):', interval);
        console.log('  siteIds:', siteIds);

        try {
          // Fetch time series data for selected sensors
          const data = await fetchTimeSeriesData(
            startDate,
            endDate,
            interval,
            siteIds,
            null
          );

          console.log('  ✅ API response length:', data.length);
          console.log('  ✅ First site data points:', data[0]?.data?.length);
          console.log('  ✅ First 3 points:', data[0]?.data?.slice(0, 3));
          console.log('Loaded time series data for selected sensors:', data);

          timeSeriesData.value = data;
          statisticsDateRange.value = {
            start: startDate,
            end: endDate
          };

          // Determine range type based on interval
          if (interval === 'hour') {
            statisticsRangeType.value = 'hour';
          } else if (interval === 'day') {
            statisticsRangeType.value = 'day';
          } else if (interval === 'month') {
            statisticsRangeType.value = 'month';
          }
        } catch (error) {
          console.error('❌ Failed to load time series data:', error);

          const startDateStr = startDate.toLocaleDateString('ru-RU', { day: '2-digit', month: 'short', year: 'numeric' });
          const endDateStr = endDate.toLocaleDateString('ru-RU', { day: '2-digit', month: 'short', year: 'numeric' });

          alert(`Ошибка загрузки данных за период ${startDateStr} - ${endDateStr}.\n\nПопробуйте выбрать другой период или обновите страницу.`);

          timeSeriesData.value = [];
        }
      } else {
        // Clear time series data for instant mode
        timeSeriesData.value = [];
        statisticsDateRange.value = null;
        statisticsRangeType.value = 'instant';
      }

      isStatisticsModalOpen.value = true;
    };

    const onRangeSelected = async (rangeData) => {
      console.log('Range selected in Timeline:', rangeData);

      const { start, end, points } = rangeData;

      // Switch to range mode
      selectionMode.value = 'range';

      // Store selected date range for statistics modal
      selectedDateRange.value = {
        start: start.startDate,
        end: end.endDate
      };

      // Store range boundaries
      rangeStart.value = start.startDate;
      rangeEnd.value = end.endDate;
      selectedDate.value = start.startDate;

      // Determine interval based on point type
      // Note: API supports only 'hour', 'day', 'month' - not 'year'
      let interval = 'hour';
      if (start.type === 'day') interval = 'day';
      else if (start.type === 'month') interval = 'month';
      else if (start.type === 'year') interval = 'month'; // Use 'month' for year ranges

      // Store interval for later use
      timelineScale.value = start.type; // Keep original type for UI

      // Extract site IDs from base sensors (static list)
      const siteIds = baseSensors.value.length > 0 ? baseSensors.value.map(s => s.id) : null;

      try {
        // Load average data for map display (one record per sensor)
        const averageData = await fetchAverageData(
          start.startDate,
          end.endDate,
          interval,
          siteIds,
          null
        );

        console.log('Loaded average data for map:', averageData);
        mapSensors.value = averageData;

        // Don't automatically open modal - wait for user to select district/sensors
        console.log('Range selected. Select district or sensors to view statistics.');
      } catch (error) {
        console.error('Failed to load average data:', error);

        const startDateStr = start.startDate.toLocaleDateString('ru-RU', { day: '2-digit', month: 'short', year: 'numeric' });
        const endDateStr = end.endDate.toLocaleDateString('ru-RU', { day: '2-digit', month: 'short', year: 'numeric' });

        alert(`Ошибка загрузки данных за период ${startDateStr} - ${endDateStr}.\n\nПопробуйте выбрать другой период.`);

        mapSensors.value = [];
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
      baseSensors,
      mapSensors,
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
      sensorTimeSeriesData,
      sensorModalDateRange,
      sensorModalRangeType,
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
