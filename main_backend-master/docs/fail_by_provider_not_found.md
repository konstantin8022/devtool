### Any handler from provider API with prefix /cities/:provider_name 
Raises when provider_name is not in redis hash list-of-teams
### Request

#### Headers

<pre>Content-Type: application/json</pre>

#### Route

<pre>POST /movies/1/seances</pre>

#### Body

<pre>{"datetime":"2020-01-18T00:00:00.000Z"}</pre>

### Response

#### Headers

<pre>Content-Type: application/json; charset=utf-8</pre>

#### Status

<pre>404 Not Found</pre>

#### Body

<pre>{"errors":{"title": "City not found"}}</pre>


