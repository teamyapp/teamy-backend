#!/bin/bash

USER=$1
PASSWORD=$2
DATABASE=$3
MIGRATE_DIRECTION=$4
MIGRATE_STEPS=$5

migrate -source file://app/repo/db/migration -database \
	"postgres://$USER:$PASSWORD@localhost:5432/$DATABASE?sslmode=disable" \
	"$MIGRATE_DIRECTION" \
	"$MIGRATE_STEPS"
