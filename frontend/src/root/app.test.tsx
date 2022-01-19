import { render, screen } from "@testing-library/react";
import { BrowserRouter } from "react-router-dom";
import App from "./app";

test("The app renders", () => {
  render(
    <BrowserRouter>
      <App />
    </BrowserRouter>
  );
  const appElement = screen.getByTestId("app");
  expect(appElement).toBeInTheDocument();
});
