""" Инициализация приложения """
from asyncio import TimeoutError
import logging
import random
from os import getenv
import sys
from time import time
from traceback import format_exc
from aiohttp import web, ClientSession, client_exceptions, hdrs, ClientTimeout
import aioredis
from aioprometheus import Counter, Histogram, Registry, render
from src.misc import check_auth_operation_id, get_provider_timeout, get_providers, track_endpoint_using_metrics


@track_endpoint_using_metrics
async def proxy_request_to_cinema_api(request: web.Request):
    """
    Проксирование запроса в бэкенд кинотеатра
    :param request: объект с параметрами входящего запроса
    :return: ответ от кинотеатара или ошибка
    """
    def optionaly_slashed(path):
        return path + ('' if path.endswith('/') else '/')

    start_time = time()
    app = request.app
    redis = app['redis']
    city = request.match_info.get('city')
    request['city'] = city
    provider = (await get_providers(redis)).get(city)
    if not provider:
        return web.json_response({'errors': {'title': 'City not found'}}, status=404, headers=app['CORS'])

    provider_url = request.match_info.get('provider_url').rstrip('/')
    query_string = request.query_string

    if provider['apiGateway'].startswith('http'):
        base_url = provider['apiGateway']
    else:
        base_url = provider['providerURL']

    discovered_addr = (optionaly_slashed(base_url) + provider_url + (('?' + query_string) if query_string else ''))

    header_timeout = 60
    header_deadline = request.headers.getone('X-Slurm-RPC-Deadline', None)
    if header_deadline is not None:
        header_timeout = int(header_deadline) - int(time())

    provider_timeout = await get_provider_timeout(redis)

    request_timeout = min(header_timeout, provider_timeout)
    headers = {'X-Slurm-RPC-Deadline': f"{(request_timeout + time()):.2f}"}

    tracing_headers = await get_tracing_headers(request)
    headers = {**headers, **tracing_headers}

    if provider['apiGateway'].startswith('http'):
        headers = {**headers, 'X-Api-Url': provider['providerURL'] }

    method = request.method.lower()
    app['logger'].debug(f'Proxing request {method} to {discovered_addr} with headers {headers}')

    try:
        response = await getattr(app['client_session'], method)(discovered_addr,
                                                                data=await request.read(),
                                                                headers=headers,
                                                                timeout=ClientTimeout(total=request_timeout))
    except web.HTTPException as err:
        status = err.status
        if status >= 500:
            request.app['logger'].error('Failure during HTTP request', exc_info=True)
        return web.json_response({'errors': [{'title': "Failed to send HTTP request"}]},
                status=status)

    except client_exceptions.ClientConnectionError:
        app['logger'].error('Failed to connect', exc_info=True)
        return web.json_response({'errors': [{'title': 'Failed to connect'}]},
                                 status=503, headers=app['CORS'])

    except TimeoutError:
        app['logger'].error('Request timeout exceeded', exc_info=True)
        return web.json_response({'errors': [{'title': f'Request timeout exceeded', 'detail': f'request_timeout={request_timeout})'}]},
                status=503, headers=app['CORS'])

    if response.status >= 200 and response.status < 300 and not check_auth_operation_id(response.headers.getone('x-auth-operation-id')):
        app['logger'].error('Failed to validate auth service token')
        return web.json_response({'errors': [{'title': f'Failed to validate auth service token'}]}, status=503, headers=app['CORS'])

    status = response.status
    body = await response.json()
    app['logger'].info(f'Request {method} to {discovered_addr} with headers {headers} completed. Received response {status}')
    return web.json_response(body, status=status, headers=app['CORS'])


async def options_cors(request):
    """
    Ответ на CORS запрос
    """
    return web.json_response({}, headers=request.app['CORS'])


async def get_cities(request):
    """
    Получение записей с городами
    :param request: объект с параметрами входящего запроса
    :return:
    """
    cities = [{'id': city['name'], 'name': city['name']} for city in (await get_providers(request.app['redis'])).values()]
    return web.json_response({'data': cities}, headers=request.app['CORS'])

async def get_tracing_headers(request):
    tracing_need_headers = ['x-request-id', 'x-b3-traceid', 'x-b3-spanid', 'x-b3-parentspanid', 'x-b3-sampled', 'x-b3-flags']

    tracing_headers = {}

    for k,v in request.headers.items():
        if k in tracing_need_headers:
            tracing_headers[k] = v

    return tracing_headers

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

async def get_card_service_errors_probability(request):
    """
    Проверка в редисе процент ошибок для card-service
    :return: integer
    """
    rate = await request.app['redis'].get('card_service_errors_probability', encoding='utf-8')

    if rate is not None and rate != '':
        rate = rate
    else:
        rate = 0

    return web.json_response({'rate': rate})

def create_app() -> web.Application:
    """
    Создание прложения и настройка модулей
    :return: объект приложения
    """
    app = web.Application()
    app.on_startup.append(on_startup)
    app.on_shutdown.append(on_shutdown)
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
    app['CORS'] = {'Access-Control-Allow-Headers': '*',
                   'Access-Control-Allow-Methods': 'POST',
                   "Access-Control-Allow-Origin": "*"}
    app['redis'] = None
    app['client_session'] = ClientSession()
    app['redis'] = await aioredis.create_redis_pool('redis://' + getenv('REDIS_HOST'), password=getenv('REDIS_PASSWORD'))
    app.add_routes([web.get('/cities/{city}/{provider_url:.*}', proxy_request_to_cinema_api),
                    web.post('/cities/{city}/{provider_url:.*}', proxy_request_to_cinema_api),
                    web.delete('/cities/{city}/{provider_url:.*}', proxy_request_to_cinema_api),
                    web.options('/cities/{city}/{provider_url:.*}', options_cors),
                    web.get('/metrics', handle_metrics),
                    web.get('/health', healthcheck),
                    web.get('/cities', get_cities),
                    web.get('/api/v1/config/card_service', get_card_service_errors_probability)
                    ])


async def on_shutdown(app) -> None:
    """
    Колбек завершения работы приложения
    :param app: объект сервера aiohttp
    :return: None
    """
    app['redis'].close()

if __name__ == '__main__':
    application = create_app()
    web.run_app(application, host='0.0.0.0', port=int(getenv('SERVICE_PORT', '2113')))
