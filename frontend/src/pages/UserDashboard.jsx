import React, { useState, useEffect, useCallback } from "react";
import { FormControl } from "baseui/form-control";
import { Button } from "baseui/button";
import { Heading, HeadingLevel } from "baseui/heading";
import { useStyletron } from "baseui";
import { ToasterContainer, toaster } from "baseui/toast";
import { Select, TYPE } from "baseui/select";
import { Accordion, Panel } from "baseui/accordion";
import { Card, StyledBody } from "baseui/card";
import { FlexGrid, FlexGridItem } from "baseui/flex-grid";
import { Spinner } from "baseui/spinner";
import { Tag } from "baseui/tag";
import { Tabs, Tab } from "baseui/tabs-motion";
import { ProgressBar } from "baseui/progress-bar";
import Navbar from "../components/Navbar";
import TrackingMap from "../components/TrackingMap";
import { makeBooking, getPrice, getUserBookingHistory, getLocationCoordinates, getCurrentUserBooking } from "../services/api";
import _ from "lodash";

export default function UserDashboard() {
  const [css, theme] = useStyletron();
  const [activeTab, setActiveTab] = useState("0");
  const [vehicleType, setVehicleType] = useState('');
  const [price, setPrice] = useState('');
  const [isConnected, setIsConnected] = useState(false);
  const [pickupOptions, setPickupOptions] = useState([]);
  const [dropoffOptions, setDropoffOptions] = useState([]);
  const [pickup, setPickup] = useState([]);
  const [dropoff, setDropoff] = useState([]);
  const [driverLocation, setDriverLocation] = useState({ latitude: null, longitude: null });
  const [showMap, setShowMap] = useState(false);
  const [loadingPrice, setLoadingPrice] = useState(false);
  const [vehicleOptions, setVehicleOptions] = useState([
    { id: "Light Truck", value: "light_truck" },
    { id: "Van", value: "van" },
    { id: "Truck", value: "truck" },
    { id: "Heavy Truck", value: "heavy_truck" },
    { id: "Trailer", value: "trailer" },
  ]);
  const [driverName, setDriverName] = useState('');
  const [status, setStatus] = useState('');
  const [bookingHistory, setBookingHistory] = useState([]);
  const [waitingForDriver, setWaitingForDriver] = useState(false);
  const [bookingTime, setBookingTime] = useState(null);

  const debouncedFetchPickup = useCallback(_.debounce((query) => fetchLocations(query, setPickupOptions), 500), []);
  const debouncedFetchDropoff = useCallback(_.debounce((query) => fetchLocations(query, setDropoffOptions), 500), []);

  useEffect(() => {
    const fetchBookingHistory = async () => {
      try {
        const response = await getUserBookingHistory();
        setBookingHistory(response.data.bookings);
      } catch (error) {
        toaster.negative("Error fetching booking history.", {});
      }
    };
    const checkPrevBooking = async () => {
      try {
        const response = await getCurrentUserBooking();
        if (response?.data?.booking_request) {
          setWaitingForDriver(true);
          // set current time for now
          setBookingTime(response.data.booking_request.created_at);
          setVehicleType(response.data.booking_request.vehicle_type);
          setPickup([{
            id: response.data.booking_request.pickup.name,
            latitude: response.data.booking_request.pickup.latitude,
            longitude: response.data.booking_request.pickup.longitude,
          }]);
          setDropoff([{
            id: response.data.booking_request.dropoff.name,
            latitude: response.data.booking_request.dropoff.latitude,
            longitude: response.data.booking_request.dropoff.longitude,
          }]);
          setPrice(response.data.booking_request.price.toFixed(2));

          // add a timeout of 10 minutes after the response.data.booking_request.created_at time to reset the states because after 10 minutes, the request will be expired
          const currentTime = new Date();
          const timeDiff = currentTime - new Date(response.data.booking_request.created_at);
          const timeRemaining = 10 * 60 * 1000 - timeDiff;
          setTimeout(() => {
            resetStates();
          }, timeRemaining);
        }
        else if (response?.data?.booking) {
          setWaitingForDriver(false);
          setShowMap(true);
          setDriverLocation({
            latitude: parseFloat(response.data.booking.pickup.latitude),
            longitude: parseFloat(response.data.booking.pickup.longitude),
          });
          setVehicleType(response.data.booking.vehicle_type);
          setPickup([{
            id: response.data.booking.pickup.name,
            latitude: response.data.booking.pickup.latitude,
            longitude: response.data.booking.pickup.longitude,
          }]);
          setDropoff([{
            id: response.data.booking.dropoff.name,
            latitude: response.data.booking.dropoff.latitude,
            longitude: response.data.booking.dropoff.longitude,
          }]);
          // setDriverName(response.data.booking.driver_id);
          setDriverName(response.data.booking.driver_name);
          setStatus(response.data.booking.status);
          // start socket connection
          startSocketConnection();

        }
      } catch (error) {
        console.log("Error fetching previous booking:", error);
      }
    }
    fetchBookingHistory();
    checkPrevBooking();
  }, []);


  const fetchLocations = async (query, setOptions) => {
    if (query.length < 3) return;
    try {
      // const response = await axios.get(`https://nominatim.openstreetmap.org/search?q=${query}&format=json&addressdetails=1&limit=5`);
      const response = await getLocationCoordinates(query);
      const options = response.data.map((location) => ({
        id: location.display_name,
        latitude: location.lat,
        longitude: location.lon,
      }));
      setOptions(options);
    } catch (error) {
      toaster.negative("Error fetching locations. Please try again.", {});
    }
  };

  const startSocketConnection = () => {
    const socket = new WebSocket(`ws://localhost:8080/user/ws`);
    socket.onopen = () => {
      setIsConnected(true);
      toaster.info("Connected. Waiting for a driver to accept your request.", {});
      socket.send(JSON.stringify({ token: localStorage.getItem('token') }));
    };
    socket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      console.log("Received data:", data);
      if (data.status) {
        // setDriverName(data.driver_id);
        setStatus(data.status);
        if (data.status === "booked") {
          setDriverLocation({
            latitude: parseFloat(pickup[0].latitude),
            longitude: parseFloat(pickup[0].longitude),
          });
          setWaitingForDriver(false);
          setShowMap(true);
          setDriverName(data.driver_name)
          toaster.positive("A driver has accepted your request!", {});
        }
        if (data.status === "completed") {
          toaster.positive("Booking completed. Thank you for using our service.", {});
          resetStates();
        }
      } else {
        setDriverLocation({
          latitude: data.location.latitude,
          longitude: data.location.longitude,
        });
      }
    };
  };

  const resetStates = () => {
    setVehicleType('');
    setPrice('');
    setPickupOptions([]);
    setDropoffOptions([]);
    setPickup([]);
    setDropoff([]);
    setDriverLocation({ latitude: null, longitude: null });
    setIsConnected(false);
    setDriverName('');
    setStatus('');
    setWaitingForDriver(false);
    setShowMap(false);
    setBookingTime(null);
  };

  const fetchPrice = async () => {
    if (!vehicleType || !pickup.length || !dropoff.length) {
      toaster.warning("Please provide all details to fetch the price.", {});
      return;
    }

    setLoadingPrice(true);
    toaster.info("Fetching price...", {});

    try {
      const response = await getPrice({
        vehicle_type: vehicleType,
        pickup: {
          "latitude": parseFloat(pickup[0].latitude),
          "longitude": parseFloat(pickup[0].longitude),
        },
        dropoff: {
          "latitude": parseFloat(dropoff[0].latitude),
          "longitude": parseFloat(dropoff[0].longitude),
        }
      });
      setPrice(response.data.price.toFixed(2));
      toaster.positive(`Estimated Price: $${response.data.price.toFixed(2)}`, {});
    } catch (error) {
      toaster.negative("Error fetching price. Please try again.", {});
    } finally {
      setLoadingPrice(false);
    }
  };

  const bookRequest = async (e) => {
    e.preventDefault();
    if (!vehicleType || !price || !pickup.length || !dropoff.length) {
      toaster.warning("Please provide all the details.", {});
      return;
    }
    try {
      const response = await makeBooking({
        pickup: {
          latitude: parseFloat(pickup[0].latitude),
          longitude: parseFloat(pickup[0].longitude),
          name: pickup[0].id,
        },
        dropoff: {
          latitude: parseFloat(dropoff[0].latitude),
          longitude: parseFloat(dropoff[0].longitude),
          name: dropoff[0].id,
        },
        vehicle_type: vehicleType,
        price: parseFloat(price),
      });
      if (response.status === 200) {
        setWaitingForDriver(true);
        setBookingTime(new Date());
        startSocketConnection();
      }
    } catch (error) {
      toaster.negative("Error making booking. Please try again.", {});
    }
  };

  const renderBookingForm = () => (
    <Card>
      <StyledBody>
        <HeadingLevel>
        <Heading styleLevel={3}>Book a Ride</Heading>
        </HeadingLevel>
        {showMap ? (
          <>
            <TrackingMap 
              lat={driverLocation.latitude} 
              lon={driverLocation.longitude} 
              finalLat={dropoff[0].latitude} 
              finalLon={dropoff[0].longitude}
            />
            <div className={css({ marginTop: theme.sizing.scale600 })}>
              <p><strong>Driver:</strong> {driverName}</p>
              <p><strong>Status:</strong> <Tag closeable={false}>{status}</Tag></p>
            </div>
          </>
        ) : waitingForDriver ? (
          <div className={css({ textAlign: 'center' })}>
            < HeadingLevel>
            <Heading styleLevel={4}>Waiting for a driver to accept your request</Heading>
            </HeadingLevel>
            <ProgressBar value={10} infinite />
            <div className={css({ marginTop: theme.sizing.scale600 })}>
              <p><strong>Vehicle Type:</strong> {vehicleType}</p>
              <p><strong>Pickup:</strong> {pickup[0].id}</p>
              <p><strong>Dropoff:</strong> {dropoff[0].id}</p>
              <p><strong>Estimated Price:</strong> ${price}</p>
              <p><strong>Booking Time:</strong> {bookingTime.toLocaleString()}</p>
            </div>
          </div>
        ) : (
          <form onSubmit={bookRequest}>
            <FormControl label="Vehicle Type">
              <Select
                options={vehicleOptions}
                labelKey="id"
                valueKey="value"
                placeholder="Select vehicle type"
                maxDropdownHeight="300px"
                type={TYPE.search}
                onChange={({ value }) => setVehicleType(value[0]?.value || '')}
                value={vehicleOptions.filter((option) => option.value === vehicleType)}
              />
            </FormControl>

            <FormControl label="Pickup Location">
              <Select
                options={pickupOptions}
                labelKey="id"
                valueKey="id"
                placeholder="Search pickup location"
                maxDropdownHeight="300px"
                type={TYPE.search}
                onInputChange={(e) => debouncedFetchPickup(e.target.value)}
                onChange={({ value }) => setPickup(value)}
                value={pickup}
              />
            </FormControl>

            <FormControl label="Dropoff Location">
              <Select
                options={dropoffOptions}
                labelKey="id"
                valueKey="id"
                placeholder="Search dropoff location"
                maxDropdownHeight="300px"
                type={TYPE.search}
                onInputChange={(e) => debouncedFetchDropoff(e.target.value)}
                onChange={({ value }) => setDropoff(value)}
                value={dropoff}
              />
            </FormControl>

            <div className={css({
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              marginTop: theme.sizing.scale600,
            })}>
              <Button
                onClick={fetchPrice}
                isLoading={loadingPrice}
                kind="secondary"
              >
                {loadingPrice ? <Spinner /> : "Get Price"}
              </Button>
              {price && <p className={css({ fontWeight: "bold" })}>Estimated Price: ${price}</p>}
            </div>

            {price && (
              <div className={css({
                marginTop: theme.sizing.scale600,
              })}>
                <Button type="submit">Book Request</Button>
              </div>
            )}
          </form>
        )}
      </StyledBody>
    </Card>
  );

  const renderBookingHistory = () => {
    if (bookingHistory.length === 0) {
      return <p>No booking history available.</p>;
    }

    return (
      <Card>
        <StyledBody>
          <HeadingLevel>
          <Heading styleLevel={3}>Booking History</Heading>
          </HeadingLevel>
          <Accordion>
            {bookingHistory.map((booking, index) => (
              <Panel key={index} title={`Booking #${index + 1} - ${booking.driver_name} - ${booking.status}`}>
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
        </StyledBody>
      </Card>
    );
  };

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
            User Dashboard
          </Heading>
        </HeadingLevel>

        <Tabs
          activeKey={activeTab}
          onChange={({ activeKey }) => setActiveTab(activeKey)}
          activateOnFocus
          renderAll
        >
          <Tab title="Book a Ride">
            <div className={css({ marginTop: theme.sizing.scale600 })}>
              {renderBookingForm()}
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
}