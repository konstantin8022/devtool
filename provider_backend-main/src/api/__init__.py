""" Индексация API """
from src.api.bookings import Bookings
from src.api.middlewares import get_auth_header_middleware, error_middleware, prometheus_middleware
from src.api.movies import Movies
