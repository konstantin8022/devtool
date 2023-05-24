import json
import pytest
import redis


@pytest.fixture(scope='session', autouse=True)
def conf_redis():
    redis_con = redis.Redis(host='localhost', password='c0WcWBm2kZjN0ivN')
    redis_con.hmset('list-of-teams', {pytest.provider_name: json.dumps({"name": pytest.provider_name,
                                                             "namespace": "default",
                                                             "providerURL": "http://provider:2122"})})
    redis_con.hmset(f'authproblems-{pytest.provider_name}', {
        'slowdown-probability': 0,
        'slowdown-min': 0,
        'slowdown-max': 0,
        'error-probability': 0,
        'active': 'true'
    })
    redis_con.hmset('authproblems-default', {
        'slowdown-probability': 0,
        'slowdown-min': 0,
        'slowdown-max': 0,
        'error-probability': 0,
        'active': 'true'
    })
    return redis_con
