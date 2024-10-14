import React, {useState} from "react";
import {
  AppNavBar
} from "baseui/app-nav-bar";
import {
  Overflow,
  CircleCheckFilled
} from "baseui/icon";
import { useStyletron } from "baseui";

export default function Navbar(){
    const [css] = useStyletron();
    const name = localStorage.getItem('name') || 'User';
    const userType = localStorage.getItem('userType');

    const changeSection = (item) => {
        if (item.label === 'Logout') {
            localStorage.clear();
            window.location.href = '/';
        }
    }

    return (
    <div className={css({
        zIndex: 100,
        width: "100vw",
      })}>
      <AppNavBar
        title="Logistics Platform"
        username={name}
        usernameSubtitle={userType}
        userItems={[
          {
            icon: CircleCheckFilled,
            label: "Profile"
          },
          {
            icon: Overflow,
            label: "Logout"
          }
        ]}
        onUserItemSelect={changeSection}
      />
      </div>
    );
  }