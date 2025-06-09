export MOT_USERNAME := "admin"
export MOT_URL := "http://localhost:8080"

# When the docker compose stack is brought up, an admin password
# is generated and printed to the server log. This script parses
# it from the log.
get_pass := '''
docker compose logs qbittorrent \
| awk '/administrator password/ {i=$NF} END {print i}'
'''

go_files := shell('find . -name \*.go -printf "%p "')

[doc("Run all tests")]
test: goldplate check-license

docker-up:
  #!/bin/env sh
  docker compose up -d qbittorrent
  # Wait for the generated password to be set
  while true; do
    docker compose logs qbittorrent | grep -q "administrator password" && break
    sleep 0.5
  done

docker-down:
  docker compose down qbittorrent

goldplate: docker-up
  #!/bin/env sh
  export MOT_PASSWORD={{shell(get_pass)}}
  goldplate --diff ./test/init

clean: docker-down
  -rm ./mot

build $CGO_ENABLED="0":
  go build

check-license:
  addlicense -check -l apache -c "Ryan White" {{go_files}}

fix-license:
  addlicense -l apache -c "Ryan White" {{go_files}}

fix: fix-license
