# SORMAS-EXPORTER
A small prometheus metrics exporter for the sormas application

## Build
```bash
docker build . -t <user>/<repo>:<tag>
```
## Configuration
| VAR | EXAMPLE |
| --- | --- |
| HOST | ```postgres``` (sormas-docker project) |
| PORT | ```5432``` |
| USER | ```sormas_user``` |
| PASSWORD | ```<your password>``` |
| RETRY | ```5``` in seconds |