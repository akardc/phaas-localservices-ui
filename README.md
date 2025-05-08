# README

## About

This tool can be used to start, stop, and view which cloned phaas services are running. Additionally you can set global
environment params to be used when starting the services (i.e. `PHAAS_VIRTUALEVENTAPIURL`).

To begin using the tool, go to the settings tab and set the paths for the shell you use (e.g. `/bin/bash`), the init
script for that shell (e.g. `/Users/username/.bash_profile`), and the path to where your repos are stored
(e.g. `/Users/username/go/github.com/BidPal`), then restart the tool. From there, you should see the populated service
list and be able to start the services.

Note that services must be using mage-lib v4.47.6 or higher for this tool to start the service.

## Future ideas

- View logs within the tool
- View and change active git branch
- Some way to select a set of services to run and automatically set their env params to point to each other
- Figure out a way to stop using using shell commands to call `mage run`

## Development

This tool is built with [Wails](https://wails.io/) and Angular. To build and run, follow directions for each of those.
