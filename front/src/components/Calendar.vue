<template>
  <div v-if="isOpen" class="calendar-backdrop" @click="close">
    <div class="calendar-content" @click.stop>
      <div class="mode-selector">
        <button
          @click="selectionMode = 'day'"
          :class="{ active: selectionMode === 'day' }"
          class="mode-btn"
        >
          Дни
        </button>
        <button
          @click="selectionMode = 'month'"
          :class="{ active: selectionMode === 'month' }"
          class="mode-btn"
        >
          Месяцы
        </button>
        <button
          @click="selectionMode = 'year'"
          :class="{ active: selectionMode === 'year' }"
          class="mode-btn"
        >
          Годы
        </button>
      </div>

      <div class="calendar-header" v-if="selectionMode !== 'year'">
        <button @click="previousMonth" class="nav-btn" v-if="selectionMode === 'day'">←</button>
        <button @click="previousYear" class="nav-btn" v-if="selectionMode === 'month'">←</button>
        <span class="month-year">{{ headerText }}</span>
        <button @click="nextMonth" class="nav-btn" v-if="selectionMode === 'day'">→</button>
        <button @click="nextYear" class="nav-btn" v-if="selectionMode === 'month'">→</button>
      </div>

      <!-- Day mode -->
      <div v-if="selectionMode === 'day'" class="calendar-grid">
        <div v-for="day in weekDays" :key="day" class="week-day">{{ day }}</div>

        <div
          v-for="(day, index) in calendarDays"
          :key="index"
          class="calendar-day"
          :class="{
            'other-month': day.otherMonth,
            'selected': isSelected(day),
            'in-range': isInRange(day),
            'range-start': isRangeStart(day),
            'range-end': isRangeEnd(day),
            'today': isToday(day)
          }"
          @click="selectDate(day)"
        >
          {{ day.date }}
        </div>
      </div>

      <!-- Month mode -->
      <div v-if="selectionMode === 'month'" class="months-grid">
        <div
          v-for="(month, index) in monthsGrid"
          :key="index"
          class="month-item"
          :class="{
            'selected': isMonthSelected(month),
            'in-range': isMonthInRange(month),
            'range-start': isMonthRangeStart(month),
            'range-end': isMonthRangeEnd(month)
          }"
          @click="selectMonth(month)"
        >
          {{ month.name }}
        </div>
      </div>

      <!-- Year mode -->
      <div v-if="selectionMode === 'year'" class="years-grid">
        <div
          v-for="year in yearsGrid"
          :key="year"
          class="year-item"
          :class="{
            'selected': isYearSelected(year),
            'in-range': isYearInRange(year),
            'range-start': isYearRangeStart(year),
            'range-end': isYearRangeEnd(year)
          }"
          @click="selectYear(year)"
        >
          {{ year }}
        </div>
      </div>

      <div class="calendar-footer">
        <button @click="selectToday" class="today-btn">Сегодня</button>
        <button @click="cancelSelection" class="cancel-btn">Отмена</button>
        <button @click="applySelection" class="apply-btn" :disabled="!pendingStart">Применить</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';

const props = defineProps({
  isOpen: {
    type: Boolean,
    default: false
  },
  selectedDate: {
    type: Date,
    default: () => new Date()
  }
});

const emit = defineEmits(['close', 'date-selected', 'date-range-selected']);

const selectionMode = ref('day'); // 'day' | 'month' | 'year'
const currentMonth = ref(new Date().getMonth());
const currentYear = ref(new Date().getFullYear());
const weekDays = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];
const startDate = ref(null);
const endDate = ref(null);
const isSelectingRange = ref(false);
const pendingStart = ref(null);
const pendingEnd = ref(null);

const monthYear = computed(() => {
  const months = [
    'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
    'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'
  ];
  return `${months[currentMonth.value]} ${currentYear.value}`;
});

const headerText = computed(() => {
  if (selectionMode.value === 'day') {
    const months = [
      'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
      'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'
    ];
    return `${months[currentMonth.value]} ${currentYear.value}`;
  } else if (selectionMode.value === 'month') {
    return `${currentYear.value}`;
  }
  return '';
});

