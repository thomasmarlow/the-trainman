# the-trainman

a minimal and extensible api gateway/reverse proxy built in go.

## project structure

```
the-trainman/
├── cmd/gateway/main.go          # application entry point + graceful shutdown
├── internal/
│   ├── config/manager.go        # hot reload config manager
│   ├── proxy/handler.go         # reverse proxy implementation
│   └── server/server.go         # http server + routing
├── mock-data/
│   ├── users.json               # mock user data for testing
│   └── orders.json              # mock order data for testing
├── config.yaml.template         # configuration template
├── config.yaml                  # yaml configuration (gitignored)
├── Makefile                     # development commands with proxy tests
├── go.mod                       # go dependencies
├── Dockerfile                   # multi-stage docker image
├── docker-compose.yml           # orchestration with mock services
├── .gitignore                   # git exclusions
└── .dockerignore                # build context optimization
```

## stage 1: basic api gateway with ping endpoint ✅

### implemented features

- ✅ http server in go with chi router
- ✅ `/ping` endpoint that responds with json
- ✅ complete dockerization with multi-stage build
- ✅ docker-compose for development
- ✅ automatic health check
- ✅ graceful shutdown
- ✅ request logging

## stage 2: configuration with hot reload ✅

### implemented features

- ✅ yaml configuration system
- ✅ hot reload with fsnotify
- ✅ polling fallback mechanism for when fsnotify has issues (e.g., on mac)
- ✅ thread-safe configuration access with sync.rwmutex
- ✅ `/ping` endpoint now uses configurable message

### usage

#### quick start with makefile

```bash
# setup configuration and start the service
make start

# test the endpoint
make test

# view logs
make logs

# stop the service
make stop

# see all available commands
make help
```

#### manual setup (alternative)

```bash
# create config from template
cp config.yaml.template config.yaml

# start the api gateway
docker-compose up --build

# test the endpoint
curl http://localhost:8080/ping

# expected response:
# {"status":"ok","message":"pong from config!"}

# stop the service
docker-compose down
```

#### test hot reload

```bash
# with the container running, modify config.yaml
echo 'message: "new config message!"' > config.yaml

# test the endpoint again (should reflect the change)
curl http://localhost:8080/ping

# expected response:
# {"status":"ok","message":"new config message!"}
```

#### verify health check

```bash
# check container status
make health
# or manually:
docker-compose ps

# should show: Up X minutes (healthy)
```

## stage 3: redirection to backend services ✅

### implemented features

- ✅ reverse proxy functionality with path rewriting
- ✅ configurable backend services in yaml
- ✅ request forwarding with pattern `/api/<service_name>/<path>`
- ✅ mock backend services using json-server
- ✅ header forwarding and hop-by-hop filtering
- ✅ x-forwarded headers for proper proxying
- ✅ comprehensive testing commands

### backend services

the gateway now supports proxying requests to backend services:

- **users service**: `http://localhost:8080/api/users/*` → `http://mock-users/*`
- **orders service**: `http://localhost:8080/api/orders/*` → `http://mock-orders/*`

### testing proxy functionality

```bash
# test all endpoints
make test-all

# test specific services
make test-users
make test-orders

# test x-request-id enforcement
make test-request-id

# manual testing (note: users service now requires x-request-id header)
curl -H "x-request-id: test-123" http://localhost:8080/api/users/users
curl -H "x-request-id: test-123" http://localhost:8080/api/users/profile
curl http://localhost:8080/api/orders/orders
curl http://localhost:8080/api/orders/status
```

## stage 4: x-request-id header enforcement ✅

### implemented features

- ✅ configurable x-request-id header enforcement per service
- ✅ global enforcement setting with service-level override capability
- ✅ customizable error messages for missing headers
- ✅ detailed logging of rejected requests with ip tracking
- ✅ hot reload support for enforcement configuration changes
- ✅ backward compatibility with existing configurations

### configuration structure

the gateway now supports granular control over x-request-id enforcement:

```yaml
# global enforcement configuration
request_id:
  require_request_id: false                   # global enforcement (default: false)
  override_service_settings: false           # if true, ignores per-service settings
  error_message: "Missing required header: x-request-id"  # customizable error message

# per-service enforcement configuration
backend_services:
  - name: "users"
    url: "http://mock-users"
    enabled: true
    require_request_id: true                  # this service requires x-request-id
  - name: "orders"
    url: "http://mock-orders"
    enabled: true
    require_request_id: false                 # this service doesn't require x-request-id
```

### enforcement logic

the system follows a hierarchical precedence model:

