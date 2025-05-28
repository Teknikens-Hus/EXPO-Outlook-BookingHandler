# Using Kubernetes deployment manifest
Use the provided example [deployment manifests](./expo-outlook-bookinghandler-deployment/) as a starting point:

# [Deployment](./expo-outlook-bookinghandler-deployment/deployment.yaml)
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: expo-outlook-bookinghandler
  labels:
    app: expo-outlook-bookinghandler
spec:
  replicas: 1
  strategy:
    type: RollingUpdate
  selector:
    matchLabels:
      app: expo-outlook-bookinghandler
  template:
    metadata:
      labels:
        app: expo-outlook-bookinghandler
    spec:
      containers:
        - image: ghcr.io/teknikens-hus/expo-outlook-bookinghandler:latest
          name: expo-outlook-bookinghandler
          env:
            - name: TZ
              value: "Europe/Stockholm"
            - name: Interval
              value: "1800"
            - name: EXPO_TOKEN
              valueFrom:
                secretKeyRef:
                  name: expo-outlook-bookinghandler-secret
                  key: expo_token
            - name: EXPO_URL
              value: "https://booking.yourdomain.com"
            - name: SENDGRID_APIKEY
              valueFrom:
                secretKeyRef:
                  name: expo-outlook-bookinghandler-secret
                  key: sendgrid_apikey
          # Adjust the resource limits as needed
          resources:
            requests:
              memory: "128Mi"
              cpu: "10m"
            limits:
              memory: "256Mi"
              cpu: "200m"
          volumeMounts:
            - name: config-volume
              mountPath: /app/config.yaml
              subPath: config.yaml
            - name: data-volume
              mountPath: /app/data
          # Make sure we run as non-root "app"
          securityContext:
            runAsUser: 1001
            runAsGroup: 2001
            fsGroup: 2001
      volumes:
        - name: config-volume
          configMap:
            name: config-configmap
        - name: data-volume
          persistentVolumeClaim:
            claimName: expo-outlook-bookinghandler-pvc
```
## [Kustomization](./expo-outlook-bookinghandler-deployment/kustomization.yaml)
The kustomization has a configMapGenerator
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: default
metadata:
  name: kustomize-expo-outlook-bookinghandler
resources:
- deployment.yaml
configMapGenerator:
  - name: config-configmap
    files:
      - config.yaml=config.yaml
    options:
      disableNameSuffixHash: false
```


## Config.yaml
Create and adjust the values in the config.yaml file

To see options for the config file, check the example file [config.yaml.example](../config.yaml.example)