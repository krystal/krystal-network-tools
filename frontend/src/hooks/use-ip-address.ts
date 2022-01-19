import useQuery from "./use-query";

const IP_ADDRESS_API_URL = "https://lg-london-v4.labs.k.io/v1/user/ip_address";

type IpAddressResponse = {
  ip_address: string;
};

const useIpAddress = () => {
  const { data, loading, error } =
    useQuery<IpAddressResponse>(IP_ADDRESS_API_URL);

  return { ipAddress: data?.ip_address, loading, error };
};

export default useIpAddress;
