# Seances API

## Success response

### GET /cities/moscow/movies/1/seances
### Request

#### Headers

<pre>Content-Type: application/json
Host: example.org
Cookie: </pre>

#### Route

<pre>GET /cities/moscow/movies/1/seances</pre>

### Response

#### Headers

<pre>Content-Type: application/json; charset=utf-8</pre>

#### Status

<pre>200 OK</pre>

#### Body

<pre>{
  "data": [
    {
      "id": 1,
      "price": 250,
      "datetime": "2020-01-16T00:00:00.000Z",
      "seats": [
        {
          "id": 1,
          "vacant": true
        },
        {
          "id": 2,
          "vacant": true
        }
      ]
    },
    {
      "id": 2,
      "price": 300,
      "datetime": "2020-02-16T00:00:00.000Z",
      "seats": [
        {
          "id": 1,
          "vacant": true
        },
        {
          "id": 2,
          "vacant": true
        }
      ]
    }
  ]
}</pre>
