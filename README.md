# BLK
ETH wallets balance delta fetcher  

## Build 
1. Install docker 
2. Create account at [getblock.io](https://www.getblock.io/) and get an access token. 
3. In a root of the project, create *.env* file and fill it with the following:
```
BLK_GETBLOCK_ACCESS_TOKEN=my0access0toke0here ## Access token
BLK_LOG_LEVEL=info                            ## Log level [debug / info]
BLK_HTTP_ADDR=0.0.0.0:8085                    ## Listen address
```

4. Build it
```bash
make up
```

## API
### GET /most-changed?blocks=$1
Request parameters: 
* blocks - type: uint (optional). Limits amount of blocks chat will be checked from head. Default: 100

Example:
```bash
curl --request GET \
  --url 'http://localhost:8085/most-changed'
```

Response:
```json
{
  "address": "0x3f0c3faeeeb9dad6ef6eb5fbab61039ff9067a07"
}
```