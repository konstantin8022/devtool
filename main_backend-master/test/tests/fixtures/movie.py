import hamcrest as hc
import mysql.connector
import pytest
import requests


@pytest.fixture(scope='function')
def create_movie(request: pytest.fixture):
    payload = {
        "name": "Test Movie",
        "description": "Test description",
        "image_url": "http://example.com/test_image.jpeg"
    }
    response = requests.post(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies/', json=payload)
    hc.assert_that(response.status_code == 200)
    response_data = response.json()
    hc.assert_that('id' in response_data['data'])
    movie_id = response_data['data']['id']
    hc.assert_that(response_data, hc.has_entries({"data": hc.has_entries({
        "type": "movies",
        "attributes": hc.has_entries(payload),
    })}))
    cnx = mysql.connector.connect(**pytest.mysql_credentials)
    cursor = cnx.cursor(dictionary=True)
    cursor.execute(R'''
        SELECT * FROM movies WHERE id = %(id)s  
    ''', {'id': movie_id})
    payload['id'] = movie_id
    data = cursor.fetchone()
    hc.assert_that(payload == data)
    cnx.commit()
    cursor.close()
    cnx.close()
    request.addfinalizer(lambda: delete_movie(movie_id))
    return movie_id


def delete_movie(movie_id):
    response = requests.delete(f'{pytest.main_backend_url}/cities/{pytest.provider_name}/movies/{movie_id}')
    hc.assert_that(response.status_code == 200)
    response_data = response.json()
    hc.assert_that(response_data, hc.has_entries({"data": hc.has_entries({
        "id": str(movie_id),
        "type": "movies",
    })}))
    cnx = mysql.connector.connect(**pytest.mysql_credentials)
    cursor = cnx.cursor(dictionary=True)
    cursor.execute(R'''
        SELECT * FROM movies WHERE id = %(id)s  
    ''', params={'id': movie_id})
    data = cursor.fetchone()
    hc.assert_that(data is None)
    cnx.commit()
    cursor.close()
    cnx.close()

create_another_movie = create_movie
