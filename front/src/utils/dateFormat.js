/**
 * Date formatting utilities for calendar and timeline
 */

/**
 * Format date as day: "01.04.2026"
 */
export function formatDay(date) {
  const day = date.getDate().toString().padStart(2, '0');
  const month = (date.getMonth() + 1).toString().padStart(2, '0');
  const year = date.getFullYear();
  return `${day}.${month}.${year}`;
}

/**
 * Format date as short day: "01.04"
 */
export function formatDayShort(date) {
  const day = date.getDate().toString().padStart(2, '0');
  const month = (date.getMonth() + 1).toString().padStart(2, '0');
  return `${day}.${month}`;
}

/**
 * Format date as month: "Апр 2026"
 */
export function formatMonth(date) {
  const months = ['Янв', 'Фев', 'Мар', 'Апр', 'Май', 'Июн', 'Июл', 'Авг', 'Сен', 'Окт', 'Ноя', 'Дек'];
  return `${months[date.getMonth()]} ${date.getFullYear()}`;
}

/**
 * Format date as year: "2026"
 */
export function formatYear(date) {
  return date.getFullYear().toString();
}

/**
 * Format date range: "01.04.2026 — 05.04.2026"
 */
export function formatDateRange(start, end) {
  if (!start || !end) return '';

  const startStr = formatDay(start);
  const endStr = formatDay(end);

  if (startStr === endStr) {
    return startStr;
  }

  return `${startStr} — ${endStr}`;
}

/**
 * Get start of day
 */
export function startOfDay(date) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate(), 0, 0, 0, 0);
}

/**
 * Get end of day
 */
export function endOfDay(date) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate(), 23, 59, 59, 999);
}

/**
 * Get start of month
 */
export function startOfMonth(date) {
  return new Date(date.getFullYear(), date.getMonth(), 1, 0, 0, 0, 0);
}

/**
 * Get end of month
 */
export function endOfMonth(date) {
  return new Date(date.getFullYear(), date.getMonth() + 1, 0, 23, 59, 59, 999);
}

/**
 * Get start of year
 */
export function startOfYear(date) {
  return new Date(date.getFullYear(), 0, 1, 0, 0, 0, 0);
}

/**
 * Get end of year
 */
export function endOfYear(date) {
  return new Date(date.getFullYear(), 11, 31, 23, 59, 59, 999);
}
