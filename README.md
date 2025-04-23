<h1>Introduction</h1>

The Storage Service is a service made for storing credentials and presentations. All content is stored cryptographically protected by [crypto engine](https://gitlab.eclipse.org/eclipse/xfsc/libraries/crypto/engine). Depending on the usage scenario, the service supports different modi which allows it to either use the service directly from remote or internally by nats. In the case of the remote the service enforces an additional encryption layer to ensure the operator exclusion. 

# Filter Logic

For getting credentials via filter out of the storage, the [presentation definition syntax](https://identity.foundation/presentation-exchange/) is directly used. During the post the credentials will be filtered over the presentation definition as follows: 
```
{
	  "id": "32f54163-7166-48f1-93d8-ff217bdb0653",
	  "input_descriptors": [
		{
		  "id": "wa_driver_license",
		  "name": "Washington State Business License",
		  "purpose": "We can only allow licensed Washington State business representatives into the WA Business Conference",
		  "constraints": {
			"fields": [
			  {
				"path": [
				  "$.credentialSubject.dateOfBirth",
				  "$.credentialSubject.dob",
				]
			  }
			]
		  }
		}
	  ]
  }
```

This example searches for credentials with dob OR dateOfBirth in the credentialSubject. The [lib](https://github.com/eclipse-xfsc/oid4-vci-vp-library/-/tree/main?ref_type=heads) is following json path logic. 


More examples can be found in the dif spec or in the [tests](https://github.com/eclipse-xfsc/oid4-vci-vp-library/-/blob/main/model/presentation/presentationDefinition_test.go?ref_type=heads#L256)


# Flows

## Remote Usage Registration

If the remote usage of the storage is enabled the Storage service requires an Registration before the storing can start. This generates in the cassandra db a record where the initial key pairs are stored: 

```mermaid
sequenceDiagram
title Cloud Storage Registration
App->>Module:Initialize Cloud Storage
Module->>Module:Create Keypair for Content Encryption(CE) and create Keypair for Device Binding (DB)
Module->>Module: Create Key Storage
Module->>Key Storage: Store Key Pair
Module->>Key Storage: Extract Public Key
Module->>Key Storage: Create Self Signed JWT
Module->>Cloud Storage: Register Account with Self Signed JWT
Cloud Storage->>Cloud Storage: Insert in Db/Register Device Key
Cloud Storage->>Module: Recovery Nonce + Nonce
Module->>Module: Sign Recovery Nonce with Device Key to RJWT
Module->>Module: Create QR Code (RJWT, CE) 
Module->>App: Please insert Password
App->>Module:Password is...
Module->>Module: Encrypt QR with KDF(PW) and base45 Encoding/Compression
Module->>App: QR Code
```

After that the storage endpoints can be used publicly.

## Remote Usage Credential/Presentation

```mermaid
sequenceDiagram
title Cloud Storage Operations
alt Add Credential
App->>Module: Store Credential
Module->>Cloud Storage: Create Session
Cloud Storage->Module: Nonce
Module->>Module: Sign JWT + Nonce with Device Key
Module->>Module: Encrypt Credential with JWE
Module->>Cloud Storage: Call Rest Api for Add Credential
Module->>Module: Decrypt Response JWE
Module->>App:  Reponse
end

alt Delete Credential
App->>Module: Delete Credential
Module->>Cloud Storage: Create Session
Cloud Storage->>Module: Nonce
Module->>Module: Sign JWT + Nonce with Device Key
Module->>Cloud Storage: Call Rest Api for Delete Credential
Module->>Module: Decrypt Response JWE
Module->>App:  Response
end

alt Get Credentials
App->>Module: Get Credentials
Module->>Cloud Storage: Create Session
Cloud Storage->Module: Nonce
Module->>Module: Sign JWT + Nonce with Device Key
Module->>Cloud Storage: Call Rest Api for Get Credential
Module->>Module: Decrypt Response JWE
Module->>App:  Credentials
end
```

## General

Within the internal mode no registration an auth is required. The api calls can be used directly, but they should be protected by any seperated network, ingress rules, istio auth rules etc. depending on the scenario.

## API Documentation (Direct Mode)
The API documentation is written in Swagger [Swagger Web UI](docs/swagger.json)

It can be accessed in local docker-compose environment [swagger_ui](http://localhost:8080/swagger/index.html).

It is generated using [swagger generation tool](https://github.com/swaggo/gin-swagger)

In case of changes in api definitions, it should be updated. Please run:
```cmd
go install github.com/swaggo/swag/cmd/swag@latest
swag init --parseDependency
```


# Dependencies

The Service requires a cassandra db and optionally an mobile protection solution(in the case of remote usage from smartphone) In case of hashicorp vault crypto plugin, a hashicorp vault is required.

# Bootstrap

The service can be started with the docker compose or via helm.  After startup: 

1. Use the CQL script to [initialize](./scripts/cql/initialize.cql) the cassandra db
2. Use Postman/Insomnia + Tokens+ PS256 Key Pairs can be created here: https://dinochiesa.github.io/jwt/

Using docker compose directly:

```	
docker compose -f docker-compose.yml rm
docker compose -f docker-compose.yml --env-file=.env up --build --detach
```

## Using Cassandra

Installation: https://cassandra.apache.org/_/quickstart.html
Python console: 
```
 pip install -U cqlsh
```
Note: After bootstrapping the docker compose file, the initialize/insert container must be executed manually so long as no error occurs, because the cassandra DB needs a while to bootstrap.

## Statements
```
DESCRIBE keyspaces; //you should see tenant_space here
select * from tenant_space.credentials;
```


# Developer Information

## Setup 

Using [makefile](./makefile)

```make docker-compose-run```

## Cassandra

#### General

The cassandra database was inserted for a very large scale of credentials combined with a high redudancy for reading. In the proper setup accross datacenters, the database can quarantee a high distributed way of reading and writing data especially for mobile devices an decentralized use cases. 

All records are <b>signed</b> during insertion which ensures that records can't be tampered outside of the software. This is realized by the [auth middleware](./internal/middleware/authMiddleware.go) The sign/verify key must be created before hand in the crypto engine. 

#### Retrieve credentials

```bash
cqlsh <cassandra host> <cassandra port> -u <cassandra user> -p <cassandra password> -e "SELECT * FROM tenant_space.credentials;"

```

#### Datamodel 

The database schema of the credential table is designed in such a way that a cassandra table clustering can be constructed accross regions (Country/Region Field) together with the first 4 bytes of the account id as cluster identifier. 

## Use Modes 

<h3><u> Remote</u> </h3>

Type: REST
Security: Client Authentication + Device Bound Tokens

The service can be used as a remote storage for mobile applications. In this case the storage verifies a device key against incoming tokens and accepts device registrations. For using this mode be aware that an additional protection by an mobile app protection framework is required to issue client certificates to establish an secure channel to the storage. An addtional fraud management can be docked by configuring the fraud management middleware.

<h3> Prerequisites</h3>

Each User must have the following preperations to use the backend: 

- Valid Account created and validated over the Account Service
- Valid Client Certificate for mTLS
- RSA PSS Certificate for Signing Tokens and Decrypting Receipts
- RSA PSS Certificate to sign Credential Envelopes (Content Certificate)

Note: In this mode, nats is only possible when the public key of the user is known otherwise the nats messages cant be processed.

<h3> <u>Direct</u> </h3>

Type: REST/NATS
Security: Standard JWK


<h2> API </h2>

<h3> Structure </h3>

TODO OPENAPI

<h3> Authorization Bearer</h3>

All credential routes are protected by a self signed token which must be created with the registered device key. The jwt must contain in the body: 

- Nonce (initial nonce or from receipt)
- Subject (account id)

All nonces are just usable once and are invalid after receiving an receipt.

Example(signed with token from this repo):

```
eyJhbGciOiJQUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOlsiL3Rlc3QyL3RlbmFudF9zcGFjZS9BQkNEMTIzIl0sImV4cCI6MTcwMDUyNDU3OSwiaWF0IjoxNzAwNTIwOTc5LCJpc3MiOiJnaXRodWIuY29tL2xlc3RycmF0LWdvL2p3eCIsIm5vbmNlIjoiTlRVMU5RPT0iLCJzdWIiOiJBQkNEMTIzIn0.hOrqF6L6PaKsSOxlF6hUCBbDMJBhwXa7trqoW0cLw28CWv8bZxbWJccnfomkIin-ygJPsw4U7d8o4OHfhdHPETx7RGo2XBktIheERzivOORl3eDpqU7zfV9FTuU-U8dH9Tmx4HSpW9AS0c0SJjxzew4iwNC_UBy5KRBkgIwS7ZDNEDLpc4NTEwT9jfXwUUR6PopfeFr9kDtVEcmsXCxRVe8QQ_x_mFk_xtMoYzx-tKTLtXpH3KYKYmmAbcjkrfajcFi5w5Cai3OAO2ugfRZ6XTEev3_8SPMVgPq2Qoz2Ka3Cq26-7A7GzLrrLo5NAWe3urUMfxvtU9wsf_Wi535niw
```

<h3> Receipt </h3>

The receipt is returned in an JWE envelop adressed to the device key. After decrypting it, the receipt contains a nonce and nonce expired value for usage in the next call. 

<h3> Usage </h3>

<b> Register Device</b>

<i> Expects</i>

Content-Type: application/json <br>
Body: None <br>
Method: GET

Example:

```
curl -H 'Authorization:Bearer eyJhbGciOiJFUzI1NiIsImp3ayI6eyJhbGciOiJFQ0RILUVTK0EyNTZLVyIsImNydiI6IlAtMjU2Iiwia3R5IjoiRUMiLCJ4IjoiU2JTdlZNWGJDbjZnZWFMN1F1c3JtQk1nU29uRC1vclVkNUNUdGtLS1BmQSIsInkiOiJ1M2Fzd0tVbFVieGZjUXFLSUZJZGppdmZSbklEa3c2N0k4dUNGUjY4NmMwIn0sInR5cCI6IkpXVCJ9.eyJhdWQiOlsiL3RlbmFudF9zcGFjZS9BQkNEMTIzL3Rlc3QzIl0sImV4cCI6NTMwMjY0ODgxMywiaWF0IjoxNzAyNjUyNDEzLCJpc3MiOiJnaXRodWIuY29tL2xlc3RycmF0LWdvL2p3eCIsInN1YiI6IkFCQ0QxMjMifQ.VhLKC4ObpQuUq95aHIrhEAZXQS1lbvDEjAoExb3K4dfvjEOvWbPNc9ZE9119TpTMGmFtnJkFjIggh9hzbalTpA' https://.../ABCD123/device/registration/register

```
<i> Returns </i>

Content-Type: application/jose
Body: JWE encrypted by RSA_OAEP_256 and AES256_GCM <br>
Example:

```
eyJjaXBoZXJ0ZXh0IjoiTS15bE9BcURqVFl5cjcxMElaaXRyQUtQRWVZeldsTlVXZTJpV25uU3hnck1ZMm5CMUJTbi16N1otdldLX1ctV1Jyanktb0hXZml2Tnd3TWVrTmRFN0VfV1dvR3FCYWk4WGlhVlZQdHp6cDRjTHJ5aU5GQjVWUE1VOVB4VlZQdUhlbGtfc2JKSVM4UUxreGk3VlEiLCJlbmNyeXB0ZWRfa2V5IjoiS2FQV0ZfeDlLV0c3aU51WlkzUUlaV2xlNUFkUFgyUFhSaTFDa3F1VkhLMWJZQ0c5VHZ3ZnBnIiwiaGVhZGVyIjp7ImFsZyI6IkVDREgtRVMrQTI1NktXIiwiZXBrIjp7ImNydiI6IlAtMjU2Iiwia3R5IjoiRUMiLCJ4IjoiX0JqX0gweURldGJ6bFo4NVQ5dXdOLW5UQWhQSWp6cjdybi1vRDB0Z1cyQSIsInkiOiJ5bjJ0T2JPUElSaGdETDQwMWdhTjFRTHFCVkliTzdSdUhTclFzbkR1NGJJIn19LCJpdiI6ImlMN3MyaWxqLVNWc2xROXoiLCJwcm90ZWN0ZWQiOiJleUpoYkdjaU9pSkZRMFJJTFVWVEswRXlOVFpMVnlJc0ltVnVZeUk2SWtFeU5UWkhRMDBpTENKbGNHc2lPbnNpWTNKMklqb2lVQzB5TlRZaUxDSnJkSGtpT2lKRlF5SXNJbmdpT2lKZlFtcGZTREI1UkdWMFlucHNXamcxVkRsMWQwNHRibFJCYUZCSmFucHlOM0p1TFc5RU1IUm5WekpCSWl3aWVTSTZJbmx1TW5SUFlrOVFTVkpvWjBSTU5EQXhaMkZPTVZGTWNVSldTV0pQTjFKMVNGTnlVWE51UkhVMFlra2lmWDAiLCJ0YWciOiJPV3VNbFp1SnBUbDk2VkR0RkM2RVBBIn0=
```

<b> Create Session</b>




<b>Add Credential</b>

This route receives a JWE message which contains the credential encrypted by the Users Content Certificate.

<i> Expects</i>

Content-Type: application/jose <br>
Body: JWE encrypted by RSA_OAEP_256 and AES256_GCM <br>
Method: PUT


<i> Returns </i>

Content-Type: application/jose
Body: JWE encrypted by RSA_OAEP_256 and AES256_GCM <br>

<b>Delete Credential</b>
<i> Expects </i>

Body: No Body
Method: DELETE


<i>Returns</i> 

Body: No Body


<b>Get Credentials</b>

<i> Expects </i>

Body: No Body
Method: GET
