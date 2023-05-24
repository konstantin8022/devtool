import hamcrest as hc
import pytest
import requests


def test_main_backend_health():
    healthcheck(pytest.main_backend_url)


def test_provider_health():
    healthcheck(pytest.provider_backend_url)


def test_auth_service_health():
    healthcheck(pytest.auth_service_url)


def healthcheck(service_url):
    response = requests.get(f'{service_url}/health')
    hc.assert_that(response.status_code == 200)
    response_data = response.json()
    hc.assert_that(response_data, hc.has_entries({'result': 'ok'}))
