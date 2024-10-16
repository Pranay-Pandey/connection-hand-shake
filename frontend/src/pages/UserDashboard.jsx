import React, { useState, useEffect, useCallback } from "react";
import { FormControl } from "baseui/form-control";
import { Input } from "baseui/input";
import { Button } from "baseui/button";
import { Heading, HeadingLevel } from "baseui/heading";
import { useStyletron } from "baseui";
import { ToasterContainer, toaster } from "baseui/toast";
import Navbar from "../components/Navbar";
import { makeBooking, getPrice, getUserBookingHistory, getLocationName, getLocationCoordinates } from "../services/api";
import { Select, TYPE } from "baseui/select";
import _ from "lodash";
import TrackingMap from "../components/TrackingMap";
import { Accordion, Panel } from "baseui/accordion"; 

export default function UserDashboard() {
  const [css] = useStyletron();
  const [vehicleType, setVehicleType] = useState('');
  const [price, setPrice] = useState('');
  const [isConnected, setIsConnected] = useState(false);
  const [pickupOptions, setPickupOptions] = useState([]);
  const [dropoffOptions, setDropoffOptions] = useState([]);
  const [pickup, setPickup] = useState([]);
  const [dropoff, setDropoff] = useState([]);
  const [driverLocation, setDriverLocation] = useState({
    latitude: null,
    longitude: null,
  });
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

  const debouncedFetchPickup = useCallback(_.debounce((query) => fetchLocations(query, setPickupOptions), 500), []);
  const debouncedFetchDropoff = useCallback(_.debounce((query) => fetchLocations(query, setDropoffOptions), 500), []);

  // Fetch booking history on mount
  useEffect(() => {
    const fetchBookingHistory = async () => {
      try {
        const response = await getUserBookingHistory();
        setBookingHistory(response.data.bookings);
      } catch (error) {
        toaster.negative("Error fetching booking history.", {});
      }
    };
    fetchBookingHistory();
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
      toaster.info("Made a booking.", {});
      socket.send(JSON.stringify({ token: localStorage.getItem('token') }));
    };
    socket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      console.log("Received data:", data);
      if (data.status) {
        setDriverName(data.driver_id);
        setStatus(data.status);
        if (data.status === "booked") { 
          setDriverLocation({
            latitude: parseFloat(pickup[0].latitude),
            longitude: parseFloat(pickup[0].longitude),
          });
        }

        if (!showMap) setShowMap(true);
        if (data.status === "completed") {
          toaster.positive("Booking completed. Thank you for using our service.", {});
          setShowMap(false);
          // Reset all states
          setVehicleType('');
          setPrice('');
          setPickupOptions([]);
          setDropoffOptions([]);
          setPickup([]);
          setDropoff([]);
          setDriverLocation({
            latitude: null,
            longitude: null,
          });
          setIsConnected(false);
          setDriverName('');
          setStatus('');
        }
      }
      else {
        setDriverLocation({
          latitude: data.location.latitude,
          longitude: data.location.longitude,
        });
        console.log()
      }// Switch to map view once the booking is accepted
    };
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
        startSocketConnection();
        toaster.positive("Booking request successful. Connecting to driver...", {});
      }
    } catch (error) {
      toaster.negative("Error making booking. Please try again.", {});
    }
  };

  const renderBookingHistory = () => {
    if (bookingHistory.length === 0) {
      return <p>No booking history available.</p>;
    }

    return (
      <Accordion>
        {bookingHistory.map((booking, index) => (
          <Panel key={index} title={`Booking Status: ${booking.status}`}>
            <div className={css({
              backgroundColor: "#fff",
              padding: "10px",
              borderRadius: "8px",
              boxShadow: "0 2px 4px rgba(0, 0, 0, 0.1)",
              marginBottom: "10px"
            })}>
              <p><strong>Price:</strong> ${booking.price}</p>
              <p><strong>Pickup </strong> {booking?.pickup?.name} </p>
              <p><strong>Dropoff </strong> {booking?.dropoff?.name} </p>
              <p><strong>Created At:</strong> {new Date(booking.created_at).toLocaleString()}</p>
              <p><strong>Completed At:</strong> {booking.completed_at ? new Date(booking.completed_at).toLocaleString() : "N/A"}</p>
            </div>
          </Panel>
        ))}
      </Accordion>
    );
  };

  return (
    <div>
      <Navbar />
      <ToasterContainer autoHideDuration={3000} />
      <div className={css({
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        flexDirection: "column",
        padding: "20px",
        marginTop: "20px",
        background: "linear-gradient(135deg, #f0f4f8, #c8d6e5)",
        minHeight: "100vh",
      })}>
        <HeadingLevel>
          <Heading className={css({
            fontSize: "2rem",
            color: "#2c3e50",
            textAlign: "center",
          })}>
            User Dashboard
          </Heading>
        </HeadingLevel>

        {showMap ? (
          <>
          <TrackingMap lat={driverLocation.latitude} lon={driverLocation.longitude} 
            finalLat={dropoff[0].latitude} finalLon={dropoff[0].longitude}
          />

          <div className={css({
            backgroundColor: "#fff",
            padding: "20px",
            borderRadius: "10px",
            boxShadow: "0 4px 8px rgba(0, 0, 0, 0.1)",
            width: "100%",
            maxWidth: "400px",
            transition: "all 0.3s ease-in-out",
            marginTop: "20px",
          })}>
            <p className={css({ fontWeight: "bold" })}>Driver: {driverName}</p>
            <p className={css({ fontWeight: "bold" })}>Status: {status}</p>
          </div>

          </>
        ) : (
          <div className={css({
            backgroundColor: "#fff",
            padding: "20px",
            borderRadius: "10px",
            boxShadow: "0 4px 8px rgba(0, 0, 0, 0.1)",
            width: "100%",
            maxWidth: "400px",
            transition: "all 0.3s ease-in-out",
          })}>
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
                textAlign: 'center',
                marginTop: '20px',
              })}>
                <Button
                  type="button"
                  onClick={fetchPrice}
                  isLoading={loadingPrice}
                  overrides={{
                    BaseButton: {
                      style: {
                        width: "100%",
                        backgroundColor: "#1abc9c",
                        color: "#fff",
                      },
                    },
                  }}
                >
                  {loadingPrice ? "Fetching Price..." : "Get Price"}
                </Button>
                {price && <p className={css({ marginTop: "10px", fontWeight: "bold" })}>Estimated Price: ${price}</p>}
              </div>

              { price &&
              <div className={css({
                textAlign: "center",
                marginTop: "20px",
              })}>
                <Button type="submit">Book Request</Button>
              </div>
                }
            </form>
          </div>
        )}

        {/* Booking History Section */}
        <div className={css({
          width: "100%",
          maxWidth: "600px",
          marginTop: "40px",
          backgroundColor: "#fff",
          padding: "20px",
          borderRadius: "10px",
          boxShadow: "0 4px 8px rgba(0, 0, 0, 0.1)",
        })}>
          <HeadingLevel>
            <Heading className={css({
              fontSize: "1.5rem",
              color: "#34495e",
              textAlign: "center",
              marginBottom: "20px",
            })}>
              Booking History
            </Heading>
          </HeadingLevel>
          {renderBookingHistory()}
        </div>
      </div>
    </div>
  );
}
