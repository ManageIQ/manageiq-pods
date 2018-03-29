#!/bin/bash

psql --command "ALTER ROLE \"${POSTGRESQL_USER}\" SUPERUSER;"
