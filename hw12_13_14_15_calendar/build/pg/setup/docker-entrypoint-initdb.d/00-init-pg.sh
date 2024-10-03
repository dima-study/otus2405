#!/bin/bash

set -e

cp /docker-entrypoint-initdb.d/cfg/* $PGDATA/
