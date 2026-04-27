<template>
  <div class="timeline">
    <button class="calendar-btn" @click="$emit('open-calendar')">
      <span class="calendar-icon">📅</span>
      {{ formattedDate }}
    </button>

    <span class="page-counter">{{ currentPage }}/{{ totalPages }}</span>
    
    <button class="nav-btn" @click="previousPage" :disabled="currentPage === 1">
      ←
    </button>
    
    <div class="timeline-body" :style="{ gridTemplateColumns: `repeat(${pointsPerPage}, 1fr)` }">
      <div class="line"></div>
      <div
        v-for="(point, index) in visiblePoints"
        :key="index"
        class="point"
        :class="{
          selected: index === selectedIndex,
          'in-range': isInSelectedRange(index)
        }"
        @click="selectPoint(index)"
      >
        <div class="point-circle" :style="{ background: point.color }"></div>
        <div class="point-text">{{ point.time }}</div>
      </div>
    </div>
    
    <button class="nav-btn" @click="nextPage" :disabled="currentPage === totalPages">
      →
    </button>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import { formatDateISO, formatDateRangeISO } from '../utils/dateFormat';

const props = defineProps({
  date: {
    type: Date,
    default: () => new Date()
  },
  timePoints: {
    type: Array,
    default: () => []
  },
  dateRange: {
    type: Object,
    default: null
  },
  selectedTimePoint: {
    type: Object,
    default: null
  }
});

const emit = defineEmits(['open-calendar', 'time-selected', 'range-selected']);

const currentPage = ref(1);
const pointsPerPage = ref(15);
const selectedIndex = ref(-1); // -1 means no selection

// Calculate points per page based on window width
const calculatePointsPerPage = () => {
  const width = window.innerWidth;
  // Reserve space for calendar button (~190px), counter (~50px), nav buttons (~64px), gaps (~48px)
  const reservedSpace = 352;
  const availableWidth = width - reservedSpace;
  // Each point needs approximately 60px (circle + text + gap)
  const pointWidth = 60;
  const points = Math.floor(availableWidth / pointWidth);
  // Minimum 5 points, maximum 20 points
  pointsPerPage.value = Math.max(5, Math.min(20, points));
};

// Watch for window resize
let resizeTimeout;
const handleResize = () => {
  clearTimeout(resizeTimeout);
  resizeTimeout = setTimeout(() => {
    const oldPointsPerPage = pointsPerPage.value;
    calculatePointsPerPage();
    // If points per page changed, recalculate current page
    if (oldPointsPerPage !== pointsPerPage.value) {
      currentPage.value = 1;
      selectedIndex.value = -1;
    }
  }, 200);
};

onMounted(() => {
  calculatePointsPerPage();
  window.addEventListener('resize', handleResize);
});

onUnmounted(() => {
  window.removeEventListener('resize', handleResize);
  clearTimeout(resizeTimeout);
});

// Watch for date changes and reset pagination
watch(() => props.date, () => {
  currentPage.value = 1;
  selectedIndex.value = -1; // -1 means no selection
});

// Watch for selectedTimePoint changes from parent
watch(() => props.selectedTimePoint, (newTimePoint) => {
  if (newTimePoint && props.timePoints.length > 0) {
    // Find the index of the selected time point
    const globalIndex = props.timePoints.findIndex(p => p.time === newTimePoint.time);
    if (globalIndex !== -1) {
      // Calculate which page this point is on
      const page = Math.floor(globalIndex / pointsPerPage.value) + 1;
      currentPage.value = page;
      // Calculate local index on current page
      selectedIndex.value = globalIndex % pointsPerPage.value;
    }
  }
}, { immediate: true });

const formattedDate = computed(() => {
  if (props.dateRange && props.dateRange.start && props.dateRange.end) {
    return formatDateRangeISO(props.dateRange.start, props.dateRange.end);
  }

  return formatDateISO(props.date);
});

const totalPages = computed(() => {
  return Math.ceil(props.timePoints.length / pointsPerPage.value);
});

const visiblePoints = computed(() => {
  const start = (currentPage.value - 1) * pointsPerPage.value;
  const end = start + pointsPerPage.value;
  return props.timePoints.slice(start, end);
});

const selectPoint = (index) => {
  const globalIndex = (currentPage.value - 1) * pointsPerPage.value + index;
  const point = props.timePoints[globalIndex];

  selectedIndex.value = index;
  emit('time-selected', point);
};

const isInSelectedRange = (index) => {
  return false;
};

const previousPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--;
    selectedIndex.value = -1; // Clear selection when changing page
  }
};

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++;
    selectedIndex.value = -1; // Clear selection when changing page
  }
};
</script>

<style scoped>
.timeline {
  position: absolute;
  bottom: 0;
  width: 100%;
  height: 60px;
  background: white;
  display: flex;
  align-items: center;
  padding: 0.5rem;
  gap: 1rem;
  z-index: 2;
}

.calendar-btn {
  width: 190px;
  padding: 0.375rem 0.625rem;
  border: 1px solid #ccc;
  border-radius: 6px;
  background: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 14px;
}

.calendar-btn:hover {
  background: #f5f5f5;
}

.calendar-icon {
  font-size: 16px;
}

.page-counter {
  font-size: 14px;
  color: #333;
}

.nav-btn {
  padding: 0.375rem;
  border: none;
  border-radius: 6px;
  background: #3b82f6;
  color: white;
  cursor: pointer;
  font-size: 20px;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.nav-btn:hover:not(:disabled) {
  background: #2563eb;
}

.nav-btn:disabled {
  background: #9ca3af;
  cursor: not-allowed;
}

.timeline-body {
  display: grid;
  place-items: center;
  position: relative;
  width: 100%;
  gap: 2px;
}

.line {
  position: absolute;
  top: 8px;
  left: 0;
  right: 0;
  height: 4px;
  background: #ccc;
  z-index: 0;
}

.point {
  display: flex;
  flex-direction: column;
  align-items: center;
  position: relative;
  z-index: 1;
  cursor: pointer;
  font-size: 1rem;
  white-space: nowrap;
}

.point:hover .point-circle {
  background: #ccc !important;
}

.point-circle {
  width: 20px;
  height: 20px;
  border-radius: 20px;
  border: 2px solid #000;
}

.point.selected .point-circle {
  box-shadow: 0 0 0 3px #3b82f6;
}

.point.selected {
  font-weight: 500;
}

.point.in-range .point-circle {
  background: #93c5fd !important;
  box-shadow: 0 0 0 2px #3b82f6;
}

.point.in-range {
  font-weight: 500;
  color: #1e40af;
}

.point-text {
  text-align: center;
  font-size: 12px;
  margin-top: 4px;
}
</style>
