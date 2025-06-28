# the-trainman

a minimal and extensible api gateway/reverse proxy built in go.

## stage 1: basic api gateway with ping endpoint

### implemented features

- ✅ http server in go with chi router
- ✅ `/ping` endpoint that responds with json
- ✅ complete dockerization with multi-stage build
- ✅ docker-compose for development
- ✅ automatic health check
- ✅ graceful shutdown
- ✅ request logging

### project structure

```
the-trainman/
├── cmd/gateway/main.go          # application entry point
├── internal/server/server.go    # http server logic
├── go.mod                       # go dependencies
├── Dockerfile                   # multi-stage docker image
├── docker-compose.yml           # local orchestration
└── .dockerignore               # build context optimization
```

### usage

#### local development

```bash
# start the api gateway
docker-compose up --build

# test the endpoint
curl http://localhost:8080/ping

# expected response:
# {"status":"ok","message":"pong"}

# stop the service
docker-compose down
```

#### verify health check

```bash
# check container status
docker-compose ps

# should show: Up X minutes (healthy)
```

### next stages

- [ ] stage 2: config reading with hot reload
- [ ] stage 3: redirection to backend services
- [ ] stage 4: enforce x-request-id header
- [ ] stage 5: enforce api key authentication
- [ ] stage 6: deploy on aws

### tech stack

- **go 1.21** - main language
- **chi router** - lightweight and powerful http routing
- **docker** - containerization
- **alpine linux** - minimal base image
