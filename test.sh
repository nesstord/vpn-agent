if [[ "$1" == "build" ]]; then
  GOOS=linux GOARCH=$2 go build -o vpn-agent main.go
  echo "Successfully built"
elif [[ "$1" == "test" ]]; then
  GOOS=linux go build -o vpn-agent main.go
  docker-compose up --force-recreate
elif [[ "$1" == "run" ]]; then
  if ls | grep -q vpn-agent; then
    docker-compose up
  else
    echo "Vpn-agent binary file not found. Please run 'build' command"
  fi
else
  echo "Unknown command: $1"
fi