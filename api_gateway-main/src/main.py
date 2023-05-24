""" Инициализация приложения """
from asyncio import TimeoutError
import logging
from os import getenv
import sys
from time import time
from traceback import format_exc
from aiohttp import web, ClientSession, client_exceptions, hdrs, ClientTimeout
import aioredis
from aioprometheus import Counter, Histogram, Registry, render
from src.misc import card_service_middleware, get_tracing_headers
import json

@card_service_middleware
async def proxy_request_to_cinema_api(request: web.Request):
    """
    Проксирование запроса в бэкенд кинотеатра
    :param request: объект с параметрами входящего запроса
    :return: ответ от кинотеатара или ошибка
    """

    method = request.method.lower()
    base_url = request.headers.getone('X-Api-Url')
    extended_url = request.match_info.get('tail').rstrip('/')
    query_string = request.query_string

    discovered_addr = (base_url + '/' + extended_url + (('?' + query_string) if query_string else ''))

    try:
        payload = json.loads(await request.read())
        payload.pop('card')
    except KeyError:
        payload = json.loads(await request.read())
    except Exception:
        payload = {}

    headers = await get_tracing_headers(request)
    header_deadline = request.headers.getone('X-Slurm-RPC-Deadline', None)

    if header_deadline is not None:
        headers['X-Slurm-RPC-Deadline'] = header_deadline

    request.app['logger'].debug(f'Proxing request {method} to {discovered_addr} with headers {headers} and payload {payload}')

    response = await getattr(request.app['client_session'], method)(discovered_addr, data=json.dumps(payload), headers=headers)
    x_auth_header = response.headers.getone('x-auth-operation-id', None)

    if x_auth_header is not None:
        headers = {'x-auth-operation-id': x_auth_header}
    else:
        headers = {}

    return web.json_response(await response.json(), status=response.status, headers=headers)

async def handle_metrics(request):
    content, http_headers = render(request.app['prometheus_registry'], request.headers.getall(hdrs.ACCEPT, []))
    return web.Response(body=content, headers=http_headers)


async def healthcheck(request):
    """
    Хелзчек, который всегда 200
    :param request: объект с параметрами входящего запроса
    :return: 200 OK
    """
    return web.json_response({'result': 'ok'})

def create_app() -> web.Application:
    """
    Создание прложения и настройка модулей
    :return: объект приложения
    """
    app = web.Application()
    app.on_startup.append(on_startup)
    return app


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
    prometheus_registry = Registry()
    application['requests_total'] = Counter('http_server_requests_total',
                                            'The total number of HTTP requests handled by the application')
    application['requests_duration'] = Histogram('http_server_request_duration_seconds',
                       'The HTTP response duration',
                       buckets=[0.005, 0.01, 0.025, 0.05, 0.1, 0.3, 0.5, 0.7, 0.9, 1, 1.1, 1.3, 1.5, 1.7, 1.9, 2.0, 2.5, 5, 10])
    prometheus_registry.register(application['requests_total'])
    prometheus_registry.register(application['requests_duration'])
    application['prometheus_registry'] = prometheus_registry

    app['client_session'] = ClientSession()

    app.add_routes([web.get('/metrics', handle_metrics),
                    web.get('/health', healthcheck),
                    web.get('/{tail:.*}', proxy_request_to_cinema_api),
                    web.post('/{tail:.*}', proxy_request_to_cinema_api),
                    web.delete('/{tail:.*}', proxy_request_to_cinema_api),
                    ])


if __name__ == '__main__':
    application = create_app()
    web.run_app(application, host='0.0.0.0', port=int(getenv('SERVICE_PORT', '2113')))
