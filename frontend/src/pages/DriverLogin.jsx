import React, {useState} from "react";
import { FormControl } from "baseui/form-control";
import { Input } from "baseui/input";
import { Button } from "baseui/button";
import {Heading, HeadingLevel} from 'baseui/heading';
import { loginDriver } from "../services/api";
import { StyledLink } from "baseui/link";
import { useStyletron } from "baseui";

export default function UserLogin() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [css] = useStyletron();

  const login = async () => {
    try {
      const response = await loginDriver({ email, password });
      if (response.status === 200) {
        console.log(response.data);
        localStorage.setItem('token', response.data.token);
        localStorage.setItem('userType', 'driver');
        localStorage.setItem('driverID', response.data.ID);
        localStorage.setItem('name', response.data.name);
        window.location.href = '/driver/dashboard';
      }
    }
    catch (error) {
      console.error(error);
    }
  };


  return (
    <div>
    <HeadingLevel>
      <Heading>Login</Heading>
    </HeadingLevel>
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

    <div>
    <StyledLink href="/driver/register">
      Don't have an account? Register here.
    </StyledLink>
    </div>

    <div className={css({margin: '0 auto', marginTop: '20px'})}>
    <Button onClick={login}>
      Login</Button>
    </div>
    </div>
  );

}