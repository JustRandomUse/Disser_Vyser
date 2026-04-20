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

<script>
export default {
  name: 'Timeline',
  props: {
    date: {
      type: Date,
      default: () => new Date()
    },
    timePoints: {
      type: Array,
      default: () => []
    }
  },
  data() {
    return {
      currentPage: 1,
      pointsPerPage: 15,
      selectedIndex: 0
    };
  },
  computed: {
    formattedDate() {
      return this.date.toLocaleDateString('ru-RU', {
        day: 'numeric',
        month: 'short',
        year: 'numeric'
      });
    },
    totalPages() {
      return Math.ceil(this.timePoints.length / this.pointsPerPage);
    },
    visiblePoints() {
      const start = (this.currentPage - 1) * this.pointsPerPage;
      const end = start + this.pointsPerPage;
      return this.timePoints.slice(start, end);
    }
  },
  methods: {
    selectPoint(index) {
      this.selectedIndex = index;
      const globalIndex = (this.currentPage - 1) * this.pointsPerPage + index;
      this.$emit('time-selected', this.timePoints[globalIndex]);
    },
    previousPage() {
      if (this.currentPage > 1) {
        this.currentPage--;
        this.selectedIndex = 0;
      }
    },
    nextPage() {
      if (this.currentPage < this.totalPages) {
        this.currentPage++;
        this.selectedIndex = 0;
      }
    }
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
