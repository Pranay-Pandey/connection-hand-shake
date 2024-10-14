import React, { useState, useEffect } from "react";
import { FormControl } from "baseui/form-control";
import { Input } from "baseui/input";
import { Button } from "baseui/button";
import { Heading, HeadingLevel } from "baseui/heading";
import { useStyletron } from "baseui";
import Navbar from "../components/Navbar";
import { makeBooking } from "../services/api";

export default function UserDashboard() {
  if (!localStorage.getItem('token') || localStorage.getItem('userType') !== 'user') {
    window.location.href = '/user/login';
  }
  
  const [css] = useStyletron();
  const [vehicleType, setVehicleType] = useState('');
  const [price, setPrice] = useState('');
  const [latitude, setLatitude] = useState(null);
  const [longitude, setLongitude] = useState(null);
  const [ws, setWs] = useState(null);
  const [isConnected, setIsConnected] = useState(false);
  const token = localStorage.getItem('token');

  const getUpdatedLocation = () => {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition((position) => {
        setLatitude(position.coords.latitude);
        setLongitude(position.coords.longitude);
      });
    }
  }

  const startSocketConnection = () => {
    const socket = new WebSocket(`ws://localhost:8080/user/ws`);

    socket.onopen = () => {
      console.log('WebSocket connected');
      setWs(socket);
      setIsConnected(true);
      socket.send(JSON.stringify({ token }));
    };

    socket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      console.log(data);
    }

    socket.onerror = (error) => {
      console.error('WebSocket error:', error);
    }

    socket.onclose = () => {
      console.log('WebSocket disconnected');
    }
  }

  // Automatically get the user's location
  useEffect(() => {
    getUpdatedLocation();
  }, []);

  const bookRequest = async (e) => {
    getUpdatedLocation();
    e.preventDefault();
    if (!vehicleType || !price) {
      alert("Please provide all the details.");
      return;
    }

    if (!latitude || !longitude) {
      alert("Could not get your location. Please enable location services.");
      return;
    }
    try {
      const response = await makeBooking({
        pickup: { latitude, longitude },
        dropoff: { latitude, longitude }, // Can replace with actual dropoff data
        vehicle_type: vehicleType,
        price: parseFloat(price)
      });
      console.log(response);
      startSocketConnection();
    } catch (error) {
      console.error(error);
    }
  };

  return (
    <div>
      <Navbar />
      <div className={css({
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        flexDirection: "column",
        padding: "20px",
        marginTop: "20px",
      })}>
        <HeadingLevel>
          <Heading>User Dashboard</Heading>
        </HeadingLevel>

        <div className={css({
          backgroundColor: "#f0f4f8",
          padding: "20px",
          borderRadius: "10px",
          boxShadow: "0 4px 8px rgba(0, 0, 0, 0.1)",
          width: "100%",
          maxWidth: "400px",
        })}>
          <form onSubmit={bookRequest}>
            <FormControl label="Vehicle Type">
              <Input
                value={vehicleType}
                onChange={(e) => setVehicleType(e.target.value)}
                placeholder="Enter vehicle type"
              />
            </FormControl>

            <FormControl label="Price">
              <Input
                value={price}
                onChange={(e) => setPrice(e.target.value)}
                placeholder="Enter price"
                type="number"
              />
            </FormControl>

            <div className={css({ textAlign: 'center', marginTop: '20px' })}>
              <Button type="submit">Make a Booking</Button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
