import hamcrest as hc
import pytest
import requests


def test_main_backend_health():
    check_metrics_endpoint(pytest.main_backend_url)


def test_provider_health():
    check_metrics_endpoint(pytest.provider_backend_url)


def test_auth_service_health():
    check_metrics_endpoint(pytest.auth_service_url)


def check_metrics_endpoint(service_url):
    response = requests.get(f'{service_url}/metrics')
    hc.assert_that(response.status_code == 200)
