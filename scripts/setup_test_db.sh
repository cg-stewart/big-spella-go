#!/bin/bash

# Create test database
psql -U postgres -c "DROP DATABASE IF EXISTS bigspella_test;"
psql -U postgres -c "CREATE DATABASE bigspella_test;"

# Apply migrations
psql -U postgres -d bigspella_test -f migrations/001_initial_schema.sql
