Install go 1.17
===============

	curl https://go.dev/dl/go1.17.3.linux-amd64.tar.gz
	tar -C /usr/local -xzf go1.17.3.linux-amd64.tar.gz

Install swaggo and swag command
===============================

	/usr/local/go/bin/go get -u github.com/swaggo/http-swagger
	CGO_ENABLED=0 /usr/local/go/bin/go get -u github.com/alecthomas/template
	CGO_ENABLED=0 /usr/local/go/bin/go get -u github.com/swaggo/swag/cmd/swag
	~/go/bin/swag init --parseDependency --parseInternal --parseDepth 1

Build for alpine container
==========================

	export CGO_ENABLED=0                 # if building for alpine
	export PATH=/usr/local/bin/go:$PATH  # if go is installed there
	go build

Testing
=======

Add a api container to an existing collector service:

	[container#api]
	image = alpine
	netns = container#0
	restart = 1
	rm = true
	volume_mounts = /home/cvaroqui/dev/collector-api:/collector-api
			{volume#0.name}/conf/nginx/ssl/:/ssl
			{volume#0.name}/conf/db/:/db
	environment = JWT_SIGN_KEY=/ssl/private_key
		      JWT_VERIFY_KEY=/ssl/certificate_chain
		      DB_PASSWORD_FILE=/db/password
	entrypoint = /collector-api/collector-api
	subset = front

Get a JWT from node credentials:

	PW=$(om node get --kw node.uuid)
	TOKEN=$(curl -s -u "$HOSTNAME:$PW" -H "Content-Type: application/json" https://collector/api/auth/nodes/login | jq .token | sed "s/\"//g")

Use the API authenticating with this token:

	curl -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" "https://collector/api/nodes"

Environment variables defaults
==============================

	DB_HOST=127.0.0.1
	DB_PORT=3306
	DB_USER=opensvc
	DB_PASSWORD or DB_PASSWORD_FILE (required)
	DB_SLOW_QUERY_THRESHOLD=1s
	DB_LOG_LEVEL=silent

	JWT_SIGN_KEY (required)
	JWT_VERIFY_KEY=

