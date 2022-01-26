import { FC } from "react";
import { BrowserRouter, Routes, Route } from "react-router-dom";

import App from "./app";

import Home from "../pages/home/home.view";
import Ping from "../pages/ping/ping.view";
import Traceroute from "../pages/traceroute/traceroute.view";
import Whois from "../pages/whois/whois.view";
import Dns from "../pages/dns/dns.view";
import ReverseDns from "../pages/reverse-dns/reverse-dns.view";
import BgpRoutes from "../pages/bgp-route/bgp-route.view";

const AppRouter: FC = () => {
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

export default AppRouter;
