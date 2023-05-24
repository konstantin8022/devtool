import hamcrest as hc
import mysql.connector
import pytest
import requests


@pytest.fixture(scope='function')
def create_booking(request: pytest.fixture, create_seats):
    def impl(user_email: str, seats: list, seance_id: int, is_direct_request=False):
        payload = {"email": user_email,
                   "seatsIds": seats}
        url = f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies/1/seances/{seance_id}/bookings'
        if is_direct_request:
            payload["seance_id"] = seance_id
            url = f'{pytest.main_backend_url}/cities/{pytest.provider_name}/bookings'
        response = requests.post(url, json=payload)
        hc.assert_that(response.status_code == 200)
        response_data = response.json()
        hc.assert_that("data" in response_data)
        bookings = response_data['data']
        hc.assert_that(bookings, hc.has_length(len(seats)))
        cnx = mysql.connector.connect(**pytest.mysql_credentials)
        cursor = cnx.cursor()
        cursor.execute(R'''
            SELECT bookings.id 
            FROM bookings
            JOIN users ON users.id = bookings.user_id 
            WHERE users.email = %(email)s AND seance_id = %(seance_id)s 
        ''', {'email': user_email,
              'seance_id': seance_id})
        data = cursor.fetchall()
        hc.assert_that([tuple(bookings)] == data)
        cnx.commit()
        cursor.close()
        cnx.close()
        request.addfinalizer(lambda: delete_booking(bookings))
        return bookings
    return impl


def delete_booking(bookings):
    cnx = mysql.connector.connect(**pytest.mysql_credentials)
    cursor = cnx.cursor()
    for _ in range(2):
        cursor.execute(R'DELETE FROM bookings WHERE id IN (%s)', (bookings))
    cnx.commit()
    cursor.close()
    cnx.close()
