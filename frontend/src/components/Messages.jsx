import React, { useEffect } from "react";
import { ToastProvider, useToasts } from "react-toast-notifications";

import styled from "styled-components";

const Messages = () => {
  const { addToast } = useToasts();

  const ws = new WebSocket("ws://localhost:80/ws");

  useEffect(() => {
    ws.onopen = () => {
      // on connecting, do nothing but log it to the console
      addToast("connected", { appearance: "success" });
    };

    ws.onmessage = (evt) => {
      // listen to data sent from the websocket server

      addToast(evt.data, { appearance: "success" });
    };

    ws.onclose = () => {
      addToast("disconnected");
    };
  }, []);

  return <Container>hello</Container>;
};

const Container = styled.div`
  width: 100vw;
`;

export default Messages;
