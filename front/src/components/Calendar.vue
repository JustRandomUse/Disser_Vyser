<template>
  <div v-if="isOpen" class="calendar-backdrop" @click="close">
    <div class="calendar-content" @click.stop>
      <div class="calendar-header">
        <button @click="previousMonth" class="nav-btn">←</button>
        <span class="month-year">{{ monthYear }}</span>
        <button @click="nextMonth" class="nav-btn">→</button>
      </div>

      <div class="calendar-grid">
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

      <div class="calendar-footer">
        <button @click="selectToday" class="today-btn">Сегодня</button>
        <button @click="close" class="close-btn">Закрыть</button>
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

const currentMonth = ref(new Date().getMonth());
const currentYear = ref(new Date().getFullYear());
const weekDays = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];
const startDate = ref(null);
const endDate = ref(null);
const isSelectingRange = ref(false);

const monthYear = computed(() => {
  const months = [
    'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
    'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'
  ];
  return `${months[currentMonth.value]} ${currentYear.value}`;
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

const selectDate = (day) => {
  if (day.otherMonth) return;

  const clickedDate = new Date(day.year, day.month, day.date);

  if (!startDate.value || (startDate.value && endDate.value)) {
    startDate.value = clickedDate;
    endDate.value = null;
    isSelectingRange.value = true;
  } else if (startDate.value && !endDate.value) {
    if (clickedDate < startDate.value) {
      endDate.value = startDate.value;
      startDate.value = clickedDate;
    } else {
      endDate.value = clickedDate;
    }
    isSelectingRange.value = false;

    emit('date-range-selected', {
      start: startDate.value,
      end: endDate.value
    });
  }
};

const selectToday = () => {
  const today = new Date();
  currentMonth.value = today.getMonth();
  currentYear.value = today.getFullYear();
  emit('date-selected', today);
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
  if (!startDate.value || day.otherMonth) return false;
  return (
    day.date === startDate.value.getDate() &&
    day.month === startDate.value.getMonth() &&
    day.year === startDate.value.getFullYear()
  );
};

const isRangeEnd = (day) => {
  if (!endDate.value || day.otherMonth) return false;
  return (
    day.date === endDate.value.getDate() &&
    day.month === endDate.value.getMonth() &&
    day.year === endDate.value.getFullYear()
  );
};

const isInRange = (day) => {
  if (!startDate.value || !endDate.value || day.otherMonth) return false;
  const dayDate = new Date(day.year, day.month, day.date);
  return dayDate > startDate.value && dayDate < endDate.value;
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

.calendar-footer {
  display: flex;
  justify-content: space-between;
  margin-top: 20px;
  gap: 10px;
}

.today-btn,
.close-btn {
  flex: 1;
  padding: 10px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
}

.today-btn {
  background: #3b82f6;
  color: white;
}

.today-btn:hover {
  background: #2563eb;
}

.close-btn {
  background: #e5e7eb;
  color: #333;
}

.close-btn:hover {
  background: #d1d5db;
}
</style>