const monthsGrid = computed(() => {
  const months = [
    'Янв', 'Фев', 'Мар', 'Апр', 'Май', 'Июн',
    'Июл', 'Авг', 'Сен', 'Окт', 'Ноя', 'Дек'
  ];
  return months.map((name, index) => ({
    name,
    month: index,
    year: currentYear.value
  }));
});

const yearsGrid = computed(() => {
  const startYear = Math.floor(currentYear.value / 12) * 12;
  const years = [];
  for (let i = 0; i < 12; i++) {
    years.push(startYear + i);
  }
  return years;
});

const calendarDays = computed(() => {
  const days = [];
  const firstDay = new Date(currentYear.value, currentMonth.value, 1);
  const lastDay = new Date(currentYear.value, currentMonth.value + 1, 0);

  let startDay = firstDay.getDay();
  startDay = startDay === 0 ? 6 : startDay - 1;

  const prevMonthLastDay = new Date(currentYear.value, currentMonth.value, 0).getDate();
  for (let i = startDay - 1; i >= 0; i--) {
    days.push({
      date: prevMonthLastDay - i,
      month: currentMonth.value - 1,
      year: currentYear.value,
      otherMonth: true
    });
  }

  for (let i = 1; i <= lastDay.getDate(); i++) {
    days.push({
      date: i,
      month: currentMonth.value,
      year: currentYear.value,
      otherMonth: false
    });
  }

  const remainingDays = 42 - days.length;
  for (let i = 1; i <= remainingDays; i++) {
    days.push({
      date: i,
      month: currentMonth.value + 1,
      year: currentYear.value,
      otherMonth: true
    });
  }

  return days;
});

const close = () => {
  emit('close');
};

const previousMonth = () => {
  if (currentMonth.value === 0) {
    currentMonth.value = 11;
    currentYear.value--;
  } else {
    currentMonth.value--;
  }
};

const nextMonth = () => {
  if (currentMonth.value === 11) {
    currentMonth.value = 0;
    currentYear.value++;
  } else {
    currentMonth.value++;
  }
};

const previousYear = () => {
  currentYear.value--;
};

const nextYear = () => {
  currentYear.value++;
};

const selectDate = (day) => {
  if (day.otherMonth) return;

  const clickedDate = new Date(day.year, day.month, day.date);

  if (!pendingStart.value || (pendingStart.value && pendingEnd.value)) {
    pendingStart.value = clickedDate;
    pendingEnd.value = null;
    isSelectingRange.value = true;
  } else if (pendingStart.value && !pendingEnd.value) {
    if (clickedDate < pendingStart.value) {
      pendingEnd.value = pendingStart.value;
      pendingStart.value = clickedDate;
    } else {
      pendingEnd.value = clickedDate;
    }
    isSelectingRange.value = false;
  }
};

const selectMonth = (monthData) => {
  const startOfMonth = new Date(monthData.year, monthData.month, 1);
  const endOfMonth = new Date(monthData.year, monthData.month + 1, 0);

  if (!pendingStart.value || (pendingStart.value && pendingEnd.value)) {
    // First click - set start only
    pendingStart.value = startOfMonth;
    pendingEnd.value = null;
    isSelectingRange.value = true;
  } else if (pendingStart.value && !pendingEnd.value) {
    // Second click - set end
    if (startOfMonth < pendingStart.value) {
      // Clicked earlier month - swap
      const tempEnd = new Date(pendingStart.value.getFullYear(), pendingStart.value.getMonth() + 1, 0);
      pendingStart.value = startOfMonth;
      pendingEnd.value = tempEnd;
    } else {
      // Clicked later month - set as end
      pendingEnd.value = endOfMonth;
    }
    isSelectingRange.value = false;
  }
};

