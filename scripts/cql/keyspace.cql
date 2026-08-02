-- Create a keyspace
CREATE KEYSPACE IF NOT EXISTS ocm WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : '1' };
