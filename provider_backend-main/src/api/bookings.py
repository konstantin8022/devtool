""" Bookings API """
import re
from asyncio import shield
from aiohttp import web
from aiomysql import DictCursor
from pymysql.err import IntegrityError
from src.misc import add_optionally_slashed_route, validatable
from os import getenv


class Bookings(web.Application):
    """
    API взаимодействия с сущностью "Бронирование"
    """
    seat_err_re = re.compile(R"^Duplicate entry '\d+-(\d+)'.*", re.MULTILINE)

    def __init__(self):
        super().__init__()
        self['default_headers'] = {'charset': 'utf-8'}
        self.add_routes([*add_optionally_slashed_route(web.post, '', create_booking_directly)])

    def post_init(self, parent) -> None:
        """
        Инициализация асинхронно создаваемых объектов
        :param parent: родительское приложение aiohttp
        :return: None
        """
        self['primary_db_conn_pool'] = parent['primary_db_conn_pool']
        self['secondary_db_conn_pool'] = parent['secondary_db_conn_pool']
        self['schemas'] = parent['schemas']
        self['success_bookings_total'] = parent['success_bookings_total']


@validatable()
async def create_booking_directly(request):
    """
    Создание бронирования напрямоую (/bookings)
    :param request: объект входящего запроса aiohttp
    :return: http response
    """
    return await shield(create_booking(request.app, await request.json(), request['additional_headers']))


async def create_booking(app, payload, additional_headers):
    """
    Создание записи с сеансом
    :param app: объект приложения aiohttp
    :param payload: payload запроса
    :param additional_headers: дополнительные хидеры (авторизация)
    :return: http response
    """
    seats = payload.get('seatsIds')
    if not seats:
        return web.json_response({'errors': [{'title': "seatsIds must be an non-empty array"}]},
                                 status=400)

    async with app['primary_db_conn_pool'].acquire() as connection:
        async with app['secondary_db_conn_pool'].acquire() as secondary_connection:
            async with secondary_connection.cursor(DictCursor) as cursor:
                await shield(cursor.execute(R'''
                    SELECT id FROM users WHERE email = %(email)s
                    ''', payload))
                await shield(secondary_connection.commit())
                user = await shield(cursor.fetchone())
            async with connection.cursor(DictCursor) as cursor:
                if user is None:
                    await shield(cursor.execute(R'''
                    INSERT INTO users (email) VALUES (%(email)s)
                    ''', payload))
                    user_id = cursor.lastrowid
                else:
                    user_id = user['id']
                seance_id = payload['seance_id']
                bookings_id = []
                try:
                    for seat_id in seats:
                        await shield(cursor.execute(R'''
                            INSERT INTO bookings (seance_id, seat_id, user_id) VALUES (%s, %s, %s)
                        ''', (seance_id, seat_id, user_id)))
                        bookings_id.append(cursor.lastrowid)
                    await shield(connection.commit())
                except IntegrityError as err:
                    await shield(connection.rollback())
                    if err.args[0] == 1062:
                        taken_seat_id = re.search(Bookings.seat_err_re, err.args[1]).group(1)
                        return web.json_response({'errors': [{'title': 'Seat already taken',
                                                              'detail': f'Taken seats: {taken_seat_id}'}]},
                                                 status=409)
                    if err.args[0] == 1452:
                        return web.json_response(
                            {'errors': [{'title': 'Not found',
                                         'detail': f'seance_id {seance_id} not found'}]},
                            status=404)
                    return web.json_response({'errors': [{'title': err.args[1]}]}, status=400)

        city = getenv('PROVIDER_CITY', 'unknown')
        app['success_bookings_total'].add({'app': 'provider_backend', 'city': city}, len(bookings_id))
        #
        # PLACEHOLDER
        #

        return web.json_response({'data': bookings_id},
                                 headers={**app['default_headers'], **additional_headers})
