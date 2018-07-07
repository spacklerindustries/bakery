# Starting Bakery
## Env Vars
```
export BAKERY_ROOT=/bakery/
export DB_PATH=/bakery/piInventory.db
export KPARTX_PATH=kpartx
export NFS_ADDRESS=$(hostname -I | cut -d " " -f 1)
export BUSHWOOD_SERVER="http://$(hostname -I | cut -d " " -f 1):8090"
export BUSHWOOD_TOKEN="{token}"
export HTTP_PORT=8081
export TEMPLATE_PATH="/bakery/fileTemplates"
```

# Bootup
## Get cmdline.txt
```
curl {server}:{port}/api/v1/files/{piId}/cmdline.txt
```

# Fridge
## List Fridge
```
curl {server}:{port}/api/v1/fridge
```
## Bake PI from pool
```
curl -H "Content-Type: application/json" -X POST -d '{"bakeformName" : "{bakeform}"}' {server}:{port}/api/v1/fridge
```
## Bake PI specifying piId
```
curl -H "Content-Type: application/json" -X POST -d '{"bakeformName" : "{bakeform}", "bakeformPiId" : "{piId}"}' {server}:{port}/api/v1/fridge
```

# Oven
## List Oven
```
curl {server}:{port}/api/v1/oven
```
## List Pi in Oven
```
curl -X DELETE {server}:{port}/api/v1/oven/{piId}
```
## Delete from Oven
```
curl -X DELETE {server}:{port}/api/v1/oven/{piId}
```

# Bakeforms
## List Bakeforms
```
curl {server}:{port}/api/v1/bakeforms
```
## Upload Bakeform
```
curl -H "Content-Type: application/x-raw-disk-image" --data-binary @/home/user1/Downloads/2017-04-10-raspbian-jessie-lite.img -X POST {server}:{port}/api/v1/bakeforms/{bakeformname}
```
## Delete Bakeform
```
curl -X DELETE {server}:{port}/api/v1/bakeforms
```
