#!/bin/bash

docker compose -f build/docker-compose.yaml run -e CALENDAR_LOG_LEVEL=debug -e CALENDAR_NOTIFY_INTERVAL=5s app-calendar-integration-tests

exit_status=$?

docker compose -f build/docker-compose.yaml down

echo $exit_status
exit $exit_status
