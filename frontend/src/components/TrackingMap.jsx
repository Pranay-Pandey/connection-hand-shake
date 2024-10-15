import React, { useEffect, useRef } from "react";
import L from "leaflet";
import "leaflet/dist/leaflet.css";

const TrackingMap = ({ lat, lon, finalLat, finalLon }) => {
  const mapRef = useRef(null);
  const markerRef = useRef(null);
  const finalMarkerRef = useRef(null);

  if (!lat || !lon) {
    return <div>Loading...</div>;
  }

  useEffect(() => {
    if (!mapRef.current) {
      mapRef.current = L.map("map").setView([lat, lon], 13);
      
      L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
      }).addTo(mapRef.current);
      
      markerRef.current = L.marker([lat, lon]).addTo(mapRef.current);
      finalMarkerRef.current = L.marker([finalLat, finalLon], {
        icon: L.icon({
          iconUrl: "https://cdn.rawgit.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-red.png",
          shadowUrl: "https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.3.1/images/marker-shadow.png",
          iconSize: [25, 41],
          iconAnchor: [12, 41],
        }),
      }).addTo(mapRef.current);
    }

    markerRef.current.setLatLng([lat, lon]);  // Update marker position
    mapRef.current.setView([lat, lon], 13);   // Update map view
    finalMarkerRef.current.setLatLng([finalLat, finalLon]);  // Update final marker position
  }, [lat, lon]);

  return <div id="map" style={{ height: "50svh", width: "70vw" }}></div>;
};

export default TrackingMap;