1. **service override mode** (default): `override_service_settings: false`
   - if service has `require_request_id` configured → use service setting
   - if service doesn't have `require_request_id` configured → use global setting
   - if service doesn't exist in config → use global setting

2. **global override mode**: `override_service_settings: true`
   - always use global `require_request_id` setting
   - ignores all per-service configurations

### testing x-request-id enforcement

```bash
# test x-request-id enforcement scenarios
make test-request-id

# test with header (should pass)
curl -H "x-request-id: test-123" http://localhost:8080/api/users/users

# test without header on service that requires it (should fail with 400)
curl http://localhost:8080/api/users/users
# expected response: Missing required header: x-request-id

# test without header on service that doesn't require it (should pass)
curl http://localhost:8080/api/orders/orders

# ping endpoint is never affected by x-request-id enforcement
curl http://localhost:8080/ping
```

### enforcement behavior

- **scope**: only applies to `/api/{service}/*` routes, never to `/ping`
- **validation**: occurs before proxy forwarding to backend services
- **response**: 400 bad request with configurable error message
- **logging**: detailed rejection logs include service name and client ip
- **hot reload**: configuration changes apply immediately without restart

## stage 5: x-api-key header enforcement ✅

### implemented features

- ✅ configurable x-api-key header enforcement per service
- ✅ single global api key validation with configurable value
- ✅ global enforcement setting with service-level override capability
- ✅ customizable error messages for missing and invalid api keys
- ✅ detailed logging of rejected requests with ip tracking
- ✅ hot reload support for api key configuration changes
- ✅ backward compatibility with existing configurations

### configuration structure

the gateway now supports granular control over x-api-key enforcement:

```yaml
# global api key enforcement configuration
api_key:
  api_key: "your-secret-api-key-here"                  # the valid api key
  require_api_key: false                               # global enforcement (default: false)
  override_service_settings: false                     # if true, ignores per-service settings
  error_message: "Missing required header: x-api-key"  # customizable error message

# per-service api key enforcement configuration
backend_services:
  - name: "users"
    url: "http://mock-users"
    enabled: true
    require_request_id: true                  # this service requires x-request-id
    require_api_key: true                     # this service requires x-api-key
  - name: "orders"
    url: "http://mock-orders"
    enabled: true
    require_request_id: false                 # this service doesn't require x-request-id
    require_api_key: false                    # this service doesn't require x-api-key
```

### enforcement logic

the system follows the same hierarchical precedence model as x-request-id:

1. **service override mode** (default): `override_service_settings: false`
   - if service has `require_api_key` configured → use service setting
   - if service doesn't have `require_api_key` configured → use global setting
   - if service doesn't exist in config → use global setting

2. **global override mode**: `override_service_settings: true`
   - always use global `require_api_key` setting
   - ignores all per-service configurations

### testing x-api-key enforcement

```bash
# test x-api-key enforcement scenarios
make test-api-key

# detailed testing scenarios
make test-api-key-detailed

# test with valid api key (should pass)
curl -H "x-api-key: your-secret-api-key-here" http://localhost:8080/api/users/profile

# test without api key on service that requires it (should fail with 400)
curl http://localhost:8080/api/users/profile
# expected response: Missing required header: x-api-key

# test with invalid api key on service that requires it (should fail with 401)
curl -H "x-api-key: wrong-key" http://localhost:8080/api/users/profile
# expected response: Invalid API key

# test without api key on service that doesn't require it (should pass)
curl http://localhost:8080/api/orders/list

# ping endpoint is never affected by x-api-key enforcement
curl http://localhost:8080/ping
```

### enforcement behavior

- **scope**: only applies to `/api/{service}/*` routes, never to `/ping`
- **validation**: occurs after x-request-id validation, before proxy forwarding
- **response codes**: 
  - 400 bad request for missing x-api-key header
  - 401 unauthorized for invalid x-api-key value
- **logging**: detailed rejection logs include service name and client ip
- **hot reload**: configuration changes apply immediately without restart
- **security**: api key value is stored in configuration, not logged in rejection messages

### next stages

- [ ] stage 6: deploy on aws

### extra tasks

- [ ] make invalid api key message configurable
- [ ] document the "why"s
- [ ] clean up README, build changelog elsewhere
- [ ] forward auth
- [ ] more robust e2e testing framework
- [ ] metrics and monitoring endpoints
- [ ] rate limiting per service
- [ ] circuit breaker
- [ ] update all dependencies to latest versions

### tech stack

- **go 1.21** - main language
- **chi router** - lightweight and powerful http routing
- **fsnotify** - file system event notifications
- **yaml** - configuration format
- **docker** - containerization
- **alpine linux** - minimal base image
