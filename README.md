
# 🖥️ Yadro test task 🖥️



## FAQ

#### Possible problem 1

permission denied while trying to connect to the Docker daemon socket at unix:///var/run/docker.sock: Get "http://%2Fvar%2Frun%2Fdocker.sock/v1.24/containers/json": dial unix /var/run/docker.sock: connect: permission denied

- answer: before any docker-command type sudo.
For example: "sudo docker build ...; sudo docker run -it"

#### Possible problem 2

Forget to stop docker-container

- answer: close the container after all work

```bash
    docker stop <container_id>
```
## Run Locally

Clone the project

```bash
  git clone https://github.com/Sem4kok/pc-club
```

Go to the project directory

```bash
  cd pc-club
```

You can run the application from here 

```bash
  go run main.go <path/to/test/file>
```

## DOCKER

docker imaging 

```bash
  docker build -t pc-club .
```

docker container launch
```bash
  docker run -it pc-club 
```
inside the container, use to run my tests
```bash
  ./task tests/<choose any my test>
```
## If you want to run your tests, you need to: 
- Do not close the container.
- Open a second terminal on your host machine
- Type

```bash
    docker ps
```
you need to copy the id of the running container (in this case 7e50d0f30bc0)
```
CONTAINER ID   IMAGE        COMMAND   CREATED         STATUS         PORTS     NAMES
7e50d0f30bc0   yadro-club   "bash"    3 minutes ago   Up 3 minutes             reverent_hermann
```

copy the file you need into a docker container
```bash
    docker cp <path/to/test_file> 7e50d0f30bc0:/app/
```

go to the terminal where the container is open.
run the desired file using 
```bash
    ./task <copied_file>
```
## Contact me

- Telegram: @peso69
- e-mail:   w1usis@mail.ru
