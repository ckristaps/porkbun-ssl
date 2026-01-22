#!/bin/sh

PUID=${PUID:-1000}
PGID=${PGID:-1000}

echo "-------------------------------------"
echo "Starting porkbun-ssl with:"
echo "  UID: $PUID"
echo "  GID: $PGID"
echo "-------------------------------------"

# Modify group if PGID differs
if [ "$(id -g appuser)" != "$PGID" ]; then
    sed -i "s/appuser:x:[0-9]*:/appuser:x:${PGID}:/" /etc/group
fi

# Modify user if PUID differs  
if [ "$(id -u appuser)" != "$PUID" ]; then
    sed -i "s/appuser:x:[0-9]*:[0-9]*:/appuser:x:${PUID}:${PGID}:/" /etc/passwd
fi

# Fix ownership of app directories
chown -R appuser:appuser /app /certs

exec su-exec appuser /app/porkbun-ssl "$@"
