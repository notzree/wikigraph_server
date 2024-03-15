FROM postgres:14-alpine

# Run a script to create the pg_trgm extension in your database
# Note: This script runs every time the container starts
RUN echo "CREATE EXTENSION IF NOT EXISTS pg_trgm;" > /docker-entrypoint-initdb.d/init-pgtrgm.sql
