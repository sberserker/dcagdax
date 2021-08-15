
docker rm -f dcagdax || true
docker build -t dcagdax .
docker run -d --name dcagdax -e TZ=America/Los_Angeles  --env-file .env --restart unless-stopped dcagdax
docker logs dcagdax --follow