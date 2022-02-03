const endpoint = (
  url: string,
  params: { [name: string]: string | number | boolean }
) => {
  const searchParams = new URLSearchParams();

  if (!url.startsWith("/")) {
    url = `/${url}`;
  }

  if (!url.endsWith("/")) {
    url = `${url}/`;
  }

  Object.keys(params).forEach((name) => {
    const encodedValue = encodeURIComponent(params[name]);
    if (url.includes(`/:${name}/`)) {
      url = url.replace(`/:${name}/`, `/${encodedValue}/`);
    } else {
      searchParams.append(name, `${encodedValue}`);
    }
  });

  url = url.slice(0, -1);

  if (Array.from(searchParams).length > 0) {
    return `${url}?${searchParams.toString()}`;
  } else {
    return url;
  }
};

export default endpoint;
