# Movies API

## Success response

### GET /cities/moscow/movies
### Request

#### Headers

<pre>Content-Type: application/json
Host: example.org
Cookie: </pre>

#### Route

<pre>GET /cities/moscow/movies</pre>

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
      "title": "Movie name one",
      "comingSoon": true,
      "image": "http://example.com/movies/2.jpg"
    },
    {
      "id": 2,
      "title": "Movie name two",
      "comingSoon": true,
      "image": "http://example.com/movies/2.jpg"
    }
  ]
}</pre>
