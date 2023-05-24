# Bookings API

## 422 response

### POST /cities/moscow/movies/1/seances/1/bookings
### Request

#### Headers

<pre>Content-Type: application/json
Host: example.org
Cookie: </pre>

#### Route

<pre>POST /cities/moscow/movies/1/seances/1/bookings</pre>

#### Body

<pre>{"email":"wrong_email","seatsIds":[1],"seance_id":1}</pre>

### Response

#### Headers

<pre>Content-Type: application/json; charset=utf-8</pre>

#### Status

<pre>422 Unprocessable Entity</pre>

#### Body

<pre>{
  "errors": [
    {
      "title": "email",
      "detail": "has invalid format"
    }
  ]
}</pre>
