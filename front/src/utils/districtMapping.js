import booleanPointInPolygon from '@turf/boolean-point-in-polygon';
import { point } from '@turf/helpers';
import districtsGeoJSON from '../data/krasnoyarsk-districts.geojson?url';

/**
 * District mapping utility
 *
 * Data source: Approximate boundaries based on Krasnoyarsk administrative districts
 * Note: These are simplified polygons. For production use, obtain official boundaries from:
 * - OpenStreetMap (when available)
 * - Krasnoyarsk city administration
 * - Russian cadastral data (Rosreestr)
 *
 * License: Data derived from public administrative information
 */

// Load districts from GeoJSON
let districts = [];
let districtsLoaded = false;

// Async load GeoJSON
async function loadDistricts() {
  if (districtsLoaded) return districts;

  try {
    const response = await fetch(districtsGeoJSON);
    const data = await response.json();
    districts = data.features || [];
    districtsLoaded = true;
    return districts;
  } catch (error) {
    console.error('Failed to load districts GeoJSON:', error);
    districts = [];
    districtsLoaded = true;
    return districts;
  }
}

/**
 * Find which district a sensor belongs to based on coordinates
 * @param {number} lon - Longitude (WGS84)
 * @param {number} lat - Latitude (WGS84)
 * @returns {Promise<{ key: string, name: string } | null>} District info or null if not found
 */
export async function getDistrictForSensor(lon, lat) {
  if (!lon || !lat) return null;

  await loadDistricts();

  if (districts.length === 0) return null;

  const sensorPoint = point([lon, lat]);

  for (const district of districts) {
    try {
      if (booleanPointInPolygon(sensorPoint, district)) {
        return {
          key: district.properties.key,
          name: district.properties.name
        };
      }
    } catch (error) {
      console.error(`Error checking district ${district.properties.name}:`, error);
    }
  }

  return null;
}

/**
 * Group sensors by district
 * @param {Array} sensors - Array of sensor objects with latitude/longitude
 * @returns {Promise<Map<string, Array>>} Map of district key to sensors array
 */
export async function groupSensorsByDistrict(sensors) {
  await loadDistricts();

  const grouped = new Map();
  const unassigned = [];

  // Initialize with all districts
  districts.forEach(district => {
    grouped.set(district.properties.key, []);
  });

  // Add special category for unassigned sensors
  grouped.set('unassigned', []);

  for (const sensor of sensors) {
    const district = await getDistrictForSensor(sensor.longitude, sensor.latitude);

    if (district) {
      const districtSensors = grouped.get(district.key) || [];
      districtSensors.push(sensor);
      grouped.set(district.key, districtSensors);
    } else {
      unassigned.push(sensor);
      grouped.get('unassigned').push(sensor);
    }
  }

  return grouped;
}

/**
 * Get list of all available districts
 * @returns {Promise<Array<{key: string, name: string}>>}
 */
export async function getAvailableDistricts() {
  await loadDistricts();

  return districts.map(d => ({
    key: d.properties.key,
    name: d.properties.name
  }));
}

/**
 * Validate sensor-district mapping
 * Checks for sensors in multiple districts or without district
 * @param {Array} sensors - Array of sensor objects
 * @returns {Promise<Object>} Validation report
 */
export async function validateDistrictMapping(sensors) {
  await loadDistricts();

  const report = {
    total: sensors.length,
    assigned: 0,
    unassigned: 0,
    multipleDistricts: [],
    byDistrict: {}
  };

  const grouped = await groupSensorsByDistrict(sensors);

  grouped.forEach((districtSensors, districtKey) => {
    if (districtKey === 'unassigned') {
      report.unassigned = districtSensors.length;
    } else {
      report.byDistrict[districtKey] = districtSensors.length;
      report.assigned += districtSensors.length;
    }
  });

  // Check for sensors in multiple districts (should not happen with proper polygons)
  for (const sensor of sensors) {
    const matchingDistricts = [];
    const sensorPoint = point([sensor.longitude, sensor.latitude]);

    for (const district of districts) {
      try {
        if (booleanPointInPolygon(sensorPoint, district)) {
          matchingDistricts.push(district.properties.name);
        }
      } catch (error) {
        // Skip invalid geometries
      }
    }

    if (matchingDistricts.length > 1) {
      report.multipleDistricts.push({
        sensor: sensor.name,
        districts: matchingDistricts
      });
    }
  }

  return report;
}
