import "./App.css";

import Messages from "./components/Messages";
import React from "react";
import { ToastProvider } from "react-toast-notifications";

const App = () => {
  return (
    <ToastProvider>
      <Messages />
    </ToastProvider>
  );
};

export default App;
