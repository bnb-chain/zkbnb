# How UT works in this package

We need mock data to run the testcase under this package, the mock data is a 
snapshot of postgres after running the integration test. 

We just create a snapshot of `postgres` container and push it on github as a docker 
image and reuse it.

## How to create mock data docker images

After you have run the integration test, while the `postgres` container is not deleted:

`docker commit -m 'add zkbnb mock data' postgres   zkbnb-ut-postgres`
`docker tag zkbnb-ut-postgres ghcr.io/bnb-chain/zkbnb/zkbnb-ut-postgres:latest`
`docker push ghcr.io/bnb-chain/zkbnb/zkbnb-ut-postgres:latest`

Note: you need login the docker registry before pushing.
```shell
export CR_PAT={your github token}
echo $CR_PAT | docker login ghcr.io -u {your user name}  --password-stdin
```

