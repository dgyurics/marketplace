#!/bin/bash

# VARIABLES
DB_NAME="marketplace"
DB_USER="marketplace_user"
DB_PASS="your_secure_password"

# 1. Create database
sudo -u postgres createdb $DB_NAME

# 2. Create user with password
sudo -u postgres psql -c "CREATE USER $DB_USER WITH ENCRYPTED PASSWORD '$DB_PASS';"

# 3. Grant privileges to the user on the new database
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"

# 4. Apply schema to the new database
curl -sL https://raw.githubusercontent.com/dgyurics/marketplace/main/db/migrations/01_ddl.sql -o /tmp/01_ddl.sql
sudo -u postgres psql -d $DB_NAME -f /tmp/01_ddl.sql
