## Using Docker Compose
Use the provided example `docker-compose.yaml` file as a starting point:

```yaml
services:
  expo-outlook-bookinghandler:
    image: ghcr.io/teknikens-hus/expo-outlook-bookinghandler:latest
    env_file: ".env"
    volumes:
      - "./config.yaml:/app/config.yaml"
      - "./appdata:/app/data"
    restart: unless-stopped
```
Adjust the values in the config.yaml.example file and then rename it to config.yaml

The datavolume stores the text file that keeps track of the sent emails, so if the container is restarted, it will not send the same emails again.

To see options for the config file, check the main [README.md](../../README.md)

Then run:
```bash
docker-compose up -d
```