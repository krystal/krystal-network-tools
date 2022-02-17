import { useEffect, useState } from "react";
import getApiUrl, { getLocationById, locations } from "./server-locations";
import { DEFAULT_REQUEST_OPTIONS, RequestOptions } from "./request";

const useQuery = <T>(url: string, options: RequestOptions = {}) => {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<unknown | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setLoading(true);
    fetch(getApiUrl(url, options.location ? getLocationById(options.location)! : locations[0]), {
      ...DEFAULT_REQUEST_OPTIONS,
      ...options,
    })
      .then((res) => {
        if (!res.ok) throw new Error(res.statusText);
        return res.json();
      })
      .then(setData)
      .catch(setError)
      .finally(() => setLoading(false));
  }, []); //eslint-disable-line

  return { data, loading, error };
};

export default useQuery;
