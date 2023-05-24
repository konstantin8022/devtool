import pytest


pytest_plugins = [
    "test.tests.fixtures.booking",
    "test.tests.fixtures.movie",
    "test.tests.fixtures.seance",
    "test.tests.fixtures.seats",
    "test.tests.fixtures.redis",
]


def pytest_configure():
    pytest.mysql_credentials = {'user': 'provider_user',
                                'password': 'Gj7BDvmL8SD',
                                'host': '127.0.0.1',
                                'port': 3306,
                                'database': 'provider_development'}
    pytest.main_backend_url = 'http://localhost:3000'
    pytest.provider_backend_url = 'http://localhost:2122'
    pytest.auth_service_url = 'http://localhost:2111'
    pytest.provider_name = 'voronezh'
