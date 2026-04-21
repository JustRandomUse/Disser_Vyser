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
    
    <div class="timeline-body">
      <div class="line"></div>
      <div
        v-for="(point, index) in visiblePoints"
        :key="index"
        class="point"
        :class="{ selected: index === selectedIndex }"
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
import { ref, computed, watch } from 'vue';

const props = defineProps({
  date: {
    type: Date,
    default: () => new Date()
  },
  timePoints: {
    type: Array,
    default: () => []
  }
});

const emit = defineEmits(['open-calendar', 'time-selected']);

const currentPage = ref(1);
const pointsPerPage = 15;
const selectedIndex = ref(0);

// Watch for date changes and reset pagination
watch(() => props.date, () => {
  currentPage.value = 1;
  selectedIndex.value = 0;
  if (props.timePoints.length > 0) {
    emit('time-selected', props.timePoints[0]);
  }
});

const formattedDate = computed(() => {
  return props.date.toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric'
  });
});

const totalPages = computed(() => {
  return Math.ceil(props.timePoints.length / pointsPerPage);
});

const visiblePoints = computed(() => {
  const start = (currentPage.value - 1) * pointsPerPage;
  const end = start + pointsPerPage;
  return props.timePoints.slice(start, end);
});

const selectPoint = (index) => {
  selectedIndex.value = index;
  const globalIndex = (currentPage.value - 1) * pointsPerPage + index;
  emit('time-selected', props.timePoints[globalIndex]);
};

const previousPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--;
    selectedIndex.value = 0;
  }
};

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++;
    selectedIndex.value = 0;
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
  grid-template-columns: repeat(15, 1fr);
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

.point-text {
  text-align: center;
  font-size: 12px;
  margin-top: 4px;
}
</style>
