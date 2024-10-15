import React, { useState, useEffect, useCallback } from "react";
import { FormControl } from "baseui/form-control";
import { Input } from "baseui/input";
import { Button } from "baseui/button";
import { Heading, HeadingLevel } from "baseui/heading";
import { useStyletron } from "baseui";
import Navbar from "../components/Navbar";
import { makeBooking, getPrice } from "../services/api";
import { Select, TYPE } from "baseui/select";
import _ from "lodash";
import axios from "axios";
import TrackingMap from "../components/TrackingMap";

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

  const debouncedFetchPickup = useCallback(_.debounce((query) => fetchLocations(query, setPickupOptions), 500), []);
  const debouncedFetchDropoff = useCallback(_.debounce((query) => fetchLocations(query, setDropoffOptions), 500), []);

  const fetchLocations = async (query, setOptions) => {
    if (query.length < 3) return;
    try {
      const response = await axios.get(`https://nominatim.openstreetmap.org/search?q=${query}&format=json&addressdetails=1&limit=5`);
      const options = response.data.map((location) => ({
        id: location.display_name,
        latitude: location.lat,
        longitude: location.lon,
      }));
      setOptions(options);
    } catch (error) {
      console.error("Error fetching locations:", error);
    }
  };

  const startSocketConnection = () => {
    const socket = new WebSocket(`ws://localhost:8080/user/ws`);
    socket.onopen = () => {
      setIsConnected(true);
      socket.send(JSON.stringify({ token: localStorage.getItem('token') }));
    };
    socket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      setDriverLocation({
        latitude: data.location.latitude,
        longitude: data.location.longitude,
      });
      if (!showMap)
        setShowMap(true); // Switch to map view once the booking is accepted
    };
  };

  const fetchPrice = async () => {
    if (!vehicleType || !pickup.length || !dropoff.length) {
      alert("Please provide all details to fetch the price.");
      return;
    }

    setLoadingPrice(true);

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
    } catch (error) {
      console.error(error);
      alert("Error fetching price. Please try again.");
    } finally {
      setLoadingPrice(false);
    }
  };

  const bookRequest = async (e) => {
    e.preventDefault();
    if (!vehicleType || !price || !pickup.length || !dropoff.length) {
      alert("Please provide all the details.");
      return;
    }
    try {
      const response = await makeBooking({
        pickup: {
          latitude: parseFloat(pickup[0].latitude),
          longitude: parseFloat(pickup[0].longitude),
        },
        dropoff: {
          latitude: parseFloat(dropoff[0].latitude),
          longitude: parseFloat(dropoff[0].longitude),
        },
        vehicle_type: vehicleType,
        price: parseFloat(price),
      });
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
          <div className={css({
            width: "100%",
            maxWidth: "600px",
            height: "400px",
            margin: "20px 0",
            boxShadow: "0 6px 12px rgba(0,0,0,0.1)",
            borderRadius: "10px",
            overflow: "hidden",
          })}>
            <TrackingMap lat={driverLocation.latitude} lon={driverLocation.longitude} 
              finalLat={dropoff[0].latitude} finalLon={dropoff[0].longitude}
            />
          </div>
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
                <Input
                  value={vehicleType}
                  onChange={(e) => setVehicleType(e.target.value)}
                  placeholder="Enter vehicle type"
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
              </div>

              {price && (
                <div className={css({
                  marginTop: "20px",
                  textAlign: "center",
                  color: "#27ae60",
                  fontSize: "1.5rem",
                })}>
                  Estimated Price: ${price}
                </div>
              )}

              <div className={css({
                textAlign: 'center',
                marginTop: '20px',
              })}>
                <Button type="submit" overrides={{
                  BaseButton: {
                    style: {
                      width: "100%",
                      backgroundColor: "#1abc9c",
                      color: "#fff",
                    },
                  },
                }}>
                  Make a Booking
                </Button>
              </div>
            </form>
          </div>
        )}
      </div>
    </div>
  );
}
