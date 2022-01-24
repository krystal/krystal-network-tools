export enum ServerLocation {
  LONDON = "london",
  US_EAST = "us-east",
  US_WEST = "us-west",
}

const getBackendUrl = (location: ServerLocation) => {
  switch (location) {
    case "london":
      return process.env.REACT_APP_BACKEND_LONDON_ORIGIN;
    case "us-east":
      return process.env.REACT_APP_BACKEND_US_EAST_ORIGIN;
    case "us-west":
      return process.env.REACT_APP_BACKEND_US_WEST_ORIGIN;
  }
};

const getApiUrl = (
  endpoint: string,
  location: ServerLocation = ServerLocation.LONDON
) => {
  return getBackendUrl(location) + "/v1" + endpoint;
};

export default getApiUrl;
