# tickets

## Project setup
```shell
npm install
```

### Compiles and hot-reloads for development
```shell
npm run serve
```

### Compiles and minifies for production
```shell
npm run build
```

### Run selenium test
```
npm run test
```

### Customize configuration
See [Configuration Reference](https://cli.vuejs.org/config/).

# minikube setup

```shell
kubectl create secret docker-registry registry-slurm-io --docker-server=registry.slurm.io --docker-username=USER --docker-password=PASS
helm upgrade --install --atomic --values .helm/values.yaml tickets .helm
kubectl port-forward deployment/tickets 8080 &
curl -s http://localhost:8080

```
