# controlplane

controlplane для SRE тренинга

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
предыдущем шаге. Полученный секрет *registry-slurm-io* используется для
получения образов контейнеров в дальнейшем. Если вы уже создавали
секрет с таким именем, команда провалится, это нормально.

Находясь в верхней директории сервиса, выполните:

```shell
helm upgrade --install --atomic --values .helm/values.yaml controlplane .helm
```

Установка занимает какое-то время, будьте терпеливы.

Далее, нужно пробросить порт сервиса на локальную машину:

```shell
kubectl port-forward deployment/controlplane 4000
```

Теперь сервис доступен на локальном порту 4000, проверяем:

```shell
curl -s http://localhost:4000/list-of-teams 
```
