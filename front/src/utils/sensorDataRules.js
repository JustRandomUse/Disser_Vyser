/**
 * Configuration for unavailable/unreliable sensor parameters
 *
 * Some sensors report incorrect values for certain parameters.
 * These values should be excluded from:
 * - UI display (show "—" instead)
 * - Statistics calculations (avg, min, max)
 * - Charts and comparisons
 */

// Normalize sensor name for reliable comparison
// Handles variations in punctuation, spacing, and case
export const normalizeSensorName = (name) => {
  if (!name) return '';

  return name
    .toLowerCase()
    .trim()
    .replace(/\s+/g, ' ')           // normalize multiple spaces
    .replace(/,\s*/g, ' ')          // remove commas
    .replace(/\./g, '')             // remove dots
    .replace(/\s+/g, '');           // remove all spaces for final comparison
};

// Configuration: unavailable parameters by sensor
// Key: normalized sensor name
// Value: array of unavailable parameters
const UNAVAILABLE_PARAMS = {
  // Temperature unavailable
  'кировский': ['temperature'],
  'омолокова': ['temperature', 'humidity', 'pressure'],
  'отатышев': ['temperature'],
  'партизана3г': ['temperature', 'pressure'],
  'ленина41': ['temperature'],
  'качинская56а': ['temperature'],
  'николаевка': ['temperature', 'humidity', 'pressure'],
  'телевизорная131': ['temperature', 'humidity', 'pressure'],
  'солонцы': ['pressure']
};

/**
 * Check if a parameter is unavailable for a given sensor
 * @param {string} sensorName - Sensor name (will be normalized)
 * @param {string} param - Parameter name (temperature, humidity, pressure, etc.)
 * @returns {boolean} - true if parameter is unavailable for this sensor
 */
export const isParamUnavailableForSensor = (sensorName, param) => {
  if (!sensorName || !param) return false;

  const normalized = normalizeSensorName(sensorName);
  const unavailableParams = UNAVAILABLE_PARAMS[normalized];

  return unavailableParams ? unavailableParams.includes(param) : false;
};

/**
 * Check if a value is valid for metrics calculations
 * @param {*} value - Value to check
 * @returns {boolean} - true if value is a valid number (not null, not NaN, not undefined)
 */
export const isValidMetricValue = (value) => {
  return value !== null && value !== undefined && typeof value === 'number' && !isNaN(value);
};

/**
 * Get display value for UI
 * Returns null if parameter is unavailable for sensor, otherwise returns the value
 * @param {string} sensorName - Sensor name
 * @param {string} param - Parameter name
 * @param {*} value - Raw value from API
 * @returns {number|null} - Value or null if unavailable
 */
export const getValidValue = (sensorName, param, value) => {
  // If parameter is unavailable for this sensor, return null
  if (isParamUnavailableForSensor(sensorName, param)) {
    return null;
  }

  // If value is not valid, return null
  if (!isValidMetricValue(value)) {
    return null;
  }

  return value;
};

/**
 * Format value for display in UI
 * @param {*} value - Value to format
 * @param {number} decimals - Number of decimal places (default 1)
 * @returns {string} - Formatted value or "—" for unavailable
 */
export const formatDisplayValue = (value, decimals = 1) => {
  if (!isValidMetricValue(value)) {
    return '—';
  }

  return (Math.round(value * Math.pow(10, decimals)) / Math.pow(10, decimals)).toString();
};
