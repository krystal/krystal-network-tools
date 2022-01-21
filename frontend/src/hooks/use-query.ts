import { useEffect, useState } from "react";
import getApiUrl from "../helpers/get-api-url";
import { DEFAULT_REQUEST_OPTIONS, RequestOptions } from "../helpers/request";

const useQuery = <T>(url: string, options: RequestOptions = {}) => {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<unknown | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setLoading(true);
    fetch(getApiUrl(url, options.location), {
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
