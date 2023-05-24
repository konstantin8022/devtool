# authorization_service

Сервис авторизации для кинотеатров.

## Переменные окружения
PROVIDER_AUTH_HEADER: Заголовок, который проверяет токен(имя) команды для идентификации

RACK_ENV: Окружение в котором запускаем веб-сервер. production,development,test

JWT_HMAC_SECRET: Секрет для генерации токена для передачи в хедере X-Auth-Operation-Id. Этот токен проверяется на main backend и должен совпадать со значением на main backend

JWT_HMAC_SECRET_EXPIRE: время жизни токена JWT

REDIS_HOST: 'controlplane-redis-slave'

REDIS_PASSWORD: controlplane-redis

Чтобы обновить документацию АПИ необходимо запустить команду rake docs:generate

# Запуск сервиса в кластере minikube

См. отдельный [файл](../sre/provider_backend/minikube.md) по вопросам установки и запуска
миникуба. В этом разделе описывается установка самого микросервиса в
этом кластере.

Переходим на
[страницу](https://gitlab.slurm.io/profile/personal_access_tokens)
создания токенов доступа

Создаём токен с произвольным именем и правами *read_registry*. Фиксируем
токен для следующего шага.

Создаём секрет для доступа в регистри:

```shell
kubectl create secret docker-registry registry-slurm-io --docker-server=registry.slurm.io --docker-username=USER --docker-password=PASS
```

где *USER* - ваш логин в gitlab.slurm.io, *PASS* - токен, полученный на
предыдущем шаге. Полученный секрет *registry-slurm-io* используется за
получения образов контейнеров в дальнейшем. Если вы уже создавали
секрет с таким именем, команда провалится, это нормально.

Находясь в верхней директории сервиса, выполните:

```shell
helm upgrade --install --atomic --values .helm/values.yaml authorization-service .helm
```

Установка занимает какое-то время, будьте терпеливы.

Далее, нужно пробросить порт сервиса на локальную машину:

```shell
kubectl port-forward deployment/authorization-service 9292
```

Теперь сервис доступен на локальном порту 9292, проверяем:

```shell
curl -s http://localhost:9292 -H "X-Slurm-Source-Provider: ticket-backend-1"
```
