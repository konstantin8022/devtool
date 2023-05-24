# Основной Main backend API

Берет информацию со всех кинотеатров

## Переменные окружения для запуска и корректной работы

JWT_HMAC_SECRET: Ключ для расшифровки значения заголовка X-Auth-Operation-Id. Этот заголовок подтверждает, что провайдер авторизовался на сервисе авторизации.
Секрет должен совпадать с секретом на сервисе авторизации.

PROVIDERS_TIMEOUT: таймаут для запросов к провайдеру. Если в controlplane есть какое-то значение, то оно имеет наивысший приоритет.
PROVIDER_TIMEOUT or REDIS_TIMEOUT or 60

REDIS_HOST: url

REDIS_PASSWORD: password

## Просмотр фильмов

[curl "/cities/:city_name/movies"](doc/api/movies/success_response.md)

## Просмотр сеансов

[curl "/cities/:city_name/movies/:movie_id/seances"](doc/api/seances/success_response.md)

## Покупка билетов

[curl "/cities/:city_id/movies/:movie_id/seances/:seance_id/bookings"](doc/api/bookings/success_response.md)

## Полное описание API

[Docs](doc/api/index.md)

Чтобы обновить документацию АПИ необходимо запустить команду rake docs:generate

# minikube setup

```shell
kubectl create secret docker-registry registry-slurm-io --docker-server=registry.slurm.io --docker-username=USER --docker-password=PASS
helm upgrade --install --atomic --values .helm/values.yaml main-backend .helm
kubectl port-forward deployment/main-backend 5000:3000 &
curl -s http://localhost:5000/cities
```
