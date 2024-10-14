import React, { useEffect, useState } from 'react';
import { NOTIFICATION_URL } from '../services/api';
import { Label } from 'baseui/form-control/styled-components';
import Navbar from '../components/Navbar';

const DriverdashBoard = () => {
    if (!localStorage.getItem('token') || localStorage.getItem('userType') !== 'driver') {
        window.location.href = '/driver/login';
    } 

    const [location, setLocation] = useState(null);
    const [ws, setWs] = useState(null);
    const token = localStorage.getItem('token');
    const driverID = localStorage.getItem('driverID');
    const [isConnected, setIsConnected] = useState(false);

    const startCommunication = () => {
        if (driverID === 0) {
            console.error('Driver ID is required');
            return;
        }
        const socket = new WebSocket(`ws://${NOTIFICATION_URL}/driver/ws`);

        socket.onopen = () => {
            console.log('WebSocket connected');
            setWs(socket);
            setIsConnected(true);
            socket.send(JSON.stringify({ token, driverID }));
        };

        socket.onmessage = (event) => {
            const data = JSON.parse(event.data);
            setLocation(data);
            console.log('Received location:', data);
        };

        socket.onerror = (error) => {
            console.error('WebSocket error:', error);
        };

        socket.onclose = () => {
            console.log('WebSocket disconnected');
            setWs(null);
            setIsConnected(false);
        };

        // Cleanup function to close the socket when the component unmounts
        return () => {
            socket.close();
        };
    };

    useEffect(() => {
        const interval = setInterval(() => {
            sendLocation();
        }, 15000); // Send location every 15 seconds

        // Cleanup interval on component unmount
        return () => clearInterval(interval);
    }, [ws]); // Depend on ws to start sending only when it's open

    const sendLocation = () => {
        if (ws && ws.readyState === WebSocket.OPEN) {
            const loc = {
                latitude: Math.random() * 100,  // Replace with actual latitude
                longitude: Math.random() * 100, // Replace with actual longitude
                driverID,
                timestamp: new Date().toISOString(),
            };
            ws.send(JSON.stringify(loc));
            console.log('Sent location:', loc);
        } else {
            console.error('WebSocket is not open. Cannot send location.');
        }
    };

    return (
        <div>
            <Navbar />
            <h1>WebSocket Driver Location</h1>
            <p>Driver ID: {driverID}</p>
            <button onClick={startCommunication} disabled={isConnected}>
                Start Communication
            </button>
            {location && (
                <div>
                    <h2>Last Location:</h2>
                    <p>Driver ID: {location.driverID}</p>
                    <p>Latitude: {location.latitude}</p>
                    <p>Longitude: {location.longitude}</p>
                    <p>Timestamp: {location.timestamp}</p>
                </div>
            )}
        </div>
    );
};

export default DriverdashBoard;
