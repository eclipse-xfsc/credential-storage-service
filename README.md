# Credential Storage Service

## Introduction

The Credential Storage Service provides storage and retrieval of credentials and presentations.

All stored content is cryptographically protected using a separate crypto-provider process that is accessed exclusively over gRPC. Depending on the deployment scenario, the service supports different operation modes and can either be accessed remotely through its REST API or internally through REST/NATS.

In remote mode, an additional encryption layer is applied to stored content to ensure operator exclusion.

The storage is tenant-aware and partitions data by tenant, region, country, and account.

## API Route Structure

The service is scoped by a `tenantId`. The tenant scope is provided by the parent router group created by `microservice-core-go/pkg/server`.

The core server registers tenant-scoped handlers below:

```text
/v1/tenants/:tenantId
```

`refineRoutes` receives this existing tenant router group and only appends the storage-specific route hierarchy.

Storage-specific routes are structured below the tenant as:

```text
/v1/tenants/:tenantId/storage/:region/:country/:account/...
```

For example:

```text
/v1/tenants/tenant-123/storage/eu/de/account-456/credentials
```

The route parameters have the following meaning:

| Parameter | Description |
|---|---|
| `tenantId` | Tenant owning the storage context |
| `region` | Region used for storage partitioning |
| `country` | Country used for storage partitioning |
| `account` | Account owning the credentials or presentations |

The corresponding Gin router hierarchy is:

```go
storageGroup := rg.Group("/storage")
regionGroup := storageGroup.Group("/:region")
countryGroup := regionGroup.Group("/:country")
accountGroup := countryGroup.Group("/:account")
```

`rg` already contains the `/:tenantId` scope. Therefore, handlers and middleware can access all parameters through the Gin context:

```go
tenantId := c.Param("tenantId")
region := c.Param("region")
country := c.Param("country")
account := c.Param("account")
```

Keeping the storage hierarchy below the tenant allows additional tenant-scoped APIs to be introduced independently without requiring them to use the storage-specific `region`, `country`, or `account` parameters.

# Credential Filter Logic

Credential filtering is based on the **Digital Credentials Query Language (DCQL)** defined by OpenID for Verifiable Presentations 1.0.

The previous Presentation Exchange `presentation_definition` based filtering is no longer used.

The storage service accepts a DCQL query and evaluates the contained credential queries against the credentials available for the requested account.

DCQL allows credentials to be selected based on properties such as:

- credential format,
- credential-specific metadata,
- requested claims,
- alternative claim sets,
- and credential sets.

The exact metadata fields depend on the credential format.

## Example DCQL Query

A query for an SD-JWT VC can look like:

```json
{
  "credentials": [
    {
      "id": "identity_credential",
      "format": "dc+sd-jwt",
      "meta": {
        "vct_values": [
          "https://example.com/credentials/identity"
        ]
      },
      "claims": [
        {
          "path": ["given_name"]
        },
        {
          "path": ["family_name"]
        },
        {
          "path": ["date_of_birth"]
        }
      ]
    }
  ]
}
```

The `credentials` array contains individual credential queries.

Each credential query has an `id` which identifies the requested credential within the DCQL query.

The `format` identifies the credential format. Format-specific matching information is supplied through `meta`.

For an SD-JWT VC, for example:

```json
"meta": {
  "vct_values": [
    "https://example.com/credentials/identity"
  ]
}
```

can be used to restrict the query to credentials with a matching Verifiable Credential Type (`vct`).

The `claims` section describes which claims are requested from matching credentials:

```json
"claims": [
  {
    "path": ["given_name"]
  },
  {
    "path": ["family_name"]
  }
]
```

Unlike the previous Presentation Exchange implementation, these paths follow **DCQL claim path semantics** and are not JSONPath expressions.

