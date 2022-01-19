import { render, screen } from "@testing-library/react";
import App from "./app";

test("The app renders", () => {
  render(<App />);
  const appElement = screen.getByTestId("app");
  expect(appElement).toBeInTheDocument();
});
