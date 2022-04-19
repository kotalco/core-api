#! /bin/bash
read -p "Type seed to seed or Type trunc to truncate: " Answer
case $Answer in
 seed | SEED)
 go run pkg/cmd/seeder/main.go -a seed
  ;;
trunc)
  go run pkg/cmd/seeder/main.go -a trunc
esac