The DCQL model and matching functionality are provided by the [Eclipse XFSC OID4VCI/VP Library](https://github.com/eclipse-xfsc/oid4-vci-vp-library).

The normative definition of DCQL is part of [OpenID for Verifiable Presentations 1.0](https://openid.net/specs/openid-4-verifiable-presentations-1_0-final.html).

## Credential Sets

DCQL can also describe combinations or alternatives of credentials through `credential_sets`.

The storage implementation must therefore not treat a DCQL query as a simple flat list of independent JSON filters. Credential queries are evaluated according to the DCQL query semantics implemented by the OID4VCI/VP library.

# Flows

## Remote Usage Registration

If remote usage of the storage is enabled, the Storage Service requires registration before credentials or presentations can be stored.

Registration creates the initial account and device-key information in Cassandra.

```mermaid
sequenceDiagram
title Cloud Storage Registration
App->>Module: Initialize Cloud Storage
Module->>Module: Create Keypair for Content Encryption (CE)
Module->>Module: Create Keypair for Device Binding (DB)
Module->>Module: Create Key Storage
Module->>Key Storage: Store Key Pair
Module->>Key Storage: Extract Public Key
Module->>Key Storage: Create Self Signed JWT
Module->>Cloud Storage: Register Account with Self Signed JWT
Cloud Storage->>Cloud Storage: Insert Account / Register Device Key
Cloud Storage->>Module: Recovery Nonce + Nonce
Module->>Module: Sign Recovery Nonce with Device Key to RJWT
Module->>Module: Create QR Code (RJWT, CE)
Module->>App: Request Password
App->>Module: Password
Module->>Module: Encrypt QR with KDF(PW), Base45 Encoding and Compression
Module->>App: QR Code
```

After registration, the remote storage endpoints can be used.

## Remote Usage Credential/Presentation

```mermaid
sequenceDiagram
title Cloud Storage Operations

alt Add Credential
App->>Module: Store Credential
Module->>Cloud Storage: Create Session
Cloud Storage->>Module: Nonce
Module->>Module: Sign JWT + Nonce with Device Key
Module->>Module: Encrypt Credential with JWE
Module->>Cloud Storage: Call REST API for Add Credential
Module->>Module: Decrypt Response JWE
Module->>App: Response
end

alt Delete Credential
App->>Module: Delete Credential
Module->>Cloud Storage: Create Session
Cloud Storage->>Module: Nonce
Module->>Module: Sign JWT + Nonce with Device Key
Module->>Cloud Storage: Call REST API for Delete Credential
Module->>Module: Decrypt Response JWE
Module->>App: Response
end

alt Get Credentials
App->>Module: Get Credentials
Module->>Cloud Storage: Create Session
Cloud Storage->>Module: Nonce
Module->>Module: Sign JWT + Nonce with Device Key
Module->>Cloud Storage: Call REST API for Get Credentials with DCQL Query
Cloud Storage->>Cloud Storage: Match stored credentials against DCQL
Cloud Storage->>Module: Matching Credentials
Module->>Module: Decrypt Response JWE
Module->>App: Credentials
end
```

## General

In Direct Mode, no remote device registration or device authentication is required.

The API can be accessed directly, but the service must be protected by appropriate infrastructure controls depending on the deployment scenario, for example private networks, Kubernetes NetworkPolicies, ingress restrictions, service mesh authorization policies, or equivalent access-control mechanisms.

# API Documentation

## Direct Mode

The API documentation is provided as Swagger/OpenAPI in `docs/swagger.json`.

In the local Docker Compose environment, the Swagger UI is available at:

```text
http://localhost:8080/swagger/index.html
```

After changing API definitions, regenerate it with:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init --parseDependency
```

# Dependencies

The service requires Cassandra and a reachable crypto-provider gRPC endpoint. The selected crypto-provider implementation may additionally require OpenBao/Vault, an HSM, or another backing key store.

# Bootstrap

The service can be started using Docker Compose or Helm.

After startup:

1. Initialize the Cassandra database using [`scripts/cql/initialize.cql`](./scripts/cql/initialize.cql).
2. Configure the required signing and encryption keys.
3. Use Postman, Insomnia, or another REST client to interact with the API.

```bash
docker compose -f docker-compose.yml rm
docker compose -f docker-compose.yml --env-file=.env up --build --detach
```

## Using Cassandra

Installation: https://cassandra.apache.org/_/quickstart.html

```bash
pip install -U cqlsh
```

After bootstrapping the Docker Compose environment, the Cassandra initialization container may need to be executed after Cassandra becomes ready.

## Statements

```sql
DESCRIBE keyspaces;
SELECT * FROM ocm.credentials;
```

# Developer Information

## Setup

```bash
make docker-compose-run
```

## Cassandra

### General

The Cassandra data model is designed for large-scale credential storage and highly distributed read access.

Records are signed during insertion to protect them against modification outside of the Storage Service. Signature creation and verification are performed through `internal/middleware/authMiddleware.go`. The required signing and verification keys must be provisioned in the configured crypto-provider beforehand.

### Retrieve Credentials

```bash
cqlsh <cassandra host> <cassandra port> \
  -u <cassandra user> \
  -p <cassandra password> \
  -e "SELECT * FROM ocm.credentials;"
```

### Data Model

The credential table is designed so that Cassandra partitioning and clustering can take the storage location into account.

The API exposes the relevant storage context through `tenantId`, `region`, `country`, and `account`.

`region` and `country` are explicitly part of the storage route and are therefore available to the storage and authentication layers through the Gin request context.

# Use Modes

## Remote

**Type:** REST  
**Security:** Client Authentication + Device-Bound Tokens

The service can be used as remote storage for mobile applications. In this mode, the Storage Service verifies a registered device key against incoming tokens and supports device registration.

### Prerequisites

Each user requires:

- a valid account created and validated through the Account Service,
- a valid client certificate for mTLS,
- an RSA-PSS certificate for signing tokens and decrypting receipts,
- an RSA-PSS certificate for signing credential envelopes/content.

When using NATS in this mode, the user's public key must be known to the processing component so that protected messages can be processed correctly.

## Direct

**Type:** REST / NATS  
**Security:** Standard JWK

Direct Mode is intended for trusted internal communication and does not require remote device registration.

# API

## Route Structure

Storage APIs are located below:

```text
/v1/tenants/:tenantId/storage/:region/:country/:account
```

Individual resources are appended below the account scope, for example:

```text
/v1/tenants/:tenantId/storage/:region/:country/:account/credentials
/v1/tenants/:tenantId/storage/:region/:country/:account/presentations
```

The actual route suffixes are defined by the REST handlers and documented in the generated Swagger/OpenAPI specification.

## Authorization Bearer

Credential routes in Remote Mode are protected by a self-signed token created with the registered device key.

The JWT must contain the current nonce and the account identifier as subject. Nonces are single-use and become invalid after successful processing of a request and issuance of the next receipt.

## Receipt

A receipt is returned in a JWE envelope addressed to the registered device key. After decryption, the receipt contains the nonce and its expiration information for use with the next request.

## Add Credential

A credential is stored under a caller-provided credential ID.

The route is:

```text
PUT /v1/tenants/:tenantId/storage/:region/:country/:account/credentials/:id
```

The `:id` path parameter is the storage identifier used to address the credential later, for example when deleting it.

### Add JSON / LDP Credential

In Direct Mode the service uses `Content-Type: application/json`. The credential is sent directly as the request body.

Example:

```bash
curl -X PUT \
  'http://localhost:8080/v1/tenants/tenant-123/storage/eu/de/account-456/credentials/credential-1' \
  -H 'Content-Type: application/json' \
  -d '{
    "@context": [
      "https://www.w3.org/ns/credentials/v2"
    ],
    "type": [
      "VerifiableCredential"
    ],
    "issuer": "did:example:issuer",
    "credentialSubject": {
      "id": "did:example:holder",
      "given_name": "Arthur",
      "family_name": "Dent",
      "age": 42,
      "country": "DE"
    }
  }'
```

This credential can later be queried with a DCQL credential query using `format: "ldp_vc"` and claim paths such as:

```json
{
  "path": ["credentialSubject", "age"],
  "values": [42]
}
```

### Add SD-JWT VC

An SD-JWT VC uses the `dc+sd-jwt` credential format. The payload contains the SD-JWT and its disclosures in the compact SD-JWT representation.

A compact SD-JWT VC has the general form:

```text
<issuer-signed-jwt>~<disclosure-1>~<disclosure-2>~...
```

For example, assuming the SD-JWT VC has already been issued:

```bash
SD_JWT='eyJ0eXAiOiJzZCtqd3QiLCJhbGciOiJFUzI1NiJ9.eyJpZCI6IjEyMzQiLCJ2Y3QiOiJodHRwczovL2NyZWRlbnRpYWxzLmV4YW1wbGUuY29tL2lkZW50aXR5X2NyZWRlbnRpYWwiLCJpc3MiOiJodHRwczovL2lzc3Vlci5leGFtcGxlLmNvbSIsInN1YiI6ImRpZDpleGFtcGxlOmhvbGRlciIsIl9zZCI6WyJMcHlzRERfY2Fka0ZJby00WDVHWTN4UUtYLVdyYUpGbUV2N1ZTcmZ1M053IiwiVEZybVBBS2liOG1IeU54M1RPOUtGS2pGWEhNNEFQX3hKa2ZvWURxWXQ5WSIsIl9qSl93dVBTSXY0TlRNX1BGWS13cnVUOTZua2lGQ0YyYzhVeU5WSm9JUlUiXSwiX3NkX2FsZyI6IlNIQS0yNTYifQ.EZMU63KT3KODUAEMjesXiu6R-zayyL1xVwQXjT0mxqefW68-bi6-B2l03L2NmnFnlU6YMwLvl-nOY5uDm_Wcuw~WyI5ZTZkZDFjYWM0NDQ0NGNlIiwiZmlyc3RuYW1lIiwiSm9obiJd~WyJkMjBjMmUxOWQxMDU5NTMyIiwibGFzdG5hbWUiLCJEb2UiXQ~WyI2NDJiOGIyNGVjMDAxZGE1Iiwic3NuIiwiMTIzLTQ1LTY3ODkiXQ~'

curl -X PUT \
  'http://localhost:8080/v1/tenants/tenant-123/storage/eu/de/account-456/credentials/identity-sd-jwt' \
  -H 'Content-Type: application/json' \
  --data-binary "$SD_JWT"
```

The token above is intentionally illustrative and is **not a cryptographically valid credential**. Replace it with the compact SD-JWT VC returned by the issuer.

For example, a decoded SD-JWT VC payload relevant to the DCQL examples could contain:

```json
{
  "vct": "https://credentials.example.com/identity_credential",
  "_sd_alg": "sha-256",
  "_sd": [
    "..."
  ]
}
```

with a disclosure for `given_name`.

The corresponding DCQL query can then select the credential by format, `vct`, and disclosed claim:

```json
{
  "credentials": [
    {
      "id": "pid",
      "format": "dc+sd-jwt",
      "meta": {
        "vct_values": [
          "https://credentials.example.com/identity_credential"
        ]
      },
      "claims": [
        {
          "path": ["given_name"]
        }
      ]
    }
  ]
}
```

The `vct` value contained in the SD-JWT VC must match one of the values supplied in `meta.vct_values` for this query to match.

> **Note:** The example above documents the compact SD-JWT VC representation expected conceptually by the DCQL implementation. The exact request representation accepted by the storage endpoint depends on how the storage service maps an incoming `application/json` body to `types.Credential`. Do not wrap the token in an invented JSON property such as `{"credential": "..."}` unless the service API explicitly defines such a request structure.

### Remote Mode

In Remote Mode the same route is used:

```text
PUT /v1/tenants/:tenantId/storage/:region/:country/:account/credentials/:id
```

The remote storage flow uses `Content-Type: application/jose`. The request body contains the JWE expected by the remote storage protocol rather than the plain credential payload shown in the Direct Mode examples.

## Get Credentials with DCQL

In Direct Mode, credentials are retrieved and filtered using:

```text
POST /v1/tenants/:tenantId/storage/:region/:country/:account/credentials
```

The POST body is the **DCQL query itself**. There is no additional `dcql_query` wrapper around it.

A DCQL query must contain at least one entry in `credentials`. If the payload is empty, all credentials of account will be returned.

```
curl -X POST \
  'http://localhost:8080/v1/tenants/tenant-123/storage/eu/de/account-456/credentials' \
  -H 'Content-Type: application/json'
```

### Minimal DCQL Query

The smallest useful query contains a credential query ID, a credential format, and the format metadata object:

```json
{
  "credentials": [
    {
      "id": "pid",
      "format": "ldp_vc",
      "meta": {}
    }
  ]
}
```

Example request:

```bash
curl -X POST \
  'http://localhost:8080/v1/tenants/tenant-123/storage/eu/de/account-456/credentials' \
  -H 'Content-Type: application/json' \
  -d '{
    "credentials": [
      {
        "id": "pid",
        "format": "ldp_vc",
        "meta": {}
      }
    ]
  }'
