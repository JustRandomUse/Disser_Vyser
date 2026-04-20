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

<script>
export default {
  name: 'Calendar',
  props: {
    isOpen: {
      type: Boolean,
      default: false
    },
    selectedDate: {
      type: Date,
      default: () => new Date()
    }
  },
  data() {
    return {
      currentMonth: new Date().getMonth(),
      currentYear: new Date().getFullYear(),
      weekDays: ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'],
      startDate: null,
      endDate: null,
      isSelectingRange: false
    };
  },
  computed: {
    monthYear() {
      const months = [
        'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
        'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'
      ];
      return `${months[this.currentMonth]} ${this.currentYear}`;
    },
    calendarDays() {
      const days = [];
      const firstDay = new Date(this.currentYear, this.currentMonth, 1);
      const lastDay = new Date(this.currentYear, this.currentMonth + 1, 0);
      
      // Get day of week (0 = Sunday, 1 = Monday, etc.)
      let startDay = firstDay.getDay();
      startDay = startDay === 0 ? 6 : startDay - 1; // Convert to Monday = 0
      
      // Previous month days
      const prevMonthLastDay = new Date(this.currentYear, this.currentMonth, 0).getDate();
      for (let i = startDay - 1; i >= 0; i--) {
        days.push({
          date: prevMonthLastDay - i,
          month: this.currentMonth - 1,
          year: this.currentYear,
          otherMonth: true
        });
      }
      
      // Current month days
      for (let i = 1; i <= lastDay.getDate(); i++) {
        days.push({
          date: i,
          month: this.currentMonth,
          year: this.currentYear,
          otherMonth: false
        });
      }
      
      // Next month days
      const remainingDays = 42 - days.length; // 6 rows * 7 days
      for (let i = 1; i <= remainingDays; i++) {
        days.push({
          date: i,
          month: this.currentMonth + 1,
          year: this.currentYear,
          otherMonth: true
        });
      }
      
      return days;
    }
  },
  methods: {
    close() {
      this.$emit('close');
    },
    previousMonth() {
      if (this.currentMonth === 0) {
        this.currentMonth = 11;
        this.currentYear--;
      } else {
        this.currentMonth--;
      }
    },
    nextMonth() {
      if (this.currentMonth === 11) {
        this.currentMonth = 0;
        this.currentYear++;
      } else {
        this.currentMonth++;
      }
    },
    selectDate(day) {
      if (day.otherMonth) return;

      const clickedDate = new Date(day.year, day.month, day.date);

      // First click - set start date
      if (!this.startDate || (this.startDate && this.endDate)) {
        this.startDate = clickedDate;
        this.endDate = null;
        this.isSelectingRange = true;
      }
      // Second click - set end date
      else if (this.startDate && !this.endDate) {
        if (clickedDate < this.startDate) {
          // If clicked date is before start, swap them
          this.endDate = this.startDate;
          this.startDate = clickedDate;
        } else {
          this.endDate = clickedDate;
        }
        this.isSelectingRange = false;

        // Emit the range
        this.$emit('date-range-selected', {
          start: this.startDate,
          end: this.endDate
        });
      }
    },
    selectToday() {
      const today = new Date();
      this.currentMonth = today.getMonth();
      this.currentYear = today.getFullYear();
      this.$emit('date-selected', today);
      this.close();
    },
    isSelected(day) {
      if (!this.selectedDate || day.otherMonth) return false;
      return (
        day.date === this.selectedDate.getDate() &&
        day.month === this.selectedDate.getMonth() &&
        day.year === this.selectedDate.getFullYear()
      );
    },
    isToday(day) {
      const today = new Date();
      return (
        day.date === today.getDate() &&
        day.month === today.getMonth() &&
        day.year === today.getFullYear()
      );
    },
    isRangeStart(day) {
      if (!this.startDate || day.otherMonth) return false;
      return (
        day.date === this.startDate.getDate() &&
        day.month === this.startDate.getMonth() &&
        day.year === this.startDate.getFullYear()
      );
    },
    isRangeEnd(day) {
      if (!this.endDate || day.otherMonth) return false;
      return (
        day.date === this.endDate.getDate() &&
        day.month === this.endDate.getMonth() &&
        day.year === this.endDate.getFullYear()
      );
    },
    isInRange(day) {
      if (!this.startDate || !this.endDate || day.otherMonth) return false;
      const dayDate = new Date(day.year, day.month, day.date);
      return dayDate > this.startDate && dayDate < this.endDate;
    }
  }
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
