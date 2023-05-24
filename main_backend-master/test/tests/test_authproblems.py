import time
import hamcrest as hc
import pytest
import requests


def test_slowdown(conf_redis):
    conf_redis.hmset(f'authproblems-{pytest.provider_name}', {
        'slowdown-probability': 100,
        'slowdown-min': 3000,
        'slowdown-max': 3000,
        'error-probability': 0,
        'active': 'true'
    })
    start = time.time()
    response = requests.get(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies')
    slowdown = time.time() - start
    hc.assert_that(response.status_code == 200)
    hc.assert_that(3 < slowdown < 4)


def test_error(conf_redis):
    conf_redis.hmset(f'authproblems-{pytest.provider_name}', {
        'slowdown-probability': 0,
        'slowdown-min': 0,
        'slowdown-max': 0,
        'error-probability': 1000,
        'active': 'true'
    })
    response = requests.get(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies')
    hc.assert_that(response.status_code == 500)
    hc.assert_that(response.json(), hc.has_entries({'errors': [
        {"title": "Auth service: []"}
    ]}))


def test_default_conf(conf_redis):
    conf_redis.hmset(f'authproblems-{pytest.provider_name}', {
        'slowdown-probability': 0,
        'slowdown-min': 0,
        'slowdown-max': 0,
        'error-probability': 1000,
        'active': 'false'
    })
    response = requests.get(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies')
    hc.assert_that(response.status_code == 200)
