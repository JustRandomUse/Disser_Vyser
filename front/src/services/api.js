import axios from 'axios';
import { getValidValue } from '../utils/sensorDataRules';

// Use local backend proxy instead of direct API calls
const API_BASE_URL = '/api';

// Cache for site coordinates from dataset metadata
let siteCoordinatesCache = null;

// Fetch site coordinates from dataset metadata
const fetchSiteCoordinates = async () => {
  if (siteCoordinatesCache) {
    return siteCoordinatesCache;
  }

  try {
    const response = await axios.get(`${API_BASE_URL}/datasets/knc-air`);

    if (response.data && response.data.data && response.data.data.sites) {
      const coordinates = {};
      response.data.data.sites.forEach(site => {
        coordinates[site.id] = {
          lat: site.geom_y,
          lon: site.geom_x,
          name: site.name,
          code: site.code
        };
      });
      siteCoordinatesCache = coordinates;
      return coordinates;
    }
  } catch (error) {
    console.error('Error fetching site coordinates:', error);
  }

  return {};
};

export const fetchAirQualityData = async (date = null, hour = null) => {
  try {
    // Get coordinates first
    const coordinates = await fetchSiteCoordinates();

    let endpoint = `${API_BASE_URL}/datasets/knc-air/last`;

    // If date and hour are provided, fetch historical data
    if (date && hour !== null) {
      const dateStr = date.toISOString().split('T')[0];
      const hourStr = hour.toString().padStart(2, '0');
      endpoint = `${API_BASE_URL}/datasets/knc-air/data?date=${dateStr}&hour=${hourStr}`;
    }

    // Use our backend endpoint for data
    const response = await axios.get(endpoint);

    if (response.data && response.data.status && response.data.status === 'success') {
      const sensors = response.data.data.map(item => {
        const coords = coordinates[item.site];
        if (!coords) return null;

        const sensorName = coords.name;

        return {
          id: item.site,
          name: sensorName,
          latitude: coords.lat,
          longitude: coords.lon,
          pm25: item['p-pm2'] || 0,
          pm10: item['p-pm10'] || 0,
          temperature: getValidValue(sensorName, 'temperature', item['m-t']),
          humidity: getValidValue(sensorName, 'humidity', item['m-h']),
          pressure: getValidValue(sensorName, 'pressure', item['m-p']),
          aqi: item.iaqi || 0,
          time: item.time
        };
      }).filter(item => item !== null);

      return sensors;
    }

    return [];
  } catch (error) {
    console.error('Error fetching air quality data:', error);
    throw error;
  }
};

export const fetchAggregatedData = async (startDate, endDate, interval = 'hour') => {
  try {
    const coordinates = await fetchSiteCoordinates();

    const timeBegin = startDate.toISOString();
    const timeEnd = endDate.toISOString();

    const endpoint = `${API_BASE_URL}/datasets/knc-air/aggregated?time_begin=${timeBegin}&time_end=${timeEnd}&interval=${interval}`;

    const response = await axios.get(endpoint);

    if (response.data && response.data.status === 'success') {
      const sensors = response.data.data.map(item => {
        const coords = coordinates[item.site];
        if (!coords) return null;

        const sensorName = coords.name;

        return {
          id: item.site,
          name: sensorName,
          latitude: coords.lat,
          longitude: coords.lon,
          pm25: item['p-pm2'] || 0,
          pm10: item['p-pm10'] || 0,
          temperature: getValidValue(sensorName, 'temperature', item['m-t']),
          humidity: getValidValue(sensorName, 'humidity', item['m-h']),
          pressure: getValidValue(sensorName, 'pressure', item['m-p']),
          aqi: item.iaqi || 0,
          time: item.time
        };
      }).filter(item => item !== null);

      return sensors;
    }

    return [];
  } catch (error) {
    console.error('Error fetching aggregated data:', error);
    throw error;
  }
};

/**
 * Fetch average values for a date range
 * Returns averaged data across the entire period
 */