const selectYear = (year) => {
  const startOfYear = new Date(year, 0, 1);
  const endOfYear = new Date(year, 11, 31);

  if (!pendingStart.value || (pendingStart.value && pendingEnd.value)) {
    // First click - set start only
    pendingStart.value = startOfYear;
    pendingEnd.value = null;
    isSelectingRange.value = true;
  } else if (pendingStart.value && !pendingEnd.value) {
    // Second click - set end
    if (startOfYear < pendingStart.value) {
      // Clicked earlier year - swap
      const tempEnd = new Date(pendingStart.value.getFullYear(), 11, 31);
      pendingStart.value = startOfYear;
      pendingEnd.value = tempEnd;
    } else {
      // Clicked later year - set as end
      pendingEnd.value = endOfYear;
    }
    isSelectingRange.value = false;
  }
};

const selectToday = () => {
  const today = new Date();
  currentMonth.value = today.getMonth();
  currentYear.value = today.getFullYear();

  // Immediately emit and close - no need for "Apply" button
  emit('date-selected', today);
  close();
};

const applySelection = () => {
  if (!pendingStart.value) return;

  if (pendingEnd.value) {
    // Range selected
    startDate.value = pendingStart.value;
    endDate.value = pendingEnd.value;
    emit('date-range-selected', {
      start: startDate.value,
      end: endDate.value
    });
  } else {
    // Single date selected
    emit('date-selected', pendingStart.value);
  }
  close();
};

const cancelSelection = () => {
  pendingStart.value = null;
  pendingEnd.value = null;
  isSelectingRange.value = false;
  close();
};

const isSelected = (day) => {
  if (!props.selectedDate || day.otherMonth) return false;
  return (
    day.date === props.selectedDate.getDate() &&
    day.month === props.selectedDate.getMonth() &&
    day.year === props.selectedDate.getFullYear()
  );
};

const isToday = (day) => {
  const today = new Date();
  return (
    day.date === today.getDate() &&
    day.month === today.getMonth() &&
    day.year === today.getFullYear()
  );
};

const isRangeStart = (day) => {
  if (!pendingStart.value || day.otherMonth) return false;
  return (
    day.date === pendingStart.value.getDate() &&
    day.month === pendingStart.value.getMonth() &&
    day.year === pendingStart.value.getFullYear()
  );
};

const isRangeEnd = (day) => {
  if (!pendingEnd.value || day.otherMonth) return false;
  return (
    day.date === pendingEnd.value.getDate() &&
    day.month === pendingEnd.value.getMonth() &&
    day.year === pendingEnd.value.getFullYear()
  );
};

const isInRange = (day) => {
  if (!pendingStart.value || !pendingEnd.value || day.otherMonth) return false;
  const dayDate = new Date(day.year, day.month, day.date);
  return dayDate > pendingStart.value && dayDate < pendingEnd.value;
};

const isMonthSelected = (month) => {
  if (!pendingStart.value) return false;
  const monthStart = new Date(month.year, month.month, 1);
  const monthEnd = new Date(month.year, month.month + 1, 0);

  return (
    (pendingStart.value >= monthStart && pendingStart.value <= monthEnd) ||
    (pendingEnd.value && pendingEnd.value >= monthStart && pendingEnd.value <= monthEnd)
  );
};

const isMonthInRange = (month) => {
  if (!pendingStart.value || !pendingEnd.value) return false;
  const monthStart = new Date(month.year, month.month, 1);
  return monthStart > pendingStart.value && monthStart < pendingEnd.value;
};

const isYearSelected = (year) => {
  if (!pendingStart.value) return false;
  return (
    pendingStart.value.getFullYear() === year ||
    (pendingEnd.value && pendingEnd.value.getFullYear() === year)
  );
};

const isYearInRange = (year) => {
  if (!pendingStart.value || !pendingEnd.value) return false;
  return year > pendingStart.value.getFullYear() && year < pendingEnd.value.getFullYear();
};

const isMonthRangeStart = (month) => {
  if (!pendingStart.value) return false;
  const monthStart = new Date(month.year, month.month, 1);
  return (
    monthStart.getMonth() === pendingStart.value.getMonth() &&
    monthStart.getFullYear() === pendingStart.value.getFullYear()
  );
};

