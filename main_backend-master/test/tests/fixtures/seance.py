import hamcrest as hc
import mysql.connector
import pytest
import requests


@pytest.fixture(scope='function')
def create_seance(request: pytest.fixture, create_seats):
    def impl(movie_id, datetime_val: str):
        payload = {"datetime": datetime_val,
                   "price": 250}
        seats = [{'id': seat_id, 'vacant': True} for seat_id in create_seats]
        response = requests.post(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies/{movie_id}/seances', json=payload)
        hc.assert_that(response.status_code == 200)
        response_data = response.json()
        hc.assert_that('id' in response_data['data'])
        seance_id = response_data['data']['id']
        hc.assert_that(response_data, hc.has_entries({"data": hc.has_entries({
            "type": "seances",
            "attributes": hc.has_entries(payload),
            "seats": hc.has_items(*seats)
        })}))
        cnx = mysql.connector.connect(**pytest.mysql_credentials)
        cursor = cnx.cursor(dictionary=True)
        cursor.execute(R'''
            SELECT * FROM seances WHERE id = %(id)s  
        ''', {'id': seance_id})
        payload['id'] = seance_id
        payload['movie_id'] = movie_id
        data = cursor.fetchone()
        data['datetime'] = data['seance_datetime'].strftime('%Y-%m-%dT%H:%M:%S.%f')[:-3] + 'Z'
        data['price'] = int(data['price'])
        del data['seance_datetime']
        hc.assert_that(payload == data)
        cnx.commit()
        cursor.close()
        cnx.close()
        request.addfinalizer(lambda: delete_seance(movie_id, seance_id))
        return seance_id
    return impl


def delete_seance(movie_id, seance_id):
    response = requests.delete(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies/{movie_id}/seances/{seance_id}/')
    hc.assert_that(response.status_code == 200)
    response_data = response.json()
    hc.assert_that(response_data, hc.has_entries({"data": hc.has_entries({
        "id": str(seance_id),
        "type": "seances",
    })}))
    cnx = mysql.connector.connect(**pytest.mysql_credentials)
    cursor = cnx.cursor(dictionary=True)
    cursor.execute(R'''
        SELECT * FROM seances WHERE id = %(id)s  
    ''', params={'id': seance_id})
    data = cursor.fetchone()
    hc.assert_that(data is None)
    cursor.close()
    cnx.close()

create_another_seance = create_seance
