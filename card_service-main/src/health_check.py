# Copyright 2020 gRPC authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import grpc
import logging
import traceback

from aiohttp import web
from src import purchase_pb2
from src import purchase_pb2_grpc
from os import getenv
import sys

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

    app.add_routes([web.get('/health', healthcheck)])

if __name__ == '__main__':
    application = create_app()
    web.run_app(application, host='0.0.0.0', port=int(getenv('HEALTHCHECK_PORT', '2113')))
