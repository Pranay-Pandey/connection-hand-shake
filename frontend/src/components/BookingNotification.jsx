import React from 'react';
import { Card, StyledBody } from 'baseui/card';
import { Button } from 'baseui/button';
import { useStyletron } from 'baseui';

const BookingNotification = ({ booking, onConfirm, onIgnore }) => {
  const [css] = useStyletron();

  return (
    <Card overrides={{ Root: { style: { width: '400px', margin: '20px auto' } } }}>
      <StyledBody>
        <h3>New Booking Request</h3>
        <p><strong>User:</strong> {booking.user_name}</p>
        <p><strong>Pickup:</strong> {booking.pickup.latitude}, {booking.pickup.longitude}</p>
        <p><strong>Dropoff:</strong> {booking.dropoff.latitude}, {booking.dropoff.longitude}</p>
        <p><strong>Price:</strong> ${booking.price}</p>
      </StyledBody>
      <div className={css({ display: 'flex', justifyContent: 'space-between', padding: '10px 20px' })}>
        <Button onClick={onIgnore} kind="secondary">Ignore</Button>
        <Button onClick={onConfirm}>Confirm</Button>
      </div>
    </Card>
  );
};

export default BookingNotification;
