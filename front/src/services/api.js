import axios from 'axios';

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

        return {
          id: item.site,
          name: coords.name,
          latitude: coords.lat,
          longitude: coords.lon,
          pm25: item['p-pm2'] || 0,
          pm10: item['p-pm10'] || 0,
          temperature: item['m-t'] || 0,
          humidity: item['m-h'] || 0,
          pressure: item['m-p'] || 0,
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

        return {
          id: item.site,
          name: coords.name,
          latitude: coords.lat,
          longitude: coords.lon,
          pm25: item['p-pm2'] || 0,
          pm10: item['p-pm10'] || 0,
          temperature: item['m-t'] || 0,
          humidity: item['m-h'] || 0,
          pressure: item['m-p'] || 0,
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
        if (!siteData[siteId]) {
          siteData[siteId] = {
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
        if (item['m-t']) siteData[siteId].temperature.push(item['m-t']);
        if (item['m-h']) siteData[siteId].humidity.push(item['m-h']);
        if (item['m-p']) siteData[siteId].pressure.push(item['m-p']);
        if (item.iaqi) siteData[siteId].aqi.push(item.iaqi);
      });

      // Calculate averages
      const averages = Object.keys(siteData).map(siteId => {
        const coords = coordinates[siteId];
        if (!coords) return null;

        const data = siteData[siteId];

        const avg = (arr) => arr.length > 0 ? arr.reduce((a, b) => a + b, 0) / arr.length : 0;
        const min = (arr) => arr.length > 0 ? Math.min(...arr) : 0;
        const max = (arr) => arr.length > 0 ? Math.max(...arr) : 0;

        return {
          id: siteId,
          name: coords.name,
          latitude: coords.lat,
          longitude: coords.lon,
          pm25: Math.round(avg(data.pm25) * 10) / 10,
          pm10: Math.round(avg(data.pm10) * 10) / 10,
          temperature: Math.round(avg(data.temperature) * 10) / 10,
          humidity: Math.round(avg(data.humidity) * 10) / 10,
          pressure: Math.round(avg(data.pressure) * 10) / 10,
          aqi: Math.round(avg(data.aqi) * 10) / 10,
          pm25Min: Math.round(min(data.pm25) * 10) / 10,
          pm25Max: Math.round(max(data.pm25) * 10) / 10,
          pm10Min: Math.round(min(data.pm10) * 10) / 10,
          pm10Max: Math.round(max(data.pm10) * 10) / 10,
          temperatureMin: Math.round(min(data.temperature) * 10) / 10,
          temperatureMax: Math.round(max(data.temperature) * 10) / 10
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
        if (!siteTimeSeries[siteId]) {
          const coords = coordinates[siteId];
          siteTimeSeries[siteId] = {
            id: siteId,
            name: coords ? coords.name : `Site ${siteId}`,
            data: []
          };
        }

        siteTimeSeries[siteId].data.push({
          time: item.time,
          pm25: item['p-pm2'] || 0,
          pm10: item['p-pm10'] || 0,
          temperature: item['m-t'] || 0,
          humidity: item['m-h'] || 0,
          pressure: item['m-p'] || 0,
          aqi: item.iaqi || 0
        });
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
