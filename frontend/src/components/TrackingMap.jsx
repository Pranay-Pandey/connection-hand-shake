import React, { useEffect, useRef } from "react";
import L from "leaflet";
import "leaflet/dist/leaflet.css";

const TrackingMap = ({ lat, lon, finalLat, finalLon }) => {
  const mapRef = useRef(null);
  const markerRef = useRef(null);
  const finalMarkerRef = useRef(null);

  if (mapRef.current && markerRef.current) {
    markerRef.current.setLatLng([lat, lon]);
    console.log("updated marker");
  }

  useEffect(() => {
    if (!mapRef.current) {
      mapRef.current = L.map("map").setView([lat, lon], 5);

      L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
      }).addTo(mapRef.current);

      // Add the current location marker
      markerRef.current = L.marker([lat, lon]).addTo(mapRef.current);

      // Add the final destination marker
      finalMarkerRef.current = L.marker([finalLat, finalLon], {
        icon: L.icon({
          iconUrl: "https://cdn.rawgit.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-red.png",
          shadowUrl: "https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.3.1/images/marker-shadow.png",
          iconSize: [25, 41],
          iconAnchor: [12, 41],
        }),
      }).addTo(mapRef.current);
    }
    console.log("lat: ", lat, "lon: ", lon);

  }, [lat, lon, finalLat, finalLon]); // Dependencies to trigger re-render when these values change

  return <div id="map" style={{ height: "50vh", width: "80vw" }}></div>;
};

export default TrackingMap;
