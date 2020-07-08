# redirector
Simple golang service to redirect all urls to another domain.

Host -> Redirect domain mapping is read from redis on each request, 301 is served.

Set environment variables
```
REDIS_HOST
REDIS_PORT
REDIS_KEY
PORT
```

Set keys
```
{request.domain}:redirect = destination.domain
```
