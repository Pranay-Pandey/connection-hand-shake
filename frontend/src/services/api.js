
// src/services/api.js
import axios from 'axios'

const AUTH_URL = 'http://localhost:8081' 
const BOOKING_URL = 'http://localhost:8084'
export const NOTIFICATION_URL = 'localhost:8080'
export const ADMIN_URL = 'http://localhost:8085/admin'
export const PRICING_URL = 'http://localhost:8086'

const authApi = axios.create({
  baseURL: AUTH_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

const bookingApi = axios.create({
  baseURL: BOOKING_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

const adminApi = axios.create({
  baseURL: ADMIN_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

const pricingApi = axios.create({
  baseURL: PRICING_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})
  
authApi.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers['Authorization'] = `Bearer ${token}`
  }
  return config
})

bookingApi.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers['Authorization'] = `Bearer ${token}`
  }
  return config
})

pricingApi.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers['Authorization'] = `Bearer ${token}`
  }
  return config
})

export const registerUser = (userData) => authApi.post('/user/register', userData)
export const loginUser = (credentials) => authApi.post('/user/login', credentials)
export const getUserProfile = () => authApi.get('/user/profile')

export const registerDriver = (driverData) => authApi.post('/driver/register', driverData)
export const loginDriver = (credentials) => authApi.post('/driver/login', credentials)
export const getDriverProfile = (id) => authApi.get(`/driver/profile/${id}`)

export const getPrice = (data) => pricingApi.post('/pricing/estimate', data)

export const makeBooking = (bookingData) => bookingApi.post('/booking', bookingData)
export const confirmBooking = (bookingId) => bookingApi.post(`/booking/accept`, bookingId)
export const updateBookingStatus = (userId, bookingData) => bookingApi.patch(`/booking/${userId}`, bookingData)
export const getUserBookingHistory = () => bookingApi.get('/user/booking-history')
export const getDriverBookingHistory = () => bookingApi.get('/driver/booking-history')

export const getFleetStats = () => adminApi.get('/fleet-stats')
export const getDriverPerformance = () => adminApi.get('/driver-performance')
export const getBookingAnalytics = () => adminApi.get('/booking-analytics')
export const getVehicleLocations = () => adminApi.get('/vehicle-locations')

export const getLocationName = (lat, long) => axios.get(`https://nominatim.openstreetmap.org/reverse?lat=${lat}&lon=${long}&format=json`)
export const getLocationCoordinates = (query) => axios.get(`https://nominatim.openstreetmap.org/search?q=${query}&format=json&addressdetails=1&limit=5`)