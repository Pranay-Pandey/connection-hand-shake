import React, { useEffect, useState } from 'react';
import Navbar from '../components/Navbar';
import { Button } from 'baseui/button';
import { Card } from 'baseui/card';
import { useStyletron } from 'baseui';
import BookingNotification from '../components/BookingNotification'; // Import the custom notification component

const DriverdashBoard = () => {
  if (!localStorage.getItem('token') || localStorage.getItem('userType') !== 'driver') {
    window.location.href = '/driver/login';
  }

  const [css] = useStyletron();
  const [location, setLocation] = useState(null);
  const [ws, setWs] = useState(null);
  const [isConnected, setIsConnected] = useState(false);
  const [bookingRequest, setBookingRequest] = useState(null); // State to store booking request

  const token = localStorage.getItem('token');
  const driverID = localStorage.getItem('driverID');

  const startCommunication = () => {
    if (!driverID) {
      console.error('Driver ID is required');
      return;
    }
    const socket = new WebSocket(`ws://localhost:8080/driver/ws`);

    socket.onopen = () => {
      console.log('WebSocket connected');
      setWs(socket);
      setIsConnected(true);
      socket.send(JSON.stringify({ token, driverID }));
    };

    socket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      // Set the booking request data to show the notification
      setBookingRequest(data);
    };

    socket.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    socket.onclose = () => {
      console.log('WebSocket disconnected');
      setWs(null);
      setIsConnected(false);
    };

    return () => {
      socket.close();
    };
  };

  const handleConfirmBooking = () => {
    console.log('Booking confirmed');
    // Handle confirmation logic here, e.g., send confirmation to server
    setBookingRequest(null); // Hide notification after confirmation
  };

  const handleIgnoreBooking = () => {
    console.log('Booking ignored');
    // Handle ignore logic here
    setBookingRequest(null); // Hide notification after ignoring
  };

  useEffect(() => {
    const interval = setInterval(() => {
      sendLocation();
    }, 15000);

    return () => clearInterval(interval);
  }, [ws]);

  const sendLocation = () => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      if (navigator.geolocation) {
        navigator.geolocation.getCurrentPosition((position) => {
          const loc = {
            latitude: position.coords.latitude,
            longitude: position.coords.longitude,
            driverID,
            timestamp: new Date().toISOString(),
          };
          ws.send(JSON.stringify(loc));
          console.log('Sent location:', loc);
        });
      } else {
        console.error('Location permission denied');
      }
    } else {
      console.error('WebSocket is not open. Cannot send location.');
    }
  };

  return (
    <div>
      <Navbar />
      <div className={css({
        padding: "20px",
        display: "flex",
        justifyContent: "center",
        flexDirection: "column",
        alignItems: "center",
        marginTop: "50px"
      })}>
        <Card overrides={{ Root: { style: { width: '400px' } } }}>
          <h2>Driver Dashboard</h2>
          <Button onClick={startCommunication} disabled={isConnected}>
            Start Communication
          </Button>
        </Card>
        {bookingRequest && (
          <BookingNotification
            booking={bookingRequest}
            onConfirm={handleConfirmBooking}
            onIgnore={handleIgnoreBooking}
          />
        )}
      </div>
    </div>
  );
};

export default DriverdashBoard;
