from datetime import datetime
import json
from os import getenv
import re
import jwt
from aiohttp import web


PATH_ID_RE = re.compile(R'\d+')


async def get_providers(redis):
    """
    Получение списка провайдеров из Redis
    :param redis: объект подключения к редису
    :return: список провайдеров
    """
    return {k: json.loads(v) for k, v in (await redis.hgetall('list-of-teams', encoding='utf-8')).items()}


async def get_provider_timeout(redis):
    """
    Получение таймаумта провайдера из редис или окружения
    :param redis: объект подключения к редису
    :return: таймаумт провайдера
    """
    for timeout in [await redis.get('provider_timeout'), getenv('PROVIDERS_TIMEOUT'), 60]:
        if timeout is not None:
            return float(timeout)

def check_auth_operation_id(encoded_jwt):
    """
    Проверка токена доступа???
    :param encoded_jwt: закодированный токен
    :return: раскодированный токен
    """
    try:
        jwt.decode(encoded_jwt, getenv('JWT_HMAC_SECRET'), algorithms=['HS256'])
        return True
    except jwt.InvalidTokenError:
        return False


def track_endpoint_using_metrics(decorated):
    """ Декторатор для трекинг метрик на обращение к эндпоинту"""
    async def wrapper(request: web.Request):
        """
        Трекинг метрик на обращение к эндпоинту
        :param request: объект входящего запроса
        :return: объект ответа
        """
        start = datetime.now()
        try:
            response = await decorated(request)
            code = response.status
        except Exception as err:
            code = 500
            response = web.json_response({'errors': [{'title': 'unexpected error'}]}, status=500)
        end = datetime.now() - start
        city = request['city']
        path = re.sub(PATH_ID_RE, 'ID', request.path)
        prometheus_labels = {'path': path, 'method': request.method.lower(),
                             'city': city}
        request.app['requests_duration'].observe(prometheus_labels, end.total_seconds())
        request.app['requests_total'].inc({**prometheus_labels, 'code': code})
        return response
    return wrapper
