import threading
import hamcrest as hc
import pytest
import requests


def test_get_movies(create_movie, create_another_movie):
    for _ in range(5):
        thread = threading.Thread(target=requests.get,
                                  args=(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies',))
        thread.start()
    response = requests.get(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies')
    hc.assert_that(response.status_code == 200)
    response_data = response.json()
    hc.assert_that('data' in response_data)
    movies = response_data['data']
    hc.assert_that(movies, hc.has_length(2))
    hc.assert_that(movies, hc.has_items(
        {"id": create_movie,
         "name": "Test Movie",
         "description": "Test description",
         "image_url": "http://example.com/test_image.jpeg",
         "comingSoon": False},
        {"id": create_another_movie,
         "name": "Test Movie",
         "description": "Test description",
         "image_url": "http://example.com/test_image.jpeg",
         "comingSoon": False}
    ))

    response = requests.get(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies/?max_results=1')
    hc.assert_that(response.status_code == 200)
    response_data = response.json()
    hc.assert_that('data' in response_data)
    movies = response_data['data']
    hc.assert_that(movies, hc.has_length(1))
    hc.assert_that(movies, hc.has_items(
        {"id": max(create_movie, create_another_movie),
         "name": "Test Movie",
         "description": "Test description",
         "image_url": "http://example.com/test_image.jpeg",
         "comingSoon": False}
    ))


def test_movies_delete_negative():
    response = requests.delete(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies/12321321')
    hc.assert_that(response.status_code == 404)
    response_data = response.json()
    hc.assert_that(response_data, hc.has_entries({'errors': [
        {'title': 'Not found',
         'detail': 'movie_id 12321321 not found'}
    ]}))


def test_movies_validation_422_negative():
    response = requests.post(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies', json={
        "name": [2323],
        "description": "Test description",
        "image_url": "http://example.com/test_image.jpeg"
    })
    hc.assert_that(response.status_code == 422)
    response_data = response.json()
    hc.assert_that(response_data, hc.has_entries({'errors': [
        {'title': 'Validation failed',
         'detail': "[2323] is not of type 'string'"}
    ]}))


def test_movies_validation_400_negative():
    response = requests.post(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies', json={
        "name": "Test",
        "image_url": "http://example.com/test_image.jpeg"
    })
    hc.assert_that(response.status_code == 400)
    response_data = response.json()
    hc.assert_that(response_data, hc.has_entries({'errors': [
        {'title': 'Validation failed',
         'detail': "'description' is a required property"}
    ]}))
