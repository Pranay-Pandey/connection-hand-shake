import React, {useState} from "react";
import { FormControl } from "baseui/form-control";
import { Input } from "baseui/input";
import { Button } from "baseui/button";
import {Heading, HeadingLevel} from 'baseui/heading';
import { registerDriver } from "../services/api";
import { StyledLink } from "baseui/link";
import { useStyletron } from "baseui";

export default function UserRegister() {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [vehicleId, setVehicleId] = useState("");
  const [vehilceType, setVehicleType] = useState("");
  const [vehicleVolumen, setVehicleVolumen] = useState("");
  const [css] = useStyletron();

  const register = async () => {
    try {
      const response = await registerDriver({ name, email, password, vehicleId, vehilceType, vehicleVolumen });
      if (response.status === 200) {
        console.log(response.data);
        window.location.href = '/driver/login';
      }
    }
    catch (error) {
      console.error(error);
    }
  }

  return (
    <div>
    <HeadingLevel>
      <Heading>Register Driver</Heading>
    </HeadingLevel>
    <FormControl
      label="Name"
      caption="Please enter your name"
    >
      <Input
        value={name}
        onChange={(e) => setName(e.target.value)}
        placeholder="Name"
      />
    </FormControl>
    <FormControl
      label="Email"
      caption="Please enter your email address"
    >
      <Input
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="Email"
      />
    </FormControl>
    <FormControl
      label="Password"
      caption="Please enter your password"
    >
      <Input
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        placeholder="Password"
        type="password"
      />
    </FormControl>
    <FormControl
      label="Vehicle ID"
      caption="Please enter your vehicle ID"
    >
      <Input
        value={vehicleId}
        onChange={(e) => setVehicleId(e.target.value)}
        placeholder="Vehicle ID"
      />
    </FormControl>
    <FormControl
      label="Vehicle Type"
      caption="Please enter your vehicle type"
    >
      <Input
        value={vehilceType}
        onChange={(e) => setVehicleType(e.target.value)}
        placeholder="Vehicle Type"
      />
    </FormControl>
    <FormControl
      label="Vehicle Volumen"
      caption="Please enter your vehicle volumen"
    >
      <Input
        value={vehicleVolumen}
        onChange={(e) => setVehicleVolumen(e.target.value)}
        placeholder="Vehicle Volumen"
      />
    </FormControl>


    <div>

    <StyledLink href="/driver/login">
      Already have an account? Login here.
    </StyledLink>

    </div>

    <div className={css({margin: '0 auto', marginTop: '20px'})}>

    <Button onClick={register}>Register</Button>

    </div>

    </div>
  );

}