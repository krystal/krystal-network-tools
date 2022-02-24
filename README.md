<h1 align="center">Krystal Network Tools</h1>

<p align="center">
  <a href="https://github.com/krystal/krystal-network-tools/actions">
    <img src="https://img.shields.io/github/workflow/status/krystal/krystal-network-tools/CI.svg?logo=github" alt="Actions Status">
  </a>
  <a href="https://github.com/krystal/krystal-network-tools/commits/main">
    <img src="https://img.shields.io/github/last-commit/krystal/krystal-network-tools.svg?style=flat&logo=github&logoColor=white"
alt="GitHub last commit">
  </a>
  <a href="https://github.com/krystal/krystal-network-tools/issues">
    <img src="https://img.shields.io/github/issues-raw/krystal/krystal-network-tools.svg?style=flat&logo=github&logoColor=white"
alt="GitHub issues">
  </a>
  <a href="https://github.com/krystal/krystal-network-tools/pulls">
    <img src="https://img.shields.io/github/issues-pr-raw/krystal/krystal-network-tools.svg?style=flat&logo=github&logoColor=white" alt="GitHub pull requests">
  </a>
  <a href="https://github.com/krystal/krystal-network-tools/blob/main/MIT-LICENSE">
    <img src="https://img.shields.io/github/license/krystal/krystal-network-tools.svg?style=flat" alt="License Status">
  </a>
</p>

<p align="center">
    <a href="#setup">Setup</a> | <a href="#development">Development</a>
</p>

Several mini-tools compiled into one package by Krystal to handle networking. Made with love by [Krystal](https://k.io).

This package contains the following bits of functionality:
- BGP (providing bird is enabled)
- DNS
- Reverse DNS
- Ping
- WHOIS
- IP finding

## Setup
To setup the Krystal network tools, you will want to create the [multi-host configuration](#multi-host-configuration) if you are not running this on a single host. You will also want to set `DNS_SERVER` to allow for DNS changes to be found that might be cached by 1.1.1.1. From here, you have multiple options when it comes to deployment:

1) **Run using guvnor (preferred):** The preferred way to deploy this is with guvnor. To do this, you should install guvnor and copy the file `example_guvnor_config.yml` into `/etc/guvnor/services/krystal-network-tools.yaml`. You will then want to change `<dns server here>` to your dns server, `<dns host here>` as the value of the DNS hostname, 
2) **Run using Docker Compose:** To use Docker Compose, you will want to make a `.env` file. From here, add the `DNS_SERVER` environment variable described above and then add the `HTTPS_HOST` variable with the hostname of the server you are deploying to. From here, install Bird 2 (if you want this), make a `regions.yml`, and you're done!
3) **Run as a static binary:** If you so wish, you can go ahead and run the binary from the releases page. If the `HTTPS_HOST` environment variable is set, it will allow you to go ahead and setup HTTPS certificates for that host and will manage that for you by serving on port 443 and 80. Otherwise, it will serve on port 127.0.0.1:8080 (or whatever port is in `PORT`). From here, the application will abide X-Forwarded-For and you can setup your own proxy.

It is important to note that if Bird 2 is not mounted in the filesystem, the BGP functionality will not work.

### Multi-host configuration
For multiple hosts, you will want to create a regions.yml file. The file should contain a list of regions, with each item looking like the following in the file:
```yaml
- id: <short id for region>
  name: <name of region>
  url: https://region.example.com
```
From here, if you are running this outside of a container, you will want to have `regions.yml` in the current working directory when you run the binary. If it is in a container, you will want to mount it in `/var/app/regions.yml`.

## Development
Both frontend and backend can be launched on their own accords from the relevant Dockerfile's in each directory. Note that to generate the routes for the React content, you also need to edit `backend/frontend.go` with any new routes/the relevant titles.
