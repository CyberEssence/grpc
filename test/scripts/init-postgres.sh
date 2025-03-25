#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "postgres" --dbname "postgres" <<-EOSQL
    CREATE DATABASE auth_service;
    CREATE DATABASE call_service;
    GRANT ALL PRIVILEGES ON DATABASE auth_service TO postgres;
    GRANT ALL PRIVILEGES ON DATABASE call_service TO postgres;
EOSQL
