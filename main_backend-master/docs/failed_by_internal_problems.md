### Any handler (i.e. /cities/:provider_name/movies/:movie_id/seances) 
Raises when redis-conf has mistakes and provider api unavailable
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

<pre>{"errors":[
{
"title": "Main API: Cannot assign requested address",
"detail": "Traceback (most recent call last):\n  File \"/usr/local/lib/python3.8/site-packages/aiohttp/connector.py\", line 936, in _wrap_create_connection\n    return await self._loop.create_connection(*args, **kwargs)  # type: ignore  # noqa\n  File \"/usr/local/lib/python3.8/asyncio/base_events.py\", line 1025, in create_connection\n    raise exceptions[0]\n  File \"/usr/local/lib/python3.8/asyncio/base_events.py\", line 1010, in create_connection\n    sock = await self._connect_sock(\n  File \"/usr/local/lib/python3.8/asyncio/base_events.py\", line 924, in _connect_sock\n    await self.sock_connect(sock, address)\n  File \"/usr/local/lib/python3.8/asyncio/selector_events.py\", line 496, in sock_connect\n    return await fut\n  File \"/usr/local/lib/python3.8/asyncio/selector_events.py\", line 501, in _sock_connect\n    sock.connect(address)\nOSError: [Errno 99] Cannot assign requested address\n\nThe above exception was the direct cause of the following exception:\n\nTraceback (most recent call last):\n  File \"main.py\", line 37, in proxy_request_to_cinema_api\n    response = await getattr(app['client_session'], method)\\\n  File \"/usr/local/lib/python3.8/site-packages/aiohttp/client.py\", line 480, in _request\n    conn = await self._connector.connect(\n  File \"/usr/local/lib/python3.8/site-packages/aiohttp/connector.py\", line 523, in connect\n    proto = await self._create_connection(req, traces, timeout)\n  File \"/usr/local/lib/python3.8/site-packages/aiohttp/connector.py\", line 858, in _create_connection\n    _, proto = await self._create_direct_connection(\n  File \"/usr/local/lib/python3.8/site-packages/aiohttp/connector.py\", line 1004, in _create_direct_connection\n    raise last_exc\n  File \"/usr/local/lib/python3.8/site-packages/aiohttp/connector.py\", line 980, in _create_direct_connection\n    transp, proto = await self._wrap_create_connection(\n  File \"/usr/local/lib/python3.8/site-packages/aiohttp/connector.py\", line 943, in _wrap_create_connection\n    raise client_error(req.connection_key, exc) from exc\naiohttp.client_exceptions.ClientConnectorError: Cannot connect to host localhost:2112 ssl:default [Cannot assign requested address]\n"
}
]}</pre>


