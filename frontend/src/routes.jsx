import React from 'react';
import { Routes, Route } from 'react-router-dom';
import UserRegister from './components/UserRegister';
import UserLogin from './components/UserLogin';
import UserProfile from './components/UserProfile';
// import DriverRegister from './components/DriverRegister';
// import DriverLogin from './components/DriverLogin';
// import DriverProfile from './components/DriverProfile';

const AppRoutes = () => {
  return (
    <Routes>
      <Route path="/user/register" element={<UserRegister />} />
      <Route path="/user/login" element={<UserLogin />} />
      <Route path="/user/profile" element={<UserProfile />} />
      {/* <Route path="/driver/register" element={<DriverRegister />} />
      <Route path="/driver/login" element={<DriverLogin />} />
      <Route path="/driver/profile" element={<DriverProfile />} /> */}
    </Routes>
  );
};

export default AppRoutes;
