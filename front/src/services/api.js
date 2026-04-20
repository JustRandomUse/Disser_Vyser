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

export const fetchAirQualityData = async () => {
  try {
    // Get coordinates first
    const coordinates = await fetchSiteCoordinates();

    // Use our backend endpoint for latest data
    const response = await axios.get(`${API_BASE_URL}/datasets/knc-air/last`);

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

export const getPollutionLevel = (pm25) => {
  if (pm25 <= 12) return { level: 'Good', color: '#00e400' };
  if (pm25 <= 35.4) return { level: 'Moderate', color: '#ffff00' };
  if (pm25 <= 55.4) return { level: 'Unhealthy for Sensitive', color: '#ff7e00' };
  if (pm25 <= 150.4) return { level: 'Unhealthy', color: '#ff0000' };
  if (pm25 <= 250.4) return { level: 'Very Unhealthy', color: '#8f3f97' };
  return { level: 'Hazardous', color: '#7e0023' };
};
