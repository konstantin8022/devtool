""" Инициализация приложения """
from json import loads
import logging
from os import listdir, getenv
from os.path import isfile, join
import sys
import aiomysql
from aiohttp import web, ClientSession
from aiohttp.hdrs import ACCEPT
from aioprometheus import Counter, Histogram, Registry, render
from src.api import Bookings, Movies, get_auth_header_middleware, error_middleware, prometheus_middleware
from src.database import init_db


routes = web.RouteTableDef()


@routes.get('/health')
async def healthcheck(request):
    """
    Хелзчек, который всегда 200
    :param request: объект с параметрами входящего запроса
    :return: 200 OK
    """
    return web.json_response({'result': 'ok'})


@routes.get('/metrics')
async def handle_metrics(request):
    """
    Отправка метрик в prometheus
    """
    content, http_headers = render(request.app['prometheus_registry'], request.headers.getall(ACCEPT, []))
    return web.Response(body=content, headers=http_headers)


def create_app() -> web.Application:
    """
    Создание прложения и настройка модулей
    :return: объект приложения
    """
    app = web.Application(middlewares=[prometheus_middleware, error_middleware, get_auth_header_middleware])
    movies_module = Movies()
    bookings_module = Bookings()
    app['modules'] = [movies_module, bookings_module]
    app.add_subapp('/movies/', movies_module)
    app.add_subapp('/bookings/', bookings_module)
    app.on_startup.append(on_startup)
    app.on_shutdown.append(on_shutdown)
    app.add_routes(routes)
    return app


def read_schemas() -> tuple:
    """
    Загрузка json-схем в память
    :return: кортеж: имя схемы, содержимое
    """
    path = 'src/etc/schema/'
    for raw_file in listdir(path):
        if isfile(file_path := join(path, raw_file)):
            with open(file_path) as schema_file:
                yield raw_file.rsplit('.', 1)[0], loads(schema_file.read())


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
    app['primary_db_conn_pool'] = await aiomysql.create_pool(maxsize=int(getenv('DB_CONN_POOL_SIZE', '50')),
                                                            host=getenv('DB_HOST'), port=int(getenv('DB_PORT')),
                                                            user=getenv('DB_USER'), password=getenv('DB_PASS'),
                                                            db=getenv('DB_NAME'))
    app['secondary_db_conn_pool'] = await aiomysql.create_pool(maxsize=int(getenv('RDB_CONN_POOL_SIZE', '50')),
                                                           host=getenv('RDB_HOST'), port=int(getenv('RDB_PORT')),
                                                           user=getenv('RDB_USER'), password=getenv('RDB_PASS'),
                                                           db=getenv('RDB_NAME'))
    async with app['primary_db_conn_pool'].acquire() as conn:
        await init_db(conn)
    app['schemas'] = dict(read_schemas())
    app['client_session'] = ClientSession()
    prometheus_registry = Registry()
    app['success_bookings_total'] = Counter('success_bookings_total', 'The total number of successfull bookings')
    prometheus_registry.register(app['success_bookings_total'])
    app['requests_total'] = Counter('http_server_requests_total',
                                            'The total number of HTTP requests handled by the application')
    app['request_duration_seconds'] = Histogram('http_server_request_duration_seconds',
                   'The HTTP response duration',
                   buckets=[0.005, 0.01, 0.025, 0.05, 0.1, 0.3, 0.5, 0.7, 0.9, 1, 1.1, 1.3, 1.5, 1.7, 1.9, 2.0, 2.5, 5, 10])
    prometheus_registry.register(app['requests_total'])
    prometheus_registry.register(app['request_duration_seconds'])
    app['prometheus_registry'] = prometheus_registry
    for module in app['modules']:
        module.post_init(app)


async def on_shutdown(app) -> None:
    """
    Колбек завершения работы приложения
    :param app: объект сервера aiohttp
    :return: None
    """
    app['primary_db_conn_pool'].close()
    await app['primary_db_conn_pool'].wait_closed()
    app['secondary_db_conn_pool'].close()
    await app['secondary_db_conn_pool'].wait_closed()

if __name__ == '__main__':
    application = create_app()
    web.run_app(application, host='0.0.0.0', port=int(getenv('SERVICE_PORT', '2112')))
