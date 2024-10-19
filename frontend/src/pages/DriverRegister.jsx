import React, {useState} from "react";
import { FormControl } from "baseui/form-control";
import { Input } from "baseui/input";
import { Button } from "baseui/button";
import {Heading, HeadingLevel} from 'baseui/heading';
import { registerDriver } from "../services/api";
import { Select, TYPE } from "baseui/select";
import { StyledLink } from "baseui/link";
import { useStyletron } from "baseui";

export default function UserRegister() {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [vehicleId, setVehicleId] = useState("");
  const [vehilceType, setVehicleType] = useState("");
  const [vehicleVolumen, setVehicleVolumen] = useState("");
  const [vehicleOptions, setVehicleOptions] = useState([
    { id: "Light Truck", value: "light_truck" },
    { id: "Van", value: "van" },
    { id: "Truck", value: "truck" },
    { id: "Heavy Truck", value: "heavy_truck" },
    { id: "Trailer", value: "trailer" },
  ]);
  const [css] = useStyletron();

  const register = async () => {
    try {
      if (!name || !email || !password || !vehicleId || !vehilceType || !vehicleVolumen) {
        alert('Please fill out all fields');
        return;
      }
      const response = await registerDriver({ name, email, password, vehicleId, "vehicleType": vehilceType, "vehicleVolume": vehicleVolumen });
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
      <Select
        options={vehicleOptions}
        labelKey="id"
        valueKey="value"
        type={TYPE.search}
        onChange={({ value }) => setVehicleType(value[0]?.value || '')}
        value={vehicleOptions.filter((option) => option.value === vehilceType)}
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