```

### Filter by Claim Value

DCQL claim paths are arrays of path components and are **not JSONPath expressions**.

For example, to find an `ldp_vc` credential whose `credentialSubject.age` is `42`:

```bash
curl -X POST \
  'http://localhost:8080/v1/tenants/tenant-123/storage/eu/de/account-456/credentials' \
  -H 'Content-Type: application/json' \
  -d '{
    "credentials": [
      {
        "id": "age-query",
        "format": "ldp_vc",
        "meta": {},
        "claims": [
          {
            "path": ["credentialSubject", "age"],
            "values": [42]
          }
        ]
      }
    ]
  }'
```

A claim query can allow multiple values. This example matches credentials where `credentialSubject.country` is either `DE` or `FR`:

```bash
curl -X POST \
  'http://localhost:8080/v1/tenants/tenant-123/storage/eu/de/account-456/credentials' \
  -H 'Content-Type: application/json' \
  -d '{
    "credentials": [
      {
        "id": "eu-country",
        "format": "ldp_vc",
        "meta": {},
        "claims": [
          {
            "path": ["credentialSubject", "country"],
            "values": ["DE", "FR"]
          }
        ]
      }
    ]
  }'
```

### Require Multiple Claims

Without `claim_sets`, all requested claims of the credential query must match.

```bash
curl -X POST \
  'http://localhost:8080/v1/tenants/tenant-123/storage/eu/de/account-456/credentials' \
  -H 'Content-Type: application/json' \
  -d '{
    "credentials": [
      {
        "id": "person",
        "format": "ldp_vc",
        "meta": {},
        "claims": [
          {
            "path": ["credentialSubject", "age"],
            "values": [42]
          },
          {
            "path": ["credentialSubject", "country"],
            "values": ["DE"]
          }
        ]
      }
    ]
  }'
