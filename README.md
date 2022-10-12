# :fire: CLOUD API

CLOUD API is upgraded version of [COMMUNITY API](https://github.com/kotalco/community-api) where user can leverage modules like:
- WORKSPACE
- RBAC
- ENDPOINTS

## :hammer_and_wrench: Prerequisites
Running the CLOUD API server against real k8s cluster requires:

- [kotal operator](https://github.com/kotalco/kotal) to be deployed in the cluster
- api server to be deployed with correct role and role bindings
- valid activation key

## :closed_lock_with_key:	 Environment Variables
This is a list of the environment variavbles you need to use the software.

### Mendatory Envrionment Variables
- `SEND_GRID_API_KEY`
- `ECC_PUBLIC_KEY` hexEncoded key of the elibtic curve public key used to verify the signed ciphers responses from the subscription platfrom, 
this will only works as envrionment variables in development evnvironment.<br />
in production, staging environment it should passed as a build time variables
- `SUBSCRIPTION_API_BASE_URL` subscription platfrom base_url used to activate the user activation key
- `DB_SERVER_URL`  postgres://postgres:secret@localhost:5432/db-name-goes-here

### Optional Envrionment Variables
- `CLOUD_API_SERVER_PORT` any port except the one the operator listening to
- `ENVIRONMENT` could be development or production
- `SERVER_READ_TIMEOUT`
- `ACCESS_SECRET` jwt symmetric key used to sign the Json Web Token
- `JWT_SECRET_KEY_EXPIRE_HOURS_COUNT` jwt token expiry period in hours
- `JWT_SECRET_KEY_EXPIRE_HOURS_COUNT_REMEMBER_ME` jwt token expiry when the user choose remomber me option with signing in
- `DB_TESTING_SERVER_URL`
- `DB_MAX_CONNECTIONS`
- `DB_MAX_IDLE_CONNECTIONS`
- `DB_MAX_LIFETIME_CONNECTIONS`
- `FRONT_BASE_URL` the dashboard used to redirect user to when sending emails 
- `VERIFICATION_TOKEN_LENGTH` the length of the verification tokens used by the system idl > 50 chars
- `VERIFICATION_TOKEN_EXPIRY_HOURS` 
- `SEND_GRID_SENDER_NAME` the username of the emails sent to the users
- `SEND_GRID_SENDER_EMAIL` the email address used to send the emails with
- `2_FACTOR_SECRET` symmetric key used to sign the user verification key
- `RATE_LIMITER_PER_MINUTE` 



## :closed_lock_with_key:	 Build-Time Variables
This is a list of the build-time variavbles you need pass when building the applicaion
- `ECC_PUBLIC_KEY` 
