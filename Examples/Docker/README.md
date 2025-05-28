## Using Docker Compose
Use the provided example `docker-compose.yaml` file as a starting point:

```yaml
services:
  expo-outlook-bookinghandler:
    image: ghcr.io/teknikens-hus/expo-outlook-bookinghandler:latest
    env_file: ".env"
    volumes:
      - "./config.yaml:/app/config.yaml"
    restart: unless-stopped
```
Adjust the values in the config.yaml.example file and then rename it to config.yaml

To see options for the config file, check the main [README.md](../../README.md)

Then run:
```bash
docker-compose up -d
```