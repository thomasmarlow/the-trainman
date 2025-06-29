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

# manual testing
curl http://localhost:8080/api/users/users
curl http://localhost:8080/api/users/profile
curl http://localhost:8080/api/orders/orders
curl http://localhost:8080/api/orders/status
```

### next stages

- [ ] stage 4: enforce x-request-id header
- [ ] stage 5: enforce api key authentication
- [ ] stage 6: deploy on aws

### tech stack

- **go 1.21** - main language
- **chi router** - lightweight and powerful http routing
- **fsnotify** - file system event notifications
- **yaml** - configuration format
- **docker** - containerization
- **alpine linux** - minimal base image
