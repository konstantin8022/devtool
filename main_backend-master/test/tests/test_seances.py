import datetime
import hamcrest as hc
import pytest
import requests


def test_get_seances(create_movie, create_seance, create_another_seance, create_seats, create_another_movie):
    seance_datetime = "2020-01-18T00:00:00.000Z"
    seance = create_seance(create_movie, seance_datetime)
    another_seance_datetime = datetime.datetime.now().strftime('%Y-%m-%dT%H:%M:%S') + '.000Z'
    another_seance = create_another_seance(create_movie, another_seance_datetime)
    response = requests.get(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies/{create_movie}/seances')
    hc.assert_that(response.status_code == 200)
    response_data = response.json()
    hc.assert_that('data' in response_data)
    seances = response_data['data']
    hc.assert_that(seances, hc.has_length(2))
    seats = [{'id': seat_id, 'vacant': True} for seat_id in create_seats]
    seance_data = {"id": seance,
                   "price": 250,
                   "datetime": seance_datetime,
                   "seats": seats}
    another_seance_data = {"id": another_seance,
                           "price": 250,
                           "datetime": another_seance_datetime,
                           "seats": seats
                           }
    hc.assert_that(seances, hc.has_items(seance_data, another_seance_data))

    response = requests.get(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies/{create_movie}/seances?max_results=1')
    hc.assert_that(response.status_code == 200)
    response_data = response.json()
    hc.assert_that('data' in response_data)
    seances = response_data['data']
    hc.assert_that(seances, hc.has_length(1))
    hc.assert_that(seances, hc.has_items(seance_data if max(seance, another_seance) == seance else another_seance_data))

    response = requests.get(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies?with_seances=1')
    hc.assert_that(response.status_code == 200)
    response_data = response.json()
    hc.assert_that('data' in response_data)
    movies = response_data['data']
    hc.assert_that(movies, hc.has_length(1))
    hc.assert_that(movies, hc.has_items(
        {"id": create_movie,
         "name": "Test Movie",
         "description": "Test description",
         "image_url": "http://example.com/test_image.jpeg",
         "comingSoon": True}
    ))


def test_seances_delete_negative():
    response = requests.delete(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies/1/seances/23232323')
    hc.assert_that(response.status_code == 404)
    response_data = response.json()
    hc.assert_that(response_data, hc.has_entries({'errors': [
        {'title': 'Not found',
         'detail': 'seance_id 23232323 not found'}
    ]}))


def test_seances_get_by_movie_negative():
    response = requests.get(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies/4535435576/seances/')
    hc.assert_that(response.status_code == 200)
    response_data = response.json()
    hc.assert_that(response_data['data'], hc.has_length(0))


def test_movies_validation_422_negative():
    response = requests.post(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies/1/seances', json={
        "datetime": 'qweqwewqe',
        "price": 234,
    })
    hc.assert_that(response.status_code == 422)
    response_data = response.json()
    hc.assert_that(response_data, hc.has_entries({'errors': [
        {'title': 'Validation failed',
         'detail': "'qweqwewqe' is not a 'datetimez'"}
    ]}))


def test_movies_validation_400_negative():
    response = requests.post(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies/1/seances/', json={
        "datetime": "2020-01-18T00:00:00.000Z",
    })
    hc.assert_that(response.status_code == 400)
    response_data = response.json()
    hc.assert_that(response_data, hc.has_entries({'errors': [
        {'title': 'Validation failed',
         'detail': "'price' is a required property"}
    ]}))

