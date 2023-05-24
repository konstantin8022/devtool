# Bookings API

## Failed response

### POST /cities/moscow/movies/1/seances/1/bookings
### Request

#### Headers

<pre>Content-Type: application/json
Host: example.org
Cookie: </pre>

#### Route

<pre>POST /cities/moscow/movies/1/seances/1/bookings</pre>

#### Body

<pre>{}</pre>

### Response

#### Headers

<pre>Content-Type: application/json; charset=utf-8</pre>

#### Status

<pre>503 Service Unavailable</pre>

#### Body

<pre>{
  "errors": [
    {
      "title": "Provider Москва: Exception",
      "detail": "detail"
    }
  ]
}</pre>
