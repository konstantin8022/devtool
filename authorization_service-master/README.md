###Переменные окружения для запуска и корректной работы         
------
| Переменная | Назначение | Пример |
|------------|------------|--------|
| SERVICE_PORT  | Порт, по которому доступен сервис | 2111 | 
| REDIS_HOST  | Хост Redis | localhost | 
| REDIS_PASSWORD  | Пароль Redis | ws@2weaQ |
| JWT_HMAC_SECRET_EXPIRE  | время жизни токена JWT | 60 |
| JWT_HMAC_SECRET  | Секрет для генерации токена для передачи в хедере X-Auth-Operation-Id. Этот токен проверяется на main backend и должен совпадать со значением на main backend | 72884861-ea71-44ab-8f55-b8f2a13f46a8 | 
| PROVIDER_AUTH_HEADER  | Заголовок, который проверяет токен(имя) команды для идентификации | X-Slurm-Source-Provider |



