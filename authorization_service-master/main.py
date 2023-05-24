""" Инициализация приложения """
import logging
from os import getenv
import sys
import time
import uuid
from asyncio import sleep
from aiohttp import web
from aiohttp.hdrs import ACCEPT
import aioredis
from aioprometheus import Counter, Histogram, Registry, render
import jwt
from misc import compute_slowdown, should_error, fetch_settings_for, track_endpoint_using_metrics

routes = web.RouteTableDef()


async def has_team(team_name, redis):
    return team_name in list(map(str, ((await redis.hgetall('list-of-teams', encoding='utf-8')).keys())))


def get_expire_time():
    env_expire_time = int(getenv('JWT_HMAC_SECRET_EXPIRE'))
    return int(time.time()) + (env_expire_time if env_expire_time > 0 else 60)


@routes.get('/')
@track_endpoint_using_metrics
async def auth(request: web.Request) -> web.Response:
    redis = request.app['redis']
    team_name = getenv('PROVIDER_AUTH_HEADER')
    client_identity = request.headers.getone(team_name, None)
    request['city'] = client_identity
    if not client_identity:
        request.app['logger'].debug(f'Empty header. Request is not authorized')
        return web.json_response(['Not Authorized'], status=401)
    settings = await fetch_settings_for(client_identity, redis)
    request.app['logger'].debug(f'Fetched problems settings {settings} for {client_identity}')
    if slowdown := compute_slowdown(settings['slowdown_probability'],
                                    settings['slowdown_min'],
                                    settings['slowdown_max']):
        request.app['logger'].debug(f'{client_identity} sleeps for {slowdown} sec.')
        await sleep(slowdown)
    if should_error(settings['error_probability']):
        request.app['logger'].debug(f'Send 500 for {client_identity} ')
        return web.json_response([], status=500)
    if not await has_team(client_identity, redis):
        request.app['logger'].debug(f'{client_identity} is not authorized')
        return web.json_response(['Not Authorized'], status=401)
    token = jwt.encode({'token': str(uuid.uuid4()), 'exp': get_expire_time()},
                       getenv('JWT_HMAC_SECRET'),
                       algorithm='HS256')
    return web.json_response([], headers={'X-Auth-Operation-Id': token.decode('utf-8')})


@routes.get('/metrics')
async def handle_metrics(request):
    content, http_headers = render(request.app['prometheus_registry'], request.headers.getall(ACCEPT, []))
    return web.Response(body=content, headers=http_headers)


@routes.get('/health')
async def healthcheck(request):
    """
    Хелзчек, который всегда 200
    :param request: объект с параметрами входящего запроса
    :return: 200 OK
    """
    return web.json_response({'result': 'ok'})


async def on_startup(app) -> None:
    """
    Колбек инициализации приложения
    :param app: объект сервера aiohttp
    :return: None
    """
    logger = logging.getLogger()
    logger.setLevel(logging.DEBUG)
    handler = logging.StreamHandler(sys.stdout)
    handler.setLevel(logging.DEBUG)
    formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
    handler.setFormatter(formatter)
    logger.addHandler(handler)
    app['logger'] = logger
    logger.info('startup event')
    app['redis'] = await aioredis.create_redis_pool('redis://' + getenv('REDIS_HOST'), password=getenv('REDIS_PASSWORD'))


async def on_shutdown(app) -> None:
    """
    Колбек завершения работы приложения
    :param app: объект сервера aiohttp
    :return: None
    """
    app['logger'].info('finish event')

if __name__ == '__main__':
    application = web.Application()
    prometheus_registry = Registry()
    application['requests_total'] = Counter('http_server_requests_total',
                                            'The total number of HTTP requests handled by the application')
    application['requests_duration'] = Histogram('http_server_request_duration_seconds',
                                                 'The HTTP response duration',
                       buckets=[0.005, 0.01, 0.025, 0.05, 0.1, 0.3, 0.5, 0.7, 0.9, 1, 1.1, 1.3, 1.5, 1.7, 1.9, 2.0, 2.5, 5, 10])
    prometheus_registry.register(application['requests_total'])
    prometheus_registry.register(application['requests_duration'])
    application['prometheus_registry'] = prometheus_registry
    application.add_routes(routes)
    application.on_startup.append(on_startup)
    application.on_shutdown.append(on_shutdown)
    web.run_app(application, host='0.0.0.0', port=int(getenv('SERVICE_PORT', '2111')))
