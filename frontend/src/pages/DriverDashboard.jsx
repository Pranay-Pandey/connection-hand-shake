import React, { useEffect, useState } from 'react';
import Navbar from '../components/Navbar';
import { Button } from 'baseui/button';
import { FormControl } from "baseui/form-control";
import { Input } from "baseui/input";
import { Select, TYPE } from 'baseui/select';
import { Card } from 'baseui/card';
import { useStyletron } from 'baseui';
import BookingNotification from '../components/BookingNotification'; 
import { confirmBooking, updateBookingStatus } from '../services/api';

const DriverdashBoard = () => {
  const [css] = useStyletron();
  const [location, setLocation] = useState(null);
  const [ws, setWs] = useState(null);
  const [isConnected, setIsConnected] = useState(false);
  const [bookingRequest, setBookingRequest] = useState(null);
  const [journey, setJourney] = useState(false);
  const [userId, setUserId] = useState(null);
  const [journeyStatus, setJourneyStatus] = useState([
    { label: 'Enroute to Pickup', id: 'enroute_to_pickup' },
  ]);
  const [statusOptions, setStatusOptions] = useState([
    { label: 'Enroute to Pickup', id: 'enroute_to_pickup' },
    { label: 'Picked up', id: 'picked_up' },
    { label: 'completed', id: 'completed' },
  ]);

  const token = localStorage.getItem('token');
  const driverID = localStorage.getItem('driverID');

  useEffect(() => {
    const interval = setInterval(() => {
      sendLocation();
    }, 15000);

    return () => clearInterval(interval);
  }, [ws]);

  useEffect(() => {
    startCommunication();
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

  const handleConfirmBooking = async () => {
    const response = await confirmBooking({ mongo_id: bookingRequest.mongo_id });
    if (response.status === 200) {
      const data = response.data;
      setJourney(true);
      setUserId(data.user_id);
    }
    setBookingRequest(null); 
  };

  const handleIgnoreBooking = () => {
    setBookingRequest(null);
  };

  const updateJourneyStatus = async () => {
    if (!journeyStatus.length || !userId) return;
    
    try {
      const response = await updateBookingStatus(userId, { status: journeyStatus[0].id });
      if (response.status === 404 || journeyStatus[0].id === 'completed') {
        resetJourney();
      }
    } catch (error) {
      resetJourney();
    }
  };

  const resetJourney = () => {
    setJourney(false);
    setUserId(null);
    setJourneyStatus(null);
  };

  return (
    <div>
      <Navbar />
      <div className={css({
        padding: '20px',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        marginTop: '50px',
      })}>
        <Card overrides={{ Root: { style: { width: '450px', padding: '20px', textAlign: 'center' } } }}>
          <h2 className={css({ marginBottom: '20px' })}>Driver Dashboard</h2>
        </Card>

        {journey && userId && (
          <Card overrides={{ Root: { style: { width: '450px', marginTop: '20px', padding: '20px', textAlign: 'center' } } }}>
            <h3>Current Journey</h3>
            <p>User ID: {userId}</p>
            <FormControl label="Update Journey Status">
              <Select
                options={statusOptions}
                labelKey="label"
                valueKey="id"
                type={TYPE.select}
                value={journeyStatus}
                onChange={({ value }) => setJourneyStatus(value)}
                maxDropdownHeight="300px"
              />
            </FormControl>
            <Button 
              onClick={updateJourneyStatus}
              overrides={{
                BaseButton: {
                  style: { 
                    width: '100%', 
                    backgroundColor: '#FF5A5F', 
                    ':hover': { backgroundColor: '#E04848' }
                  },
                },
              }}
            >
              Update Status
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