```

### Alternative Claim Sets

Claims can be assigned IDs and combined through `claim_sets`.

In the following example either `name + country` or `name + age` can satisfy the credential query:

```json
{
  "credentials": [
    {
      "id": "person",
      "format": "ldp_vc",
      "meta": {},
      "claims": [
        {
          "id": "name",
          "path": ["credentialSubject", "given_name"],
          "values": ["Arthur"]
        },
        {
          "id": "age",
          "path": ["credentialSubject", "age"],
          "values": [42]
        },
        {
          "id": "country",
          "path": ["credentialSubject", "country"],
          "values": ["DE"]
        }
      ],
      "claim_sets": [
        ["name", "country"],
        ["name", "age"]
      ]
    }
  ]
}
```

All IDs referenced by `claim_sets` must refer to claims declared in the same credential query.

### SD-JWT VC Query

For an SD-JWT VC, `meta.vct_values` can restrict the query to one or more Verifiable Credential Types.

```bash
curl -X POST \
  'http://localhost:8080/v1/tenants/tenant-123/storage/eu/de/account-456/credentials' \
  -H 'Content-Type: application/json' \
  -d '{
    "credentials": [
      {
        "id": "pid",
        "format": "dc+sd-jwt",
        "meta": {
          "vct_values": [
            "https://credentials.example.com/identity_credential"
          ]
        },
        "claims": [
          {
            "path": ["given_name"]
          }
        ]
      }
    ]
  }'
