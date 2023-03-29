# Flygon

Flygon is an experimental device controller. It includes a basic proof of concept Mode and Device Manager.

## Support and discussion

There is a [Discord server](https://discord.gg/Vjze47qchG) for support and discussion.
At this time this is likely to be mostly development discussion.

# Requirements

- [go 1.20](https://go.dev/doc/install)
- [Golbat](https://github.com/UnownHash/Golbat) (optional)
- [Flygon-Admin](https://github.com/UnownHash/Flygon-Admin) (optional)
- Database

# Instructions
1. `cp config.toml.example config.toml`
2. modify it as you want
3. `go run .`

## Run in pm2 (Recommended)
1. `go build`
2. `pm2 start ./flygon --name flygon -o "/dev/null"`

# Run in docker (Full Stack)
1. `wget -O docker-compose.yml https://raw.githubusercontent.com/UnownHash/Flygon/main/docker-compose.yml.exampl`
2. `wget -O flygon-config.toml https://raw.githubusercontent.com/UnownHash/Flygon/main/config.toml.example`
3. `wget -O golbat-config.toml https://raw.githubusercontent.com/UnownHash/Golbat/main/config.toml.example`
4. modify it as you want - adapt also admin service in `docker-compose.yml`