const isMonthRangeEnd = (month) => {
  if (!pendingEnd.value) return false;
  const monthEnd = new Date(month.year, month.month + 1, 0);
  return (
    monthEnd.getMonth() === pendingEnd.value.getMonth() &&
    monthEnd.getFullYear() === pendingEnd.value.getFullYear()
  );
};

const isYearRangeStart = (year) => {
  if (!pendingStart.value) return false;
  return year === pendingStart.value.getFullYear();
};

const isYearRangeEnd = (year) => {
  if (!pendingEnd.value) return false;
  return year === pendingEnd.value.getFullYear();
};
</script>

<style scoped>
.calendar-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 3000;
}

.calendar-content {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  min-width: 320px;
}

.mode-selector {
  display: flex;
  gap: 5px;
  margin-bottom: 15px;
  background: #f3f4f6;
  padding: 4px;
  border-radius: 8px;
}

.mode-btn {
  flex: 1;
  padding: 8px;
  border: none;
  border-radius: 6px;
  background: transparent;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  color: #666;
  transition: all 0.2s;
}

.mode-btn.active {
  background: white;
  color: #3b82f6;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.mode-btn:hover:not(.active) {
  color: #333;
}

.calendar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.month-year {
  font-size: 18px;
  font-weight: 600;
  color: #333;
}

.nav-btn {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  padding: 5px 10px;
  color: #3b82f6;
}

.nav-btn:hover {
  background: #f0f0f0;
  border-radius: 4px;
}

.calendar-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 5px;
}

.week-day {
  text-align: center;
  font-weight: 600;
  color: #666;
  padding: 8px;
  font-size: 12px;
}

.calendar-day {
  text-align: center;
  padding: 10px;
  cursor: pointer;
  border-radius: 6px;
  font-size: 14px;
  transition: background 0.2s;
}

.calendar-day:hover:not(.other-month) {
  background: #e5e7eb;
}

.calendar-day.other-month {
  color: #ccc;
  cursor: default;
}

.calendar-day.selected {
  background: #3b82f6;
  color: white;
  font-weight: 600;
}

.calendar-day.today {
  border: 2px solid #3b82f6;
}

.calendar-day.range-start,
.calendar-day.range-end {
  background: #3b82f6;
  color: white;
  font-weight: 600;
}

.calendar-day.in-range {
  background: #dbeafe;
  color: #1e40af;
}

.months-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  margin-top: 10px;
}

.month-item {
  text-align: center;
  padding: 15px;
  cursor: pointer;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  transition: background 0.2s;
  border: 2px solid transparent;
}

.month-item:hover {
  background: #e5e7eb;
}

.month-item.selected {
  background: #3b82f6;
  color: white;
}

.month-item.range-start,
.month-item.range-end {
  background: #3b82f6;
  color: white;
  font-weight: 600;
}

.month-item.in-range {
  background: #dbeafe;
  color: #1e40af;
}

.years-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  margin-top: 10px;
}

.year-item {
  text-align: center;
  padding: 15px;
  cursor: pointer;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  transition: background 0.2s;
  border: 2px solid transparent;
}

.year-item:hover {
  background: #e5e7eb;
}

.year-item.selected {
  background: #3b82f6;
  color: white;
}

.year-item.range-start,
.year-item.range-end {
  background: #3b82f6;
  color: white;
  font-weight: 600;
}

.year-item.in-range {
  background: #dbeafe;
  color: #1e40af;
}

.calendar-footer {
  display: flex;
  justify-content: space-between;
  margin-top: 20px;
  gap: 10px;
}

.today-btn,
.cancel-btn,
.apply-btn {
  padding: 10px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
}

.today-btn {
  background: #e5e7eb;
  color: #333;
  flex: 1;
}

.today-btn:hover {
  background: #d1d5db;
}

.cancel-btn {
  background: #e5e7eb;
  color: #333;
  flex: 1;
}

.cancel-btn:hover {
  background: #d1d5db;
}

.apply-btn {
  background: #3b82f6;
  color: white;
  flex: 1;
}

.apply-btn:hover:not(:disabled) {
  background: #2563eb;
}

.apply-btn:disabled {
  background: #9ca3af;
  cursor: not-allowed;
}
</style>
