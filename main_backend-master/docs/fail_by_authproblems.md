### Any handler from provider API with prefix /cities/:provider_name (i.e. /cities/:provider_name/movies/:movie_id/seances) 
### Request

#### Headers

<pre>Content-Type: application/json</pre>

#### Route

<pre>POST /cities/provider_name/movies/1/seances</pre>

#### Body

<pre>{"datetime":"2020-01-18T00:00:00.000Z"}</pre>

### Response

#### Headers

<pre>Content-Type: application/json; charset=utf-8</pre>

#### Status

<pre>500 Internal Server Error</pre>

#### Body

<pre>{"errors": [{"title": "Auth service: []"}]}</pre>
