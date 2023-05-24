""" App middlewares """
from asyncio import TimeoutError
from datetime import datetime
from os import getenv
from traceback import print_exc
from aiohttp import web, client_exceptions, ClientTimeout
import logging

STATUS_MESSAGES = {
    404: 'request path not found',
    500: 'internal server error',
    504: 'http gateway timeout',
}


@web.middleware
async def prometheus_middleware(request: web.Request, handler):
    """
    Отправка метрик в Prometheus
    :param request: объект входящего запроса
    :param handler: функция-обработчик запроса
    :return: объект ответа
    """
    start = datetime.now()
    response = await handler(request)
    end = datetime.now() - start
    city = getenv('PROVIDER_CITY', 'unknown')
    prometheus_labels = {'app': 'provider_backend', 'city': city}
    request.app['request_duration_seconds'].observe(prometheus_labels, end.total_seconds())
    request.app['requests_total'].inc({**prometheus_labels, 'code': response.status})
    return response


@web.middleware
async def error_middleware(request: web.Request, handler):
    """
    Обработка ответов на запросы, завершившихся ошибками
    :param request: объект входящего запроса
    :param handler: функция-обработчик запроса
    :return: объект ответа
    """
    try:
        response = await handler(request)
        return response
    except web.HTTPException as err:
        status = err.status
        if status >= 500:
            request.app['logger'].error('Failure during HTTP request', exc_info=True)
        return web.json_response({'errors': [{'app': 'provider_backend', 'title': STATUS_MESSAGES.get(status, err.reason)}]},
                                 status=status)
    except client_exceptions.ClientConnectionError:
        app['logger'].logger.error('Failed to connect', exc_info=True)
        return web.json_response({'errors': [{'app': 'provider_backend', 'title': f'Failed to connect'}]},
                                 status=503)

    except TimeoutError as err:
        request.app['logger'].error('Request timeout exceeded', exc_info=True)
        return web.json_response({'errors': [{'app': 'provider_backend', 'title': STATUS_MESSAGES.get(504, "http gateway timeout")}]},
                                 status=504)
    except Exception as err:
        request.app['logger'].error('Something went wrong', exc_info=True)
        return web.json_response({'errors': [{'app': 'provider_backend', 'title': str(err)}]}, status=500)


@web.middleware
async def get_auth_header_middleware(request: web.Request, handler):
    """
    Авторизация в сервисе авторизации :)
    :param request: объект входящего запроса
    :param handler: функция-обработчик запроса
    :return: ответ на входящий запрос или ответ с ошибкой
    """
    status = 0
    auth_response = None
    reason = ""
    if (path := request.path) in ['/metrics', '/health']:
        return await handler(request)

    tm = float(getenv('AUTH_SERVICE_TIMEOUT', '0'))
    timeout = tm if tm > 0 else 60

    #
    # PLACEHOLDER
    #

    tracing_need_headers = ['x-request-id', 'x-b3-traceid', 'x-b3-spanid', 'x-b3-parentspanid', 'x-b3-sampled', 'x-b3-flags']
    tracing_headers = {}

    for k,v in request.headers.items():
        if k in tracing_need_headers:
            tracing_headers[k] = v

    headers = {**{getenv('PROVIDER_SOURCE_HEADER'): getenv('PROVIDER_SOURCE_TOKEN')}, **tracing_headers}

    request.app['logger'].debug(f'Send request to auth_service with headers {headers}')

    auth_response = await request.app['client_session'] \
        .get(getenv('AUTH_SERVICE_URL'),
             headers=headers,
             timeout=ClientTimeout(total=timeout))

    status = auth_response.status
    if status >= 200 and status < 300:
        request['additional_headers'] = {'X-Auth-Operation-Id':
                                         auth_response.headers.getone('x-auth-operation-id')}
        response = await handler(request)
        return response

    if auth_response is not None:
        reason = (await auth_response.json()) if status == 401 else (await auth_response.text())

    return web.json_response({'errors': [{'title': f"Auth service", 'detail': reason}]}, status=status)
