CREATE TABLE IF NOT EXISTS ocm.credentials (
    accountPartition text,
    region text,
    country text,
    tenant text,
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
    PRIMARY KEY ((accountPartition, region, country, tenant), account)
);
