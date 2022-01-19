import { useEffect, useState } from "react";

const useQuery = <T>(url: string, options: RequestInit = {}) => {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<unknown | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setLoading(true);
    fetch(url, options)
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
