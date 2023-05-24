from datetime import datetime
import json
from os import getenv
import re
import jwt
from aiohttp import web
from traceback import format_exc
import grpc
from src import purchase_pb2
from src import purchase_pb2_grpc


PATH_ID_RE = re.compile(R'\d+')

def card_service_middleware(decorated):
    """ Миддлваря для запроса в сервис проверки пластиковой карты"""
    async def wrapper(request: web.Request):
        try:
            if request.method == 'POST' and request.path.find('bookings'):
                card = json.loads(await request.read()).pop('card')
                if card == "":
                    return web.json_response({'errors': [{'title': "Parameter card is empty"}]},status=422)

                try:
                    await call_card_service(card, request)
                    response = await decorated(request)
                except Exception:
                    request.app['logger'].error('Card service error', exc_info=True)
                    response = web.json_response({'errors': [{'title': "Unexpected error"}]},status=500)

            else:
                response = await decorated(request)
        except KeyError:
            response = web.json_response({'errors': [{'title': "Parameter card not present"}]},status=422)
        except json.decoder.JSONDecodeError:
            response = web.json_response({'errors': [{'title': "Invalid json payload"}]},status=400)

        return response
    return wrapper

async def call_card_service(card, request):
    """
    Запрос в сервис пластиковых карт
    """
    metadata = []
    tracing_headers = await get_tracing_headers(request)

    for k, v in tracing_headers.items():
        metadata.append((k, v))
    metadata.append(('x-envoy-retry-grpc-on', 'internal'))

    request.app['logger'].debug(f'Send grpc request to card service with metadata {metadata}')
    async with grpc.aio.insecure_channel(getenv('CARD_SERVICE_URL')) as channel:
        stub = purchase_pb2_grpc.GreeterStub(channel)
        response = await stub.Purchase(purchase_pb2.PurchaseRequest(card=card), metadata=metadata)
        return str(response)

async def get_tracing_headers(request):
    tracing_need_headers = ['x-request-id', 'x-b3-traceid', 'x-b3-spanid', 'x-b3-parentspanid', 'x-b3-sampled', 'x-b3-flags']

    tracing_headers = {}

    for k,v in request.headers.items():
        if k in tracing_need_headers:
            tracing_headers[k] = v

    return tracing_headers
