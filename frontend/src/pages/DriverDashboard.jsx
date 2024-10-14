import React, { useEffect, useState } from 'react';
import Navbar from '../components/Navbar';
import { Button } from 'baseui/button';
import { FormControl } from "baseui/form-control";
import { Input } from "baseui/input";
import { Card } from 'baseui/card';
import { useStyletron } from 'baseui';
import BookingNotification from '../components/BookingNotification'; // Import the custom notification component
import { confirmBooking, updateBookingStatus } from '../services/api';

const DriverdashBoard = () => {
  if (!localStorage.getItem('token') || localStorage.getItem('userType') !== 'driver') {
    window.location.href = '/driver/login';
  }

  const [css] = useStyletron();
  const [location, setLocation] = useState(null);
  const [ws, setWs] = useState(null);
  const [isConnected, setIsConnected] = useState(false);
  const [bookingRequest, setBookingRequest] = useState(null); // State to store booking request
  const [journey, setJourney] = useState(false);
  const [userId, setUserId] = useState(null);
  const [journeyStatus, setJourneyStatus] = useState(false);

  const token = localStorage.getItem('token');
  const driverID = localStorage.getItem('driverID');

  const getUpdatedLocation = () => {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition((position) => {
        setLatitude(position.coords.latitude);
        setLongitude(position.coords.longitude);
      });
    }
  };

  useEffect(() => {
    getUpdatedLocation();
  }, []);

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
      console.log('Received booking request:', data);
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

  const handleConfirmBooking = async () => {
    console.log('Booking confirmed');
    // Handle confirmation logic here
    const response = await confirmBooking({
      "mongo_id" : bookingRequest.mongo_id
    })
    if (response.status === 200) {
      const data = response.data;
      console.log('Booking confirmed successfully');
      setJourney(true);
      setUserId(data.user_id);
    }
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
            location: {
              latitude: position.coords.latitude,
              longitude: position.coords.longitude,
            }, 
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

  const resetJourney = () => {
    setJourney(false);
    setUserId(null);
    setJourneyStatus(null);
  };

  const updateJourneyStatus = async () => {
    if (!journeyStatus) {
      console.error('Journey status is required');
      return;
    }
    if (!userId) {
      console.error('User ID is required');
      return;
    }
    const data = {
      status: journeyStatus,
    };
    try{
      const response = updateBookingStatus(userId, data);
      if (response.status === 404) {
        resetJourney();
      }
    } catch (error) {
      resetJourney();
    }
    if (data.status === 'completed') {
      resetJourney();
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

        { journey && userId && (
          <Card overrides={{ Root: { style: { width: '400px', marginTop: '20px' } } }}>
            <h3>Current Journey</h3>
            <p>User ID: {userId}</p>

            <FormControl
              label="status"
              caption="Please enter the status of the journey to update"
            >
              <Input
                placeholder="Status"
                value={journeyStatus}
                onChange={(e) => setJourneyStatus(e.target.value)}
              />
            </FormControl>
            <Button onClick={updateJourneyStatus}>
              Update Journey Status
            </Button>

          </Card>
          )}

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
