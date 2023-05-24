""" Управление инициализацией БД """
from os import getenv
from datetime import datetime, timedelta
from pymysql.err import InternalError


async def safe_query(cur, query) -> None:
    """
    Проверка существования индекса
    :param cur: объект курсора к БД
    :param query: запрос
    :return: None
    """
    try:
        await cur.execute(query)
    except InternalError as err:
        if err.args[0] in (1061, 1826):
            return
        raise err


async def is_table_exist(cur, table_name: str) -> bool:
    """
    Проверка существования таблицы
    :param cur: объект курсора к БД
    :param table_name: имя таблицы, существование которой принимается
    :return: признак существования таблицы
    """
    await cur.execute(R'''
        SELECT count(*)
        FROM information_schema.TABLES
        WHERE TABLE_SCHEMA = %s
          AND TABLE_NAME = %s
    ''', (getenv('DB_NAME', 'provider_development'), table_name))
    return bool((await cur.fetchone())[0])


async def create_seats(db_conn):
    """
    Создание таблицы с сеансами
    :param db_conn: объект соединения с БД
    :return: None
    """
    async with db_conn.cursor() as cursor:
        if not await is_table_exist(cursor, 'seats'):
            await cursor.execute(R'''
                CREATE TABLE IF NOT EXISTS seats
                (
                    id BIGINT NOT NULL AUTO_INCREMENT,
                    CONSTRAINT movies_pk PRIMARY KEY (id)
                ) ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4;
            ''')
            for seat_id in range(11):
                await cursor.execute('INSERT IGNORE INTO seats SET id = %s', seat_id)
                await db_conn.commit()


async def create_cities(db_conn):
    """
    Создание таблицы с городами
    :param db_conn: объект соединения с БД
    :return: None
    """
    async with db_conn.cursor() as cursor:
        if not await is_table_exist(cursor, 'cities'):
            await cursor.execute(R'''
                CREATE TABLE IF NOT EXISTS cities
                (
                    id BIGINT NOT NULL AUTO_INCREMENT,
                    name VARCHAR(255),
                    CONSTRAINT cities_pk PRIMARY KEY (id)
                ) ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4;
            ''')
            await cursor.execute(f"INSERT INTO cities (name) VALUES ('{getenv('ST_NAMESPACE', 'Irkutsk')}')")
            await db_conn.commit()


async def create_users(db_conn):
    """
    Создание таблицы с пользователями
    :param db_conn: объект соединения с БД
    :return: None
    """
    async with db_conn.cursor() as cursor:
        if not await is_table_exist(cursor, 'users'):
            await cursor.execute(R'''
                CREATE TABLE IF NOT EXISTS users
                (
                    id BIGINT NOT NULL AUTO_INCREMENT,
                    email VARCHAR(255) NOT NULL,
                    CONSTRAINT movies_pk PRIMARY KEY (id)
                ) ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4;
            ''')
            await cursor.execute("INSERT INTO users (email) VALUES ('admin@admin.ru')")
            await db_conn.commit()


