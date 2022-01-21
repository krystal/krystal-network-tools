import getApiUrl, { ServerLocation } from "./get-api-url";

test("get-api-url", () => {
  expect(getApiUrl("/ip", ServerLocation.LONDON)).toEqual(
    "https://uk.example.com/v1/ip"
  );
  expect(getApiUrl("/another-endpoint/ip", ServerLocation.US_EAST)).toEqual(
    "https://us-east.example.com/v1/another-endpoint/ip"
  );
  expect(getApiUrl("/ip?param=true", ServerLocation.US_WEST)).toEqual(
    "https://us-west.example.com/v1/ip?param=true"
  );
});
