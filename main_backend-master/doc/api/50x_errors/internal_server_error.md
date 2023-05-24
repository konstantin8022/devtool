# 50X Errors API

## Internal server error

### GET cities/moscow/movies
### Request

#### Headers

<pre>Content-Type: application/json
Host: example.org
Cookie: </pre>

#### Route

<pre>GET cities/moscow/movies</pre>

### Response

#### Headers

<pre>Content-Type: application/json; charset=utf-8</pre>

#### Status

<pre>503 Service Unavailable</pre>

#### Body

<pre>{
  "errors": [
    {
      "title": "Main API: Exception",
      "detail": "Some backtrace"
    }
  ]
}</pre>
