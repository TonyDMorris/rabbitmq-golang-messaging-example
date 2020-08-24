import React, { useEffect } from "react";
import { ToastProvider, useToasts } from "react-toast-notifications";

import styled from "styled-components";

const Messages = () => {
  const { addToast } = useToasts();

  const ws = new WebSocket("ws://localhost:3000/ws");
  ws.onopen = () => {
    // on connecting, do nothing but log it to the console
    addToast("connected");
  };

  ws.onmessage = (evt) => {
    // listen to data sent from the websocket server
    const message = JSON.parse(evt.data);

    addToast(message);
  };

  ws.onclose = () => {
    addToast("disconnected");
  };

  return <Container>hello</Container>;
};

const Container = styled.div`
  width: 100%;
  height: 100%;
`;

export default Messages;
