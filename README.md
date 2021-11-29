Install go 1.17
===============

	curl https://go.dev/dl/go1.17.3.linux-amd64.tar.gz
	tar -C /usr/local -xzf go1.17.3.linux-amd64.tar.gz

Build for alpine container
==========================

	CGO_ENABLED=0 /usr/local/go/bin/go build

Testing
=======

Get a JWT from node credentials:

	PW=$(om node get --kw node.uuid)
	TOKEN=$(curl -s -u "$HOSTNAME:$PW" -H "Content-Type: application/json" http://collector:8080/nodes/login | jq .token | sed "s/\"//g")

Use the API authenticating with this token:

	curl -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" "http://collector:8080/nodes"

