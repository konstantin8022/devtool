import mysql.connector
import pytest


@pytest.fixture(scope='session')
def create_seats(request):
    cnx = mysql.connector.connect(**pytest.mysql_credentials)
    cursor = cnx.cursor()
    seats = []
    for _ in range(2):
        cursor.execute(R'INSERT INTO seats VALUES (DEFAULT)')
        seats.append(cursor.lastrowid)
    cnx.commit()
    cursor.close()
    cnx.close()
    request.addfinalizer(remove_seats)
    return seats


def remove_seats():
    cnx = mysql.connector.connect(**pytest.mysql_credentials)
    cursor = cnx.cursor()
    for _ in range(2):
        cursor.execute(R'DELETE FROM seats')
    cnx.commit()
    cursor.close()
    cnx.close()