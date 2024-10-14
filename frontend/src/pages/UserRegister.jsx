import React, {useState} from "react";
import { FormControl } from "baseui/form-control";
import { Input } from "baseui/input";
import { Button } from "baseui/button";
import {Heading, HeadingLevel} from 'baseui/heading';
import { registerUser } from "../services/api";
import { StyledLink } from "baseui/link";
import { useStyletron } from "baseui";

export default function UserRegister() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [name, setName] = useState("");
  const [css] = useStyletron();

  const register = async () => {
    try {
      const response = await registerUser({ email, password, name });
      if (response.status === 200) {
        console.log(response.data);
        window.location.href = '/user/login';
      }
    }
    catch (error) {
      console.error(error);
    }
  }


  return (
    <div>
    <HeadingLevel>
      <Heading>Register</Heading>
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

    <div>

    <StyledLink href="/user/login">
      Already have an account? Login here.
    </StyledLink>

    </div>

    <div className={css({margin: '0 auto', marginTop: '20px'})}>

    <Button onClick={register}>Register</Button>

    </div>

    </div>
  );

}