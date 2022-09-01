## Local Setup

#### Docker-Compose

Start...
```bash
blockNr=$(bash ./deployment/tool/tool.sh blockHeight)
bash ./deployment/tool/tool.sh all new
bash ./deployment/docker-compose/docker-compose.sh up $blockNr
```

Stop...
```bash
bash ./deployment/docker-compose/docker-compose.sh down
```

