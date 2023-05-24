import hamcrest as hc
import pytest
import requests


def get_bookings(create_movie, create_seance, create_seats, create_booking, is_short_url):
    seance_datetime = "2020-01-18T00:00:00.000Z"
    seance = create_seance(create_movie, seance_datetime)
    create_booking('test_email@slurm.io', [create_seats[0]], seance, is_short_url)
    response = requests.get(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies/{create_movie}/seances')
    hc.assert_that(response.status_code == 200)
    response_data = response.json()
    hc.assert_that('data' in response_data)
    seances = response_data['data']
    hc.assert_that(seances, hc.has_length(1))
    seats = [{'id': create_seats[0], 'vacant': False},
             {'id': create_seats[1], 'vacant': True}]
    seance_data = {"id": seance,
                   "price": 250,
                   "datetime": seance_datetime,
                   "seats": seats}
    hc.assert_that(seances, hc.has_items(seance_data))


def test_get_bookings_directly(create_movie, create_seance, create_seats, create_booking):
    get_bookings(create_movie, create_seance, create_seats, create_booking, True)


def test_get_bookings_through_movies(create_movie, create_seance, create_seats, create_booking):
    get_bookings(create_movie, create_seance, create_seats, create_booking, False)


def test_make_booking_empty_negative(create_movie, create_seance):
    seance = create_seance(create_movie, "2020-01-18T00:00:00.000Z")
    response = requests.post(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/bookings/', json={"email": 'test_email@slurm.io',
                                                                                     "seatsIds": [],
                                                                                     "seance_id": seance})
    hc.assert_that(response.status_code == 400)
    response_data = response.json()
    hc.assert_that(response_data, hc.has_entries({'errors': [
        {'title': 'seatsIds must be an non-empty array'}
    ]}))


def test_make_booking_no_exist_seance_negative():
    response = requests.post(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/bookings', json={"email": 'test_email@slurm.io',
                                                                                     "seatsIds": [1],
                                                                                     "seance_id": 123123})
    hc.assert_that(response.status_code == 404)
    response_data = response.json()
    hc.assert_that(response_data, hc.has_entries({'errors': [
        {'title': 'Not found',
         'detail': 'seance_id 123123 not found'}
    ]}))


def test_make_booking_already_taken_negative(create_movie, create_seance, create_seats, create_booking):
    seance_datetime = "2020-01-18T00:00:00.000Z"
    seance = create_seance(create_movie, seance_datetime)
    create_booking('test_email@slurm.io', [create_seats[0]], seance)
    response = requests.post(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/bookings/', json={"email": 'test_email@slurm.io',
                                                                                     "seatsIds": [create_seats[0]],
                                                                                     "seance_id": seance})
    hc.assert_that(response.status_code == 409)
    response_data = response.json()
    hc.assert_that(response_data, hc.has_entries({'errors': [
        {'title': 'Seat already taken',
         'detail': f'Taken seats: {create_seats[0]}'}
    ]}))


def test_make_booking_validation_422_negative():
    response = requests.post(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/bookings', json={"email": 'slurm.io',
                                                                                     "seatsIds": [0],
                                                                                     "seance_id": 1})
    hc.assert_that(response.status_code == 422)
    response_data = response.json()
    hc.assert_that(response_data, hc.has_entries({'errors': [
        {'title': 'Validation failed',
         'detail': "'slurm.io' is not a 'email'"}
    ]}))


def test_make_booking_validation_400_negative():
    response = requests.post(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/bookings/', json={"seatsIds": [0],
                                                                                     "seance_id": 1})
    hc.assert_that(response.status_code == 400)
    response_data = response.json()
    hc.assert_that(response_data, hc.has_entries({'errors': [
        {'title': 'Validation failed',
         'detail': "'email' is a required property"}
    ]}))
