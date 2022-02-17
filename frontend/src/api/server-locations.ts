export type ServerLocation = {
  name: string;
  id: string;
  url: string | undefined; // undefined is here in case of env variables. Embedded values should NEVER be undefined.
};

// Load in the location from where is applicable.
export let locations: ServerLocation[] = [];
const el = document.getElementById("__ktools_regions_blob");
if (el) {
  // Use the embedded values here.
  locations = JSON.parse(el.innerText) as ServerLocation[];
} else {
  // Fallback to the default values.
  locations = [
    {
      name: "London",
      id: "london",
      url: process.env.REACT_APP_BACKEND_LONDON_ORIGIN,
    },
    {
      name: "US East",
      id: "us-east",
      url: process.env.REACT_APP_BACKEND_US_EAST_ORIGIN,
    },
    {
      name: "US West",
      id: "us-west",
      url: process.env.REACT_APP_BACKEND_US_WEST_ORIGIN,
    },
  ];
}

export const getLocationById = (id: string) =>
  locations.find((l) => l.id === id);

const getApiUrl = (
  endpoint: string,
  location: ServerLocation,
) => {
  let u = location.url;
  if (!u) u = "/";
  else if (!u!.endsWith("/")) u += "/";
  u += "v1" + endpoint;
  return u;
};

export default getApiUrl;
