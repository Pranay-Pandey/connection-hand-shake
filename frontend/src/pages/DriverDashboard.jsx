import React, { useEffect, useState, useCallback } from 'react';
import { useStyletron } from 'baseui';
import { Button } from 'baseui/button';
import { FormControl } from "baseui/form-control";
import { Select, TYPE } from 'baseui/select';
import { Card, StyledBody } from 'baseui/card';
import { Heading, HeadingLevel } from "baseui/heading";
import { FlexGrid, FlexGridItem } from "baseui/flex-grid";
import { ToasterContainer, toaster } from "baseui/toast";
import { Tag } from "baseui/tag";
import { Tabs, Tab } from "baseui/tabs-motion";
import { Accordion, Panel } from "baseui/accordion";
import { ProgressBar } from "baseui/progress-bar";
import Navbar from '../components/Navbar';
import { confirmBooking, updateBookingStatus, getDriverBookingHistory, getCurrentDriverBooking } from '../services/api';

const DriverDashboard = () => {
  const [css, theme] = useStyletron();
  const [activeTab, setActiveTab] = useState("0");
  const [location, setLocation] = useState(null);
  const [ws, setWs] = useState(null);
  const [isConnected, setIsConnected] = useState(false);
  const [bookingRequest, setBookingRequest] = useState(null);
  const [journey, setJourney] = useState(false);
  const [userId, setUserId] = useState(null);
  const [userName, setUserName] = useState(null);
  const [journeyStatus, setJourneyStatus] = useState([
    { label: 'Enroute to Pickup', id: 'enroute_to_pickup' },
  ]);
  const [statusOptions, setStatusOptions] = useState([
    { label: 'Enroute to Pickup', id: 'enroute_to_pickup' },
    { label: 'Picked up', id: 'picked_up' },
    { label: 'Completed', id: 'completed' },
  ]);
  const [bookingHistory, setBookingHistory] = useState([]);
  const [journeyPickup, setJourneyPickup] = useState({
    name: "Pickup Location",
    latitude: 0,
    longitude: 0,
  });
  const [journeyDropoff, setJourneyDropoff] = useState({
    name: "Dropoff Location",
    latitude: 0,
    longitude: 0,
  });

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
    fetchBookingHistory();
    fetchCurrentBooking();
  }, []);

  const fetchCurrentBooking = async () => {
    try {
      const response = await getCurrentDriverBooking();
      if (response.status === 200 && response.data.booking) {
        const data = response.data.booking;
        setJourney(true);
        setUserId(data.user_id);
        setJourneyStatus(statusOptions.filter((status) => status.id === data.status));
        setJourneyPickup(data.pickup);
        setJourneyDropoff(data.dropoff);
        setUserName(data.user_name);
        toaster.positive("Journey in progress!", {});
      }
    } catch (error) {
      toaster.negative("Error fetching current booking.", {});
    }
  };


  const startCommunication = useCallback(() => {
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
  }, [driverID, token]);

  const sendLocation = useCallback(() => {
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
          setLocation(loc.location);
          console.log('Sent location:', loc);
        });
      } else {
        console.error('Location permission denied');
      }
    } else {
      console.error('WebSocket is not open. Cannot send location.');
    }
  }, [ws, driverID]);

  const fetchBookingHistory = async () => {
    try {
      const response = await getDriverBookingHistory();
      setBookingHistory(response.data.bookings);
    } catch (error) {
      toaster.negative("Error fetching booking history.", {});
    }
  };

  const handleConfirmBooking = async () => {
    try {
      const response = await confirmBooking({ mongo_id: bookingRequest.mongo_id });
      if (response.status === 200) {
        const data = response.data;
        setJourney(true);
        setUserId(data.user_id);
        setJourneyPickup(bookingRequest.pickup);
        setJourneyDropoff(bookingRequest.dropoff);
        setUserName(bookingRequest.user_name);
        // setJourneyPickup(data.pickup);
        // setJourneyDropoff(data.dropoff);
        toaster.positive("Booking confirmed successfully!", {});
      }
    } catch (error) {
      toaster.negative("Error confirming booking. Please try again.", {});
    }
    setBookingRequest(null);
  };

  const handleIgnoreBooking = () => {
    setBookingRequest(null);
    toaster.info("Booking request ignored.", {});
  };

  const updateJourneyStatus = async () => {
    if (!journeyStatus.length || !userId) return;
    
    try {
      const response = await updateBookingStatus(userId, { status: journeyStatus[0].id });
      if (response.status === 200) {
        toaster.positive("Journey status updated successfully!", {});
        if (journeyStatus[0].id === 'completed') {
          resetJourney();
        }
      } else if (response.status === 404) {
        resetJourney();
      }
    } catch (error) {
      toaster.negative("Error updating journey status. Please try again.", {});
      resetJourney();
    }
  };

  const resetJourney = () => {
    setJourney(false);
    setUserId(null);
    setJourneyStatus([{ label: 'Enroute to Pickup', id: 'enroute_to_pickup' }]);
  };

  const renderCurrentJourney = () => (
    <Card>
      <StyledBody>
        <HeadingLevel>
          <Heading styleLevel={3}>Current Journey</Heading>
        </HeadingLevel>
        <div className={css({ marginTop: theme.sizing.scale600 })}>
          <p><strong>User Name:</strong> {userName}</p>
          <p><strong>Pickup:</strong> {journeyPickup.name}</p>
          <p><strong>Dropoff:</strong> {journeyDropoff.name}</p>
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
                  marginTop: theme.sizing.scale600,
                  backgroundColor: theme.colors.positive, 
                  ':hover': { backgroundColor: theme.colors.positive700 }
                },
              },
            }}
          >
            Update Status
          </Button>
        </div>
      </StyledBody>
    </Card>
  );

  const renderBookingNotification = () => (
    <Card>
      <StyledBody>
        <HeadingLevel>
          <Heading styleLevel={3}>New Booking Request</Heading>
        </HeadingLevel>
        <FlexGrid flexGridColumnCount={2} flexGridColumnGap="scale300" flexGridRowGap="scale300">
          <FlexGridItem><strong>User:</strong> {bookingRequest.user_name}</FlexGridItem>
          <FlexGridItem><strong>Price:</strong> ${bookingRequest.price}</FlexGridItem>
          <FlexGridItem><strong>Pickup:</strong> {bookingRequest?.pickup?.name}</FlexGridItem>
          <FlexGridItem><strong>Dropoff:</strong> {bookingRequest?.dropoff?.name}</FlexGridItem>
        </FlexGrid>
        <div className={css({ display: 'flex', justifyContent: 'space-between', marginTop: theme.sizing.scale800 })}>
          <Button onClick={handleIgnoreBooking} kind="secondary">Ignore</Button>
          <Button onClick={handleConfirmBooking}>Confirm</Button>
        </div>
      </StyledBody>
    </Card>
  );

  const renderBookingHistory = () => (
    <Card>
      <StyledBody>
        <HeadingLevel>
          <Heading styleLevel={3}>Booking History</Heading>
        </HeadingLevel>
        {bookingHistory.length === 0 ? (
          <p>No booking history available.</p>
        ) : (
          <Accordion>
            {bookingHistory.map((booking, index) => (
              <Panel key={index} title={`Booking #${index + 1} - ${booking.user_name} - ${booking.status}`}>
                <FlexGrid flexGridColumnCount={2} flexGridColumnGap="scale300" flexGridRowGap="scale300">
                  <FlexGridItem><strong>Price:</strong> ${booking.price}</FlexGridItem>
                  <FlexGridItem><strong>Status:</strong> <Tag closeable={false}>{booking.status}</Tag></FlexGridItem>
                  <FlexGridItem><strong>Pickup:</strong> {booking?.pickup?.name}</FlexGridItem>
                  <FlexGridItem><strong>Dropoff:</strong> {booking?.dropoff?.name}</FlexGridItem>
                  <FlexGridItem><strong>Created:</strong> {new Date(booking.created_at).toLocaleString()}</FlexGridItem>
                  <FlexGridItem><strong>Completed:</strong> {booking.completed_at ? new Date(booking.completed_at).toLocaleString() : "N/A"}</FlexGridItem>
                </FlexGrid>
              </Panel>
            ))}
          </Accordion>
        )}
      </StyledBody>
    </Card>
  );

  return (
    <div>
      <Navbar />
      <ToasterContainer autoHideDuration={3000} />
      <div className={css({
        padding: "40px",
        backgroundColor: theme.colors.backgroundPrimary,
        minHeight: "100vh",
      })}>
        <HeadingLevel>
          <Heading styleLevel={2} className={css({
            marginBottom: theme.sizing.scale800,
            color: theme.colors.primary,
          })}>
            Driver Dashboard
          </Heading>
        </HeadingLevel>

        <Tabs
          activeKey={activeTab}
          onChange={({ activeKey }) => setActiveTab(activeKey)}
          activateOnFocus
          renderAll
        >
          <Tab title="Current Journey">
            <div className={css({ marginTop: theme.sizing.scale600 })}>
              {journey && userId ? renderCurrentJourney() : (
                <Card>
                  <StyledBody>
                    <p>No active journey. Waiting for booking requests...</p>
                    {isConnected ? (
                      <ProgressBar value={10} infinite />
                    ) : (
                      <Button onClick={startCommunication}>Connect to Server</Button>
                    )}
                  </StyledBody>
                </Card>
              )}
              {bookingRequest && renderBookingNotification()}
            </div>
          </Tab>
          <Tab title="Booking History">
            <div className={css({ marginTop: theme.sizing.scale600 })}>
              {renderBookingHistory()}
            </div>
          </Tab>
        </Tabs>
      </div>
    </div>
  );
};

export default DriverDashboard;