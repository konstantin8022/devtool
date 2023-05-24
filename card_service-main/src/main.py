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
"""The Python AsyncIO implementation of the GRPC helloworld.Greeter server."""

import logging
import asyncio
import grpc
import urllib.request
import functools
import random
import json
import http.client

from os import getenv
import aioredis

from src import purchase_pb2
from src import purchase_pb2_grpc

class Greeter(purchase_pb2_grpc.GreeterServicer):

    async def Purchase(self, request: purchase_pb2.PurchaseRequest,
                       context: grpc.aio.ServicerContext
                      ) -> purchase_pb2.PurchaseReply:
        if await should_be_error():
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details('Smth happened!')
            return purchase_pb2.Response()

        call_to_paypal()

        return purchase_pb2.PurchaseReply(message='Success, %s!' % request.card)

def call_to_paypal():
    if random.random() > 0.95:
        conn = http.client.HTTPSConnection('paypal.com', 443)
        conn.request('GET', '/')
        conn.getresponse().read()

async def errors_probabolity(redis):
    """
    Получение списка провайдеров из Redis
    :param redis: объект подключения к редису
    :return: список провайдеров
    """
    return {k: json.loads(v) for k, v in (await redis.hgetall('list-of-teams', encoding='utf-8')).items()}

async def should_be_error():
    url = f'{getenv("MAIN_BACKEND_HOST")}/api/v1/config/card_service'
    response = urllib.request.urlopen(url).read()
    rate = json.loads(response)['rate']

    if rate is not None and rate != '':
        return random.randint(0, 100) < float(rate)
    else:
        return false

async def serve() -> None:
    server = grpc.aio.server()
    purchase_pb2_grpc.add_GreeterServicer_to_server(Greeter(), server)
    listen_addr = f'[::]:{getenv("SERVICE_PORT")}'
    server.add_insecure_port(listen_addr)
    logging.info("Starting server on %s", listen_addr)
    await server.start()
    try:
        await server.wait_for_termination()
    except KeyboardInterrupt:
        # Shuts down the server with 0 seconds of grace period. During the
        # grace period, the server won't accept new connections and allow
        # existing RPCs to continue within the grace period.
        await server.stop(0)

if __name__ == '__main__':
    asyncio.run(serve())
