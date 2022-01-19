import { FC } from "react";
import { BrowserRouter, Routes, Route } from "react-router-dom";

import App from "./app";

import Home from "../pages/home";
import Ping from "../pages/ping";
import Traceroute from "../pages/traceroute";
import Whois from "../pages/whois";
import Dns from "../pages/dns";
import ReverseDns from "../pages/reverse-dns";
import BgpRoutes from "../pages/bgp-route";

const Router: FC = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<App />}>
          <Route index element={<Home />} />
          <Route path="/ping" element={<Ping />} />
          <Route path="/traceroute" element={<Traceroute />} />
          <Route path="/whois" element={<Whois />} />
          <Route path="/dns" element={<Dns />} />
          <Route path="/reverse-dns" element={<ReverseDns />} />
          <Route path="/bgp-route" element={<BgpRoutes />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
};

export default Router;