```

The query above matches an SD-JWT VC only if its `vct` matches one of the supplied `vct_values` and the requested `given_name` claim is available.

### Multiple Credential Queries

A DCQL request can contain multiple credential queries. Each query ID must be unique.

```json
{
  "credentials": [
    {
      "id": "identity",
      "format": "dc+sd-jwt",
      "meta": {
        "vct_values": [
          "https://credentials.example.com/identity_credential"
        ]
      }
    },
    {
      "id": "person",
      "format": "ldp_vc",
      "meta": {},
      "claims": [
        {
          "path": ["credentialSubject", "country"],
          "values": ["DE"]
        }
      ]
    }
  ]
}
```

The filter result is associated with the corresponding credential query ID. Multiple stored credentials can match the same credential query.

### Credential Sets

DCQL can express combinations or alternatives of credential queries using `credential_sets`.

References inside a credential set must point to existing credential query IDs.

For example:

```json
{
  "credentials": [
    {
      "id": "identity",
      "format": "dc+sd-jwt",
      "meta": {
        "vct_values": [
          "https://credentials.example.com/identity_credential"
        ]
      }
    },
    {
      "id": "person",
      "format": "ldp_vc",
      "meta": {}
    }
  ],
  "credential_sets": [
    {
      "options": [
        ["identity"],
        ["person"]
      ]
    }
  ]
}
```

The DCQL implementation validates references in `claim_sets` and `credential_sets`. Unknown references and duplicate credential query or claim IDs are rejected.

## Delete Credential

A credential is deleted by its storage ID.

The route is:

```text
DELETE /v1/tenants/:tenantId/storage/:region/:country/:account/credentials/:id
```

No request body is required.

Example:

```bash
curl -X DELETE \
  'http://localhost:8080/v1/tenants/tenant-123/storage/eu/de/account-456/credentials/credential-1'
```

In Direct Mode the credential is removed directly from the storage context identified by `tenantId`, `region`, `country`, and `account`.

In Remote Mode the same route is protected by the remote authentication flow and a successful operation returns the transaction receipt according to the remote storage protocol.

# Crypto Provider gRPC Configuration

The Storage Service does not load Go plugins or crypto modules into its own process.

All key generation, random generation, signing, verification, encryption, and decryption operations are delegated through `crypto-provider-core` to the configured crypto-provider gRPC endpoint.

```bash
STORAGESERVICE_CRYPTO_GRPC_ADDR=crypto-provider:50051
```

For local development, `deployment/docker/docker-compose.yml` starts the Storage Service, Cassandra, OpenBao, and a separate crypto-provider container.

Override `CRYPTO_PROVIDER_IMAGE` when a different crypto-provider implementation or version is required.
