# Seances API

## Failed response

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
