# EXPO-Outlook-BookingHandler
This program written in golang allow you to deploy a docker container/pod that fetches bookings from a [EXPO-booking](https://www.expobooking.info/) system using graphQL queries, filters out only the bookings with resources and compares them with bookings in Outlook ICS calendars. If a booking in Outlook overlaps with a booking in EXPO, it sends an email to the Outlook-booker notifying them of the overbooking. 

## Supported Calendars
Currently, the following calendars are supported:
- ICS

## Installation
Currently amd64 and arm64 are supported.

### Docker / docker-compose:
[![Docker Icon](https://skillicons.dev/icons?i=docker&theme=light)](./Examples/Docker/README.md)

### Kubernetes Deployment:
[![Kube Icon](https://skillicons.dev/icons?i=kubernetes&theme=light)](./Examples/Kubernetes/README.md)

## Configuration
The configuration is done using a config.yaml file and environment variables.
The file should be mounted into the container at `/app/config.yaml` 
You can find an example config file in the [Examples](./Examples/config.yaml.example) folder.


### ENV variables

| Key        | Description                                                                 | Example Value                          |
|------------|-----------------------------------------------------------------------------|----------------------------------------|
| EXPO_TOKEN   | Your EXPO API Token                            | ``d651vgdexnt55jzu5rnxiwak1u8x3oxjd93jx3j9xj39r4g1ao``            |
| EXPO_URL   | The base URL to your booking site                  | `https://booking.yourdomain.com`                     |
| SMTP_PASSWORD   | Your SMTP password for sending emails              | `Password123 or apikey`                             |
| SMTP_USERNAME   | Your SMTP username for sending emails              | `Username123`                             |
| SMTP_HOST   | Your SMTP host for sending emails              | `smtp.yourdomain.com`                             |
| SMTP_PORT   | Your SMTP port for sending emails              | `default is 587 if not specified`                             |
| TZ   | Your [TZ identifier](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) for your timezone                      | `Europe/Stockholm`                             |
| Interval   | The interval in seconds at which the overlap check is performed                   | `1800`

Please note these are example keys/tokens, not actual values you should use or that are valid. 


### Check logs!
If you are having issues, check the logs of the application/container. It should give you some direction of whats wrong.

## Development
There's two ways to run the application, either using golang directly or using docker-compose.
### Using Golang
1. To run the application using golang, you need to have golang installed on your machine. You can find the installation instructions on the [golang website](https://golang.org/doc/install). Check the go.mod file for the required version.
2. Clone the repository to your local machine.
3. Run `go mod download` to download the dependencies.
4. Create a `config.yaml` file next to the main.go file with the required values.
5. cd into the cmd/EXPO-Outlook-BookingHandler directory.
6. Run `go run main.go` to start the application.
7. The sent_mails.txt file might fail to create since it wants to save to /app directory (inside the container)

To create the env variables on Windows you can use the following command in PowerShell:
```powershell
$env:NAME-OF-ENV = "Value of env"
```
On MacOS/Linux you can use the following command in the terminal:
```bash
export NAME-OF-ENV="Value of env"
```


### Using Docker
1. Clone the repository
2. Install Docker Desktop
3. Create a .env file in the root of the project next to the docker-compose.yaml file with the same required values explained above.
4. Create a config.yaml file in the root of the project with the required values.
5. Run `docker-compose up --build` to start the application.

## Release new version
Make sure code works.
Increment the version in the version.txt and push. Github actions will automatically build and push the new version to the gha registry.