async def create_movies(db_conn):
    """
    Создание таблиц с фильмами и сеансами
    :param db_conn: объект соединения с БД
    :return: None
    """
    async with db_conn.cursor() as cursor:
        if not await is_table_exist(cursor, 'movies'):
            await cursor.execute(R'''
                CREATE TABLE IF NOT EXISTS movies
                (
                    id BIGINT NOT NULL AUTO_INCREMENT,
                    name VARCHAR(255) NOT NULL,
                    description TEXT NOT NULL,
                    image_url VARCHAR(255) NOT NULL,
                    CONSTRAINT movies_pk PRIMARY KEY (id)
                ) ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4;
                
                CREATE TABLE IF NOT EXISTS seances
                (
                    id BIGINT NOT NULL AUTO_INCREMENT,
                    movie_id BIGINT NOT NULL,
                    price DECIMAL NOT NULL,
                    seance_datetime DATETIME NOT NULL,
                    CONSTRAINT movies_pk PRIMARY KEY (id)
                ) ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4;
            ''')

            await safe_query(cursor, R'''
                ALTER TABLE seances
                    ADD CONSTRAINT seances_movies_movie_id_fk
                        FOREIGN KEY (movie_id) REFERENCES movies (id)
                    ON DELETE CASCADE
                    ON UPDATE CASCADE;
            ''')

            movies = [
                {'description': 'It\'s Not a Burden', 'name': 'It\'s Not a Burden',
                 'image_url': 'https://m.media-amazon.com/images/M/MV5BMWRkNmI5NDgtZjI4ZC00NTUxLThhMjUtZTZjMzY5ZDBiY2Y1XkEyXkFqcGdeQXVyMzY3MTA2OQ@@._V1_UX182_CR0,0,182,268_AL_.jpg'},
                # pylint:disable=line-too-long
                {'description': 'Seien ranbu', 'name': 'Seien ranbu',
                 'image_url': 'https://m.media-amazon.com/images/M/MV5BYTI3NThmNzctOTM5MS00MDJhLTkxNDMtMDVmODI4NDhjMDNmXkEyXkFqcGdeQXVyNDI5NTI4Mjc@._V1_UY268_CR98,0,182,268_AL_.jpg'},
                # pylint:disable=line-too-long
                {'description': 'Datsukî no ohyaku', 'name': 'Datsukî no ohyaku',
                 'image_url': 'https://m.media-amazon.com/images/M/MV5BZjk4ZThlZjAtOWY4NS00NmIyLWE0MzItOWMxMWUyZDNkOWRmXkEyXkFqcGdeQXVyNzEyMDU1MzU@._V1_UY268_CR7,0,182,268_AL_.jpg'},
                # pylint:disable=line-too-long
                {'description': 'Luis de Llano Palmer', 'name': 'Luis de Llano Palmer',
                 'image_url': 'https://m.media-amazon.com/images/M/MV5BZGFjYzZlYzMtOGE5Ni00NzYyLWFlNTItZDI3NjY3YjUyODAwXkEyXkFqcGdeQXVyMTAxMDQ0ODk@._V1_UY268_CR9,0,182,268_AL_.jpg'},
                # pylint:disable=line-too-long
            ]
            seance_date = datetime.now() + timedelta(days=1)
            seances = [
                {'datetime': seance_date, 'price': 100},
                {'datetime': seance_date, 'price': 200},
                {'datetime': seance_date, 'price': 300},
                {'datetime': seance_date, 'price': 200}
            ]
            for movie in movies:
                await cursor.execute(R'''
                                        INSERT INTO movies (name, description, image_url)
                                        VALUES (%(name)s, %(description)s, %(image_url)s)
                                        ''', movie)
                movie_id = cursor.lastrowid
                for seance in seances:
                    await cursor.execute(fR'''
                                        INSERT INTO seances (movie_id, seance_datetime, price)
                                        VALUES ({movie_id}, %(datetime)s, %(price)s)
                                            ''', seance)
            await db_conn.commit()


async def create_bookings(db_conn):
    """
    Создание таблицы с бронированиями
    :param db_conn: объект соединения с БД
    :return: None
    """
    async with db_conn.cursor() as cursor:
        if not await is_table_exist(cursor, 'bookings'):
            await cursor.execute(R'''
                CREATE TABLE IF NOT EXISTS bookings
                (
                    id BIGINT NOT NULL AUTO_INCREMENT,
                    seance_id BIGINT NOT NULL,
                    seat_id BIGINT NOT NULL,
                    user_id BIGINT NOT NULL,
                    CONSTRAINT bookings_pk PRIMARY KEY (id)
                ) ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4;
            ''')
            await safe_query(cursor, R'''
                ALTER TABLE bookings
                    ADD CONSTRAINT bookings_seances_seance_id_fk
                        FOREIGN KEY (seance_id) REFERENCES seances (id)
                    ON DELETE CASCADE
                    ON UPDATE CASCADE;
                ''')

            await safe_query(cursor, '''
                ALTER TABLE bookings
                    ADD CONSTRAINT bookings_seats_seat_id_fk
                        FOREIGN KEY (seat_id) REFERENCES seats (id)
                    ON DELETE CASCADE
                    ON UPDATE CASCADE;
               ''')

            await safe_query(cursor, R'''
                ALTER TABLE bookings
                    ADD CONSTRAINT bookings_users_user_id_fk
                        FOREIGN KEY (user_id) REFERENCES users (id)
                    ON DELETE CASCADE
                    ON UPDATE CASCADE;
                ''')

            await safe_query(cursor, R'''
                CREATE UNIQUE INDEX bookings_seance_id_seat_id_uindex
                    ON bookings (seance_id, seat_id);
                ''')
            await db_conn.commit()


async def init_db(db_conn) -> None:
    """
    Инкапсуляция инициализации БД и миграций
    :param db_conn: объект пула соединений с БД
    :return: None
    """
    await create_seats(db_conn)
    await create_users(db_conn)
    await create_cities(db_conn)
    await create_movies(db_conn)
    await create_bookings(db_conn)