export const fetchAverageData = async (startDate, endDate, interval = 'hour', sites = null, indicators = null) => {
  try {
    const coordinates = await fetchSiteCoordinates();

    const timeBegin = startDate.toISOString();
    const timeEnd = endDate.toISOString();

    let endpoint = `${API_BASE_URL}/datasets/knc-air/aggregated?time_begin=${timeBegin}&time_end=${timeEnd}&interval=${interval}`;

    if (sites && sites.length > 0) {
      endpoint += `&sites=${sites.join(',')}`;
    }

    if (indicators && indicators.length > 0) {
      endpoint += `&indicators=${indicators.join(',')}`;
    }

    const response = await axios.get(endpoint);

    if (response.data && response.data.status === 'success') {
      // Group by site and calculate averages
      const siteData = {};

      response.data.data.forEach(item => {
        const siteId = item.site;
        const coords = coordinates[siteId];
        const sensorName = coords ? coords.name : '';

        if (!siteData[siteId]) {
          siteData[siteId] = {
            name: sensorName,
            pm25: [],
            pm10: [],
            temperature: [],
            humidity: [],
            pressure: [],
            aqi: []
          };
        }

        if (item['p-pm2']) siteData[siteId].pm25.push(item['p-pm2']);
        if (item['p-pm10']) siteData[siteId].pm10.push(item['p-pm10']);

        const temp = getValidValue(sensorName, 'temperature', item['m-t']);
        const hum = getValidValue(sensorName, 'humidity', item['m-h']);
        const press = getValidValue(sensorName, 'pressure', item['m-p']);

        if (temp !== null) siteData[siteId].temperature.push(temp);
        if (hum !== null) siteData[siteId].humidity.push(hum);
        if (press !== null) siteData[siteId].pressure.push(press);
        if (item.iaqi) siteData[siteId].aqi.push(item.iaqi);
      });

      // Calculate averages
      const averages = Object.keys(siteData).map(siteId => {
        const coords = coordinates[siteId];
        if (!coords) return null;

        const data = siteData[siteId];
        const sensorName = data.name;

        const avg = (arr) => arr.length > 0 ? arr.reduce((a, b) => a + b, 0) / arr.length : null;
        const min = (arr) => arr.length > 0 ? Math.min(...arr) : null;
        const max = (arr) => arr.length > 0 ? Math.max(...arr) : null;

        const round = (val) => val !== null ? Math.round(val * 10) / 10 : null;

        return {
          id: siteId,
          name: sensorName,
          latitude: coords.lat,
          longitude: coords.lon,
          pm25: round(avg(data.pm25)),
          pm10: round(avg(data.pm10)),
          temperature: round(avg(data.temperature)),
          humidity: round(avg(data.humidity)),
          pressure: round(avg(data.pressure)),
          aqi: round(avg(data.aqi)),
          pm25Min: round(min(data.pm25)),
          pm25Max: round(max(data.pm25)),
          pm10Min: round(min(data.pm10)),
          pm10Max: round(max(data.pm10)),
          temperatureMin: round(min(data.temperature)),
          temperatureMax: round(max(data.temperature))
        };
      }).filter(item => item !== null);

      return averages;
    }

    return [];
  } catch (error) {
    console.error('Error fetching average data:', error);
    throw error;
  }
};

/**
 * Fetch time series data for charts
 * Returns data points for each interval in the range
 */
export const fetchTimeSeriesData = async (startDate, endDate, interval = 'hour', sites = null, indicators = null) => {
  try {
    const coordinates = await fetchSiteCoordinates();

    const timeBegin = startDate.toISOString();
    const timeEnd = endDate.toISOString();

    let endpoint = `${API_BASE_URL}/datasets/knc-air/aggregated?time_begin=${timeBegin}&time_end=${timeEnd}&interval=${interval}`;

    if (sites && sites.length > 0) {
      endpoint += `&sites=${sites.join(',')}`;
    }

    if (indicators && indicators.length > 0) {
      endpoint += `&indicators=${indicators.join(',')}`;
    }

    console.log('🌐 API fetchTimeSeriesData request:', endpoint);

    const response = await axios.get(endpoint);

    console.log('🌐 API response.data.data length:', response.data?.data?.length);

    if (response.data && response.data.status === 'success') {
      // Group by site
      const siteTimeSeries = {};

      response.data.data.forEach(item => {
        const siteId = item.site;
        const coords = coordinates[siteId];
        const sensorName = coords ? coords.name : '';

        if (!siteTimeSeries[siteId]) {
          siteTimeSeries[siteId] = {
            id: siteId,
            name: sensorName,
            data: []
          };
        }

        siteTimeSeries[siteId].data.push({
          time: item.time,
          pm25: item['p-pm2'] !== undefined && item['p-pm2'] !== null ? item['p-pm2'] : null,
          pm10: item['p-pm10'] !== undefined && item['p-pm10'] !== null ? item['p-pm10'] : null,
          temperature: getValidValue(sensorName, 'temperature', item['m-t']),
          humidity: getValidValue(sensorName, 'humidity', item['m-h']),
          pressure: getValidValue(sensorName, 'pressure', item['m-p']),
          aqi: item.iaqi !== undefined && item.iaqi !== null ? item.iaqi : null
        });
      });

      // Sort data by time for each site
      Object.values(siteTimeSeries).forEach(site => {
        site.data.sort((a, b) => new Date(a.time) - new Date(b.time));
      });

      return Object.values(siteTimeSeries);
    }

    return [];
  } catch (error) {
    console.error('Error fetching time series data:', error);
    throw error;
  }
};

export const getPollutionLevel = (pm25) => {
  if (pm25 <= 12) return { level: 'Good', color: '#00e400' };
  if (pm25 <= 35.4) return { level: 'Moderate', color: '#ffff00' };
  if (pm25 <= 55.4) return { level: 'Unhealthy for Sensitive', color: '#ff7e00' };
  if (pm25 <= 150.4) return { level: 'Unhealthy', color: '#ff0000' };
  if (pm25 <= 250.4) return { level: 'Very Unhealthy', color: '#8f3f97' };
  return { level: 'Hazardous', color: '#7e0023' };
};
