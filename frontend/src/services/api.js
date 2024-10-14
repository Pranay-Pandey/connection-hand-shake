
// src/services/api.js
import axios from 'axios'

const AUTH_URL = 'http://192.168.223.103:8081' 
const BOOKING_URL = 'http://192.168.223.103:8084'
export const NOTIFICATION_URL = '192.168.223.103:8080'

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

export const registerUser = (userData) => authApi.post('/user/register', userData)
export const loginUser = (credentials) => authApi.post('/user/login', credentials)
export const getUserProfile = () => authApi.get('/user/profile')

export const registerDriver = (driverData) => authApi.post('/driver/register', driverData)
export const loginDriver = (credentials) => authApi.post('/driver/login', credentials)
export const getDriverProfile = (id) => authApi.get(`/driver/profile/${id}`)

export const makeBooking = (bookingData) => bookingApi.post('/booking', bookingData)
export const confirmBooking = (bookingId) => bookingApi.post(`/booking/accept`, bookingId)
export const updateBookingStatus = (userId, bookingData) => bookingApi.patch(`/booking/${userId}`, bookingData)