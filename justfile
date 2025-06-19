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
test: goldplate license-check

[doc("Stop qBittorrent test server")]
docker-down:
  docker compose down qbittorrent

[doc("Ensure qBittorrent test server is running")]
docker-up:
  #!/bin/env sh
  # If the service is already running, exit early, otherwise the while loop
  # that greps the logs for the password will never exit.
  service_status="$(docker compose ps --services --status running qbittorrent)"
  if test ! -z "$service_status"; then
    exit 0
  fi
  # Start the service and block until the admin password is set, otherwise
  # downstream jobs will fail to authenticate. Under some circumstances, the
  # docker compose logs command may print old log entries from previous
  # executions, so we explicitly select only new log entries.
  since=$(date --iso-8601=seconds)
  docker compose up -d qbittorrent
  while true; do
    docker compose logs qbittorrent --since "$since" | grep -q "administrator password" && break
    sleep 0.5
  done

[doc("Run goldplate specs")]
goldplate: docker-up
  #!/bin/env sh
  export MOT_PASSWORD={{shell(get_pass)}}
  goldplate --diff ./test/init

clean: docker-down
  -rm ./mot

build $CGO_ENABLED="0":
  go build

[doc("Test source files for license header presence")]
license-check:
  addlicense -check -l apache -c "Ryan White" {{go_files}}

[doc("Add license header to source files")]
license-fix:
  addlicense -l apache -c "Ryan White" {{go_files}}

[doc("Run all fixes")]
fix: license-fix
