<template>
  <div class="map-container">
    <div ref="mapElement" class="map"></div>
  </div>
</template>

<script>
import 'ol/ol.css';
import Map from 'ol/Map';
import View from 'ol/View';
import TileLayer from 'ol/layer/Tile';
import OSM from 'ol/source/OSM';
import VectorLayer from 'ol/layer/Vector';
import VectorSource from 'ol/source/Vector';
import Feature from 'ol/Feature';
import Point from 'ol/geom/Point';
import { Circle as CircleStyle, Fill, Stroke, Style, Text } from 'ol/style';
import { fromLonLat } from 'ol/proj';
import { getPollutionLevel } from '../services/api';

export default {
  name: 'MapView',
  props: {
    sensors: {
      type: Array,
      default: () => []
    },
    selectedIndicator: {
      type: Number,
      default: 1
    }
  },
  data() {
    return {
      map: null,
      vectorSource: null,
      vectorLayer: null
    };
  },
  mounted() {
    this.initMap();
  },
  watch: {
    sensors: {
      handler(newSensors) {
        this.updateMarkers(newSensors);
      },
      deep: true
    },
    selectedIndicator() {
      this.updateMarkers(this.sensors);
    }
  },
  methods: {
    initMap() {
      this.vectorSource = new VectorSource();
      this.vectorLayer = new VectorLayer({
        source: this.vectorSource
      });

      this.map = new Map({
        target: this.$refs.mapElement,
        layers: [
          new TileLayer({
            source: new OSM()
          }),
          this.vectorLayer
        ],
        view: new View({
          center: fromLonLat([92.8672, 56.0153]),
          zoom: 11
        })
      });

      this.map.on('click', (evt) => {
        this.map.forEachFeatureAtPixel(evt.pixel, (feature) => {
          const sensorData = feature.get('sensorData');
          if (sensorData) {
            this.$emit('sensor-click', sensorData);
          }
        });
      });
    },
    updateMarkers(sensors) {
      this.vectorSource.clear();

      const paramKeys = ['aqi', 'pm25', 'pm10', 'temperature', 'humidity', 'pressure'];
      const selectedParam = paramKeys[this.selectedIndicator] || 'pm25';

      sensors.forEach(sensor => {
        if (sensor.latitude && sensor.longitude) {
          const feature = new Feature({
            geometry: new Point(fromLonLat([sensor.longitude, sensor.latitude]))
          });

          let value = sensor[selectedParam] || 0;
          let displayValue = Math.round(value * 10) / 10;

          // Calculate AQI if selected
          if (selectedParam === 'aqi') {
            const pm25 = sensor.pm25 || 0;
            if (pm25 <= 12) value = pm25 * 50 / 12;
            else if (pm25 <= 35.4) value = 50 + (pm25 - 12) * 50 / 23.4;
            else if (pm25 <= 55.4) value = 100 + (pm25 - 35.4) * 50 / 20;
            else if (pm25 <= 150.4) value = 150 + (pm25 - 55.4) * 100 / 95;
            else if (pm25 <= 250.4) value = 200 + (pm25 - 150.4) * 100 / 100;
            else value = 300 + (pm25 - 250.4) * 200 / 249.6;
            displayValue = Math.round(value);
          }

          const color = this.getColorForValue(selectedParam, value);

          feature.setStyle(new Style({
            image: new CircleStyle({
              radius: 20,
              fill: new Fill({
                color: color
              }),
              stroke: new Stroke({
                color: '#fff',
                width: 2
              })
            }),
            text: new Text({
              text: String(displayValue),
              font: 'bold 13px Arial',
              fill: new Fill({
                color: '#fff'
              }),
              stroke: new Stroke({
                color: '#000',
                width: 3
              })
            })
          }));

          feature.set('sensorData', sensor);
          this.vectorSource.addFeature(feature);
        }
      });
    },
    getColorForValue(param, value) {
      const colorMaps = {
        aqi: [
          { max: 50, color: '#00e400' },
          { max: 100, color: '#ffff00' },
          { max: 150, color: '#ff7e00' },
          { max: 200, color: '#ff0000' },
          { max: 300, color: '#8f3f97' },
          { max: Infinity, color: '#7e0023' }
        ],
        pm25: [
          { max: 12, color: '#00e400' },
          { max: 35.4, color: '#ffff00' },
          { max: 55.4, color: '#ff7e00' },
          { max: 150, color: '#ff0000' },
          { max: 250, color: '#8f3f97' },
          { max: Infinity, color: '#7e0023' }
        ],
        pm10: [
          { max: 54, color: '#00e400' },
          { max: 154, color: '#ffff00' },
          { max: 254, color: '#ff7e00' },
          { max: 354, color: '#ff0000' },
          { max: 424, color: '#8f3f97' },
          { max: Infinity, color: '#7e0023' }
        ],
        temperature: [
          { max: -20, color: '#2E86C1' },
          { max: -10, color: '#3498DB' },
          { max: 0, color: '#5DADE2' },
          { max: 10, color: '#85C1E9' },
          { max: 20, color: '#D5DBDB' },
          { max: 30, color: '#F7DC6F' },
          { max: 40, color: '#F0B27A' },
          { max: Infinity, color: '#E74C3C' }
        ],
        humidity: [
          { max: 20, color: '#8B4513' },
          { max: 40, color: '#F4A460' },
          { max: 60, color: '#87CEEB' },
          { max: 80, color: '#4682B4' },
          { max: Infinity, color: '#191970' }
        ],
        pressure: [
          { max: 990, color: '#8B0000' },
          { max: 1000, color: '#FF6347' },
          { max: 1010, color: '#FFD700' },
          { max: 1020, color: '#90EE90' },
          { max: 1030, color: '#4169E1' },
          { max: Infinity, color: '#4B0082' }
        ]
      };

      const colorMap = colorMaps[param] || colorMaps.pm25;
      for (const range of colorMap) {
        if (value <= range.max) {
          return range.color;
        }
      }
      return colorMap[colorMap.length - 1].color;
    },
    zoomIn() {
      const view = this.map.getView();
      const zoom = view.getZoom();
      view.setZoom(zoom + 1);
    },
    zoomOut() {
      const view = this.map.getView();
      const zoom = view.getZoom();
      view.setZoom(zoom - 1);
    }
  },
  beforeUnmount() {
    if (this.map) {
      this.map.setTarget(null);
    }
  }
};
</script>

<style scoped>
.map-container {
  position: relative;
  width: 100%;
  height: 100vh;
}

.map {
  width: 100%;
  height: 100%;
}

.zoom-controls {
  position: absolute;
  top: 20px;
  right: 20px;
  display: flex;
  flex-direction: column;
  gap: 5px;
  z-index: 1000;
}

.zoom-btn {
  width: 40px;
  height: 40px;
  background: white;
  border: 2px solid rgba(0, 0, 0, 0.2);
  border-radius: 4px;
  font-size: 24px;
  font-weight: bold;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  transition: all 0.2s;
}

.zoom-btn:hover {
  background: #f0f0f0;
  border-color: rgba(0, 0, 0, 0.4);
}

.zoom-btn:active {
  transform: scale(0.95);
}
</style>
