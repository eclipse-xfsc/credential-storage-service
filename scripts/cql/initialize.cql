-- Create a keyspace
CREATE KEYSPACE IF NOT EXISTS tenant_space WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : '1' };

-- Create a table
CREATE TABLE IF NOT EXISTS tenant_space.credentials (
accountPartition text,
region text,
country text,
account text,
last_update_timestamp timestamp,
metadata map<text,text>,
credentials map<text,text>,
presentations map<text,text>,
id text,
recovery_nonce text,
device_key text,
nonce text,
locked boolean,
signature text,
PRIMARY KEY ((accountPartition,region,country),account)
);

CREATE INDEX IF NOT EXISTS ON tenant_space.credentials (locked);
CREATE INDEX IF NOT EXISTS ON tenant_space.credentials (id);