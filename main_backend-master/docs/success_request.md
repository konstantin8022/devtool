### Any handler from provider API with prefix /cities/:provider_name (i.e. POST /cities/irkutsk/movies/1/seances/1/bookings)
Just proxy request to provider API. Returns success response or not, but request was proxied normally
### Request

#### Headers

<pre>Content-Type: application/json
Host: example.org
Cookie: </pre>

#### Route

<pre>POST /cities/moscow/movies/1/seances/1/bookings</pre>

#### Body

<pre>{"email":"hello@test.com","seatsIds":[1],"seance_id":1}</pre>

### Response

#### Headers

<pre>Content-Type: application/json; charset=utf-8</pre>

#### Status

<pre>200 OK</pre>

#### Body

<pre>{
  "data": [
    1
  ]
}</pre>
