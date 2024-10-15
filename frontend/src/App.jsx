// src/App.js
import React from 'react'
import { Routes, Route } from 'react-router-dom'
import UserRegister from './pages/UserRegister'
import UserLogin from './pages/UserLogin'
import UserDashboard from './pages/UserDashboard'
import DriverRegister from './pages/DriverRegister'
import DriverLogin from './pages/DriverLogin'
import DriverDashboard from './pages/DriverDashboard'
import Home from './pages/Home'
import AdminDashboard from './pages/AdminDashboard'

function App() {
  return (
    <Routes>
      <Route path="/user/register" element={<UserRegister />} />
      <Route path="/user/login" element={<UserLogin />} />
      <Route path="/user/dashboard" element={
        // <PrivateRoute>
          <UserDashboard />
        /* </PrivateRoute> */
      } />
      <Route path="/driver/register" element={<DriverRegister />} />
      <Route path="/driver/login" element={<DriverLogin />} />
      <Route path="/driver/dashboard" element={
          <DriverDashboard />
      } />

      <Route path="/admin/dashboard" element={<AdminDashboard />} />

      <Route path="/" element={<Home />} />
    </Routes>
  )
}

export default App