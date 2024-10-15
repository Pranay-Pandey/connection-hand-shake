import React, { useState, useEffect } from 'react';
import { Heading, HeadingLevel } from "baseui/heading";
import { useStyletron } from "baseui";
import { Table } from "baseui/table-semantic";
import { Card, StyledBody } from "baseui/card";
import { Tab, Tabs } from "baseui/tabs-motion";
import { PieChart, Pie, Cell, ResponsiveContainer, BarChart, Bar, XAxis, YAxis, Tooltip, Legend } from 'recharts';
import { getFleetStats, getBookingAnalytics, getDriverPerformance, getVehicleLocations } from '../services/api';

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884d8'];

export default function AdminDashboard() {
  const [css] = useStyletron();
  const [fleetStats, setFleetStats] = useState(null);
  const [driverPerformance, setDriverPerformance] = useState([]);
  const [bookingAnalytics, setBookingAnalytics] = useState(null);
  const [vehicleLocations, setVehicleLocations] = useState([]);
  const [activeKey, setActiveKey] = useState('0');

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const fleetStatsResponse = await getFleetStats();
      setFleetStats(fleetStatsResponse.data);

      const driverPerformanceResponse = await getDriverPerformance();
      setDriverPerformance(driverPerformanceResponse.data);

      const bookingAnalyticsResponse = await getBookingAnalytics();
      setBookingAnalytics(bookingAnalyticsResponse.data);

      const vehicleLocationsResponse = await getVehicleLocations();
      setVehicleLocations(vehicleLocationsResponse.data);
    } catch (error) {
      console.error('Error fetching data:', error);
    }
  };

  const RenderFleetStats = () => {
    if (!fleetStats) return (<></>);

    const pieChartData = Object.entries(fleetStats.vehicleTypeBreakdown).map(([name, value]) => ({ name, value }));

    return (
      <Card>
        <StyledBody>
          <HeadingLevel>
            <Heading styleLevel={6}>Fleet Statistics</Heading>
          </HeadingLevel>
          <div className={css({ display: 'flex', justifyContent: 'space-between', marginBottom: '20px' })}>
            <div>
              <strong>Total Vehicles:</strong> {fleetStats.totalVehicles}
            </div>
            <div>
              <strong>Active Vehicles:</strong> {fleetStats.activeVehicles}
            </div>
          </div>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie
                data={pieChartData}
                cx="50%"
                cy="50%"
                labelLine={false}
                outerRadius={80}
                fill="#8884d8"
                dataKey="value"
                label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
              >
                {pieChartData.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </StyledBody>
      </Card>
    );
  };

  const RenderDriverPerformance = () => {
    return (
      <Card>
        <StyledBody>
          <HeadingLevel>
            <Heading styleLevel={6}>Driver Performance</Heading>
          </HeadingLevel>
          <Table
            columns={['Driver ID', 'Name', 'Trip Count', 'Avg Trip Time (min)', 'Total Revenue']}
            data={driverPerformance.map(driver => [
              driver.driverID,
              driver.name,
              driver.tripCount,
              driver.avgTripTime.toFixed(2),
              `$${driver.totalRevenue.toFixed(2)}`
            ])}
          />
        </StyledBody>
      </Card>
    );
  };

  const RenderBookingAnalytics = () => {
    if (!bookingAnalytics) return <></>;

    const data = [
      { name: 'Completed', value: bookingAnalytics.completedBookings },
      { name: 'Cancelled', value: bookingAnalytics.cancelledBookings },
      { name: 'Active', value: bookingAnalytics.totalBookings - bookingAnalytics.completedBookings - bookingAnalytics.cancelledBookings },
    ];

    return (
      <Card>
        <StyledBody>
          <HeadingLevel>
            <Heading styleLevel={6}>Booking Analytics</Heading>
          </HeadingLevel>
          <div className={css({ display: 'flex', justifyContent: 'space-between', marginBottom: '20px' })}>
            <div>
              <strong>Total Bookings:</strong> {bookingAnalytics.totalBookings}
            </div>
            <div>
              <strong>Avg Trip Time:</strong> {bookingAnalytics.avgTripTime.toFixed(2)} min
            </div>
            <div>
              <strong>Total Revenue:</strong> ${bookingAnalytics.totalRevenue.toFixed(2)}
            </div>
          </div>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={data}>
              <XAxis dataKey="name" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Bar dataKey="value" fill="#8884d8" />
            </BarChart>
          </ResponsiveContainer>
        </StyledBody>
      </Card>
    );
  };

  const RenderVehicleLocations = () => {
    return (
      <Card>
        <StyledBody>
          <HeadingLevel>
            <Heading styleLevel={6}>Vehicle Locations</Heading>
          </HeadingLevel>
          <Table
            columns={['ID', 'Name', 'Vehicle Type', 'Latitude', 'Longitude', 'Status']}
            data={vehicleLocations.map(vehicle => [
              vehicle.id,
              vehicle.name,
              vehicle.vehicleType,
              vehicle.latitude,
              vehicle.longitude,
              vehicle.status
            ])}
          />
        </StyledBody>
      </Card>
    );
  };

  return (
    <div className={css({
      padding: '20px',
      backgroundColor: '#f0f4f8',
      minHeight: '100vh',
      width: '100vw'
    })}>
      <HeadingLevel>
        <Heading styleLevel={4}>Admin Dashboard</Heading>
      </HeadingLevel>
      <Tabs
        activeKey={activeKey}
        onChange={({ activeKey }) => setActiveKey(activeKey)}
        activateOnFocus
        renderAll
      >
        <Tab title="Fleet Stats"><RenderFleetStats /></Tab>
        <Tab title="Driver Performance"><RenderDriverPerformance/></Tab>
        <Tab title="Booking Analytics"><RenderBookingAnalytics/></Tab>
        <Tab title="Vehicle Locations"><RenderVehicleLocations/></Tab>
      </Tabs>
    </div>
  );
}