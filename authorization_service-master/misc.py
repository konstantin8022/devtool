from aiohttp import web
from datetime import datetime
import random

PROBLEM_REDIS_KEYS = ['slowdown-probability', 'slowdown-min',
                      'slowdown-max', 'error-probability', 'active']


def compute_slowdown(probability, slowdown_min, slowdown_max):
    """
    Вычисление задержки
    :param probability: вероятность отказа
    :param slowdown_min: нижняя граница задержки
    :param slowdown_max: верхняя граница задержки
    :return: время задержки
    """
    if random.randint(0, 100) < probability:
        sleep_time_ms = random.randint(slowdown_min, slowdown_max)
        return sleep_time_ms / 1000


def should_error(probability):
    return random.randint(0, 100) < probability


async def fetch_settings_for(client, redis) -> dict:
    client_key = f"authproblems-{client}"
    default_key = 'authproblems-default'

    try:
        values = await redis.hmget(client_key, *PROBLEM_REDIS_KEYS, encoding='utf-8')
        # if the city-specific values are not active, fetch and use the defaults
        is_active = values[4] == 'true'
        if not is_active:
            values = await redis.hmget(default_key, *PROBLEM_REDIS_KEYS)
    except Exception as error:
        print(f'failed to read settings from redis: {error}')
        raise error

    return {
      'slowdown_probability': int(values[0] or '0'),
      'slowdown_min': int(values[1] or '0'),
      'slowdown_max': int(values[2] or '0'),
      'error_probability': int(values[3] or '0'),
      'active': is_active
    }


def track_endpoint_using_metrics(decorated):
    """ Декторатор для трекинг метрик на обращение к эндпоинту"""
    async def wrapper(request: web.Request):
        """
        Трекинг метрик на обращение к эндпоинту
        :param request: объект входящего запроса
        :return: объект ответа
        """
        start = datetime.now()
        response = await decorated(request)
        end = datetime.now() - start
        prometheus_labels = {'path': request.path, 'method': request.method.lower(),
                             'city':  request['city']}
        request.app['requests_duration'].observe(prometheus_labels, end.total_seconds())
        request.app['requests_total'].inc({**prometheus_labels, 'code': response.status})
        return response
    return wrapper
