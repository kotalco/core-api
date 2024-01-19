# :fire: CORE API SERVER
The Core API provides a powerful interface for managing blockchain resources on Kubernetes, integrating multiple functionalities into a single, cohesive experience.
It is the essential tool for deploying and interacting with blockchain nodes and related services.

## :open_file_folder: Features
- MANAGEMENT OF BLOCKCHAIN RESOURCES LIKE IPFS PEERS, POLKADOT VALIDATOR NODES, CHAINLINK NODES, ETHEREUM NODES, AND MORE.
- WORKSPACE
- RBAC
- ENDPOINTS

## :hammer_and_wrench: Prerequisites
Running the CORE API SERVER against real k8s cluster requires:
- [kotal operator](https://github.com/kotalco/kotal) to be deployed in the cluster

## :rocket: Running the Core API Server
###  :floppy_disk: From Source Code
Clone the repository and run the API server with the following command:
```
go run main.go
```
NOTE: An actual k8s cluster with kubeconfig should be available in the default directory.
For testing with a mock server:
```
MOCK=true go run main.go
```
### :framed_picture: From Docker Image
Run the API server from a Docker image with the following command:
```
docker run -p 3000:3000 -e MOCK=true kotalco/core-api:develop
```

## :closed_lock_with_key:	 Environment Variables
This is a list of the environment variables you need to use the software.

### Mendatory Envrionment Variables
- `SEND_GRID_API_KEY` This key is used for verifying user account  (The app will panic if not provided)
- `DB_SERVER_URL`  postgres://postgres:secret@localhost:5432/db-name-goes-here  (The app will panic if not provided)

### Optional Envrionment Variables
- `CORE_API_SERVER_PORT`
- `ENVIRONMENT` could be development or production
- `SERVER_READ_TIMEOUT`
- `ACCESS_SECRET` jwt symmetric key used to sign the Json Web Token
- `JWT_SECRET_KEY_EXPIRE_HOURS_COUNT` jwt token expiry period in hours
- `JWT_SECRET_KEY_EXPIRE_HOURS_COUNT_REMEMBER_ME` jwt token expiry when the user choose remomber me option with signing in
- `DB_TESTING_SERVER_URL`
- `DB_MAX_CONNECTIONS`
- `DB_MAX_IDLE_CONNECTIONS`
- `DB_MAX_LIFETIME_CONNECTIONS`
- `VERIFICATION_TOKEN_LENGTH` the length of the verification tokens used by the system idl > 50 chars
- `VERIFICATION_TOKEN_EXPIRY_HOURS`
- `SEND_GRID_SENDER_NAME` the username of the emails sent to the users
- `SEND_GRID_SENDER_EMAIL` the email address used to send the emails with
- `2_FACTOR_SECRET` symmetric key used to sign the user verification key
- `RATE_LIMITER_PER_MINUTE` 


## :telephone_receiver: Sample cURL Calls
Create a new node:
```
curl -X POST -d '{"name": "my-node", "network": "mainnet", "client": "besu"}' -H 'content-type: application/json' localhost:3000/api/v1/ethereum/nodes
```

Get node by name:
```
curl localhost:3000/api/v1/ethereum/nodes/my-node
```

List all nodes:
```
curl localhost:3000/api/v1/ethereum/nodes
```

Update node by name:
```
curl -X PUT -d '{"rpc": true}' -H 'content-type: application/json' localhost:3000/api/v1/ethereum/nodes/my-node
```

Delete node by name:
```
curl -X DELETE localhost:3000/api/v1/ethereum/nodes/my-node
```


## :building_construction: BLOCKCHAIN RESOURCES 
- [APTOS](https://aptos.dev/nodes/aptos-api-spec/) 
- [BITCOIN](https://bitcoincore.org/en/doc/) 
- [IPFS-CLUSTER](https://ipfscluster.io/documentation/)
- [IPFS-PEER](https://docs.ipfs.tech/) 
- [ETHEREUM](https://ethereum.org/en/developers/docs) 
- [BEACON API](https://ethereum.github.io/beacon-APIs/) 
- [POLKADOT](https://polkadot.js.org/docs) 
- [CHAINLINK](https://github.com/smartcontractkit/chainlink#build-chainlink) 
- [NEAR](https://docs.near.org) 
- [FILECOIN](https://docs.filecoin.io) 
- [STACKS](https://docs.stacks.co/) 


The Core API is designed as a comprehensive solution for your blockchain infrastructure needs, providing streamlined management capabilities within Kubernetes environments.
