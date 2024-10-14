import * as React from "react";
import { Accordion, Panel } from "baseui/accordion";
import {
    Card,
    StyledBody,
    StyledAction
  } from "baseui/card";
  import { Button } from "baseui/button";

export default () => {
  return (
    <Accordion
      onChange={({ expanded }) => console.log(expanded)}
    >
      <Panel title="User">
        <Card>
            <StyledBody>
                Login as a user to book a delivery
            </StyledBody>
            <StyledAction>
                <Button
                overrides={{
                    BaseButton: {
                    style: {
                        width: "100%"
                    }
                    }
                }}
                onClick={() => {
                    window.location.href = '/user/login';
                }}
                >
                Login
                </Button>
            </StyledAction>
        </Card>
        <Card>
            <StyledBody>
                Register as a user
            </StyledBody>
            <StyledAction>
                <Button
                overrides={{
                    BaseButton: {
                    style: {
                        width: "100%"
                    }
                    }
                }}
                onClick={() => {
                    window.location.href = '/user/register';
                }}
                >
                Register
                </Button>
            </StyledAction>
        </Card>
      </Panel>
      <Panel title="Driver">
      <Card>
            <StyledBody>
                Login as a driver to accept delivery bookings
            </StyledBody>
            <StyledAction>
                <Button
                overrides={{
                    BaseButton: {
                    style: {
                        width: "100%"
                    }
                    }
                }}
                onClick={() => {
                    window.location.href = '/driver/login';
                }}
                >
                Login
                </Button>
            </StyledAction>
        </Card>
        <Card>
            <StyledBody>
                    Register as a driver. Be part of the ecosystem
            </StyledBody>
            <StyledAction>
                <Button
                overrides={{
                    BaseButton: {
                    style: {
                        width: "100%"
                    }
                    }
                }}
                onClick={() => {
                    window.location.href = '/driver/register';
                }}
                >
                Register
                </Button>
            </StyledAction>
        </Card>
      </Panel>
      <Panel title="Panel 3">Admin</Panel>
    </Accordion>
  );
}