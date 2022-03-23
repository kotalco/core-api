## Run using docker-compose

> Make sure that Docker and [docker-compose](https://docs.docker.com/compose/install/) are installed on your machine

Project is shipped with docker-compose.yaml script which can be used to dpeloy PostgreSQL database instance alongisde the cloud-api.

```
git clone git@github.com:kotalco/cloud-api.git
cd cloud-api
docker-compose up
```

After running the above command, cloud-api server will be running at port 8080.