import React, {useState} from "react";
import { FormControl } from "baseui/form-control";
import { Input } from "baseui/input";
import { Button } from "baseui/button";
import {Heading, HeadingLevel} from 'baseui/heading';
import { makeBooking } from "../services/api";
import { useStyletron } from "baseui";
import Navbar from "../components/Navbar";

export default function UserDashboard() {
  if (!localStorage.getItem('token') || localStorage.getItem('userType') !== 'user') {
    window.location.href = '/user/login';
  } 
  const [css] = useStyletron();

  const bookRequest = async (e) => {
    e.preventDefault();
    try {
      const response = await makeBooking({
        "pickup": {
          "latitude" : 9.2531256884589,
          "longitude" : 17.235698455632
        },
        "dropoff":  {
           "latitude" : 9.2531256884589,
          "longitude" : 17.235698455632
        },
        "vehicle_type": "van",
        "price": 167.6
        });
      console.log(response);
    }
    catch (error) {
      console.error(error);
    }
  };

  return (
    <div>
    <Navbar />
    <HeadingLevel>
      <Heading>User Dashboard</Heading>
    </HeadingLevel>

    <div className={css({marginTop: '16px', marginBottom: '20px'})}>
      <Button onClick={bookRequest} >Make a Booking</Button>
    </div>
    </div>
  );

}