import getApiUrl, { getLocationById, locations } from './server-locations';

export type RequestOptions = RequestInit & {
    location?: string;
};

export const DEFAULT_REQUEST_OPTIONS: RequestOptions = {
    headers: {
        'Content-Type': 'application/json',
    },
};

const request = async <T>(
    endpoint: string,
    requestOptions: RequestOptions = {}
) => {
    const {location, ...options} = requestOptions;

    const url = getApiUrl(endpoint, location ? getLocationById(location)! : locations[0]);

    const res = await fetch(url, {...DEFAULT_REQUEST_OPTIONS, ...options});
    const data = await res.json();
    if (!res.ok) throw new Error(data.message || res.statusText);
    if (res.status === 206) {
        // check if there is a partial data in the response res.results
        if (!data.results) {
            throw new Error('Partial content received, but no results found in the response.');
        }
        // if the response is partial content, return the results
        return data.results as T;
    }
    return data as T;
};

export default request;
