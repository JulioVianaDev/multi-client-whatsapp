# Database Volume

This directory contains the SQLite databases for WhatsApp instances.

## Structure

- `whatsapp_[instance_key].db` - Main SQLite database for each WhatsApp instance
- `whatsapp_[instance_key].db-shm` - SQLite shared memory file (temporary)
- `whatsapp_[instance_key].db-wal` - SQLite write-ahead log file (temporary)

## Access

The databases are stored in this volume and are accessible from your host machine. You can:

1. **View databases directly**: Open the `.db` files with any SQLite browser (like DB Browser for SQLite)
2. **Query databases**: Use SQLite command line tools
3. **Backup databases**: Copy the `.db` files to backup locations

## Docker Volume

This directory is mounted as a volume in Docker, so:

- Changes persist between container restarts
- You can access databases from your host machine
- Database files are stored in `./database_volume/` relative to your project root

## Example

When you create a WhatsApp instance with key `abc123`, a database file `whatsapp_abc123.db` will be created in this directory.
