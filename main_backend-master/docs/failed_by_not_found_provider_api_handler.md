### Any handler not found at provider api (i.e. /cities/:provider_name/handler) 
### Request

#### Headers

<pre>Content-Type: application/json</pre>

#### Route

<pre>POST /cities/provider_name/handler</pre>

### Response

#### Headers

<pre>Content-Type: application/json; charset=utf-8</pre>

#### Status

<pre>404 Not Found</pre>

#### Body

<pre>{"errors":[{"title": "request path not found"}]}</pre>


