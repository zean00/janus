docker build -f dev/Dockerfile -t klikuid/janus .
docker rmi $(docker images -f "dangling=true" -q)