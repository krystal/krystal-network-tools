import getApiUrl, { ServerLocation } from "./server-locations";

test("get-api-url", () => {
  expect(getApiUrl("/ip", {
    id: "test",
    url: "https://uk.example.com",
    name: "Example",
  })).toEqual(
    "https://uk.example.com/v1/ip"
  );
  expect(getApiUrl("/another-endpoint/ip", {
    id: "test",
    url: "https://us-east.example.com",
    name: "Example",
  })).toEqual(
    "https://us-east.example.com/v1/another-endpoint/ip"
  );
  expect(getApiUrl("/ip?param=true", {
    id: "test",
    url: "https://us-west.example.com",
    name: "Example",
  })).toEqual(
    "https://us-west.example.com/v1/ip?param=true"
  